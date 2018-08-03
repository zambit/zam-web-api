package redismq

import (
	"context"
	"encoding/json"
	"fmt"
	"git.zam.io/wallet-backend/web-api/pkg/services/broker"
	"github.com/adjust/rmq"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	queueNamePattern   = "wa:rmq:queue:%s:%s"
	rmqPanicErrPrepend = "rmq redis error is not nil"
)

var (
	errAckFailed     = errors.New("redismq: ack failed with no reason")
	errNackFailed    = errors.New("redismq: nack failed with no reason")
	errRejectFailed  = errors.New("redismq: reject failed with no reason")
	errPubFailed     = errors.New("redismq: publish failed with no reason")
	errConsumeFailed = errors.New("redismq: consume failed with no reason")
)

// Broker
type Broker struct {
	Connection rmq.Connection
	logger     logrus.FieldLogger

	guard  sync.RWMutex
	queues map[string]rmq.Queue
}

func New(c *redis.Client, logger logrus.FieldLogger) broker.IBroker {
	rmqConn := rmq.OpenConnectionWithRedisClient("tag", c)
	return &Broker{
		logger:     logger.WithField("module", "broker.redismq"),
		Connection: rmqConn,
		queues:     make(map[string]rmq.Queue),
	}
}

// delivery
type delivery struct {
	logger logrus.FieldLogger
	orig    rmq.Delivery
	ident   broker.Identifier
	data    []byte
	headers map[string]string
}

func (d *delivery) Identifier() broker.Identifier {
	return d.ident
}

func (d *delivery) Payload() []byte {
	return d.data
}

func (d *delivery) Ack() error {
	return wrapRmqPanicAsErr(func() error {
		if !d.orig.Ack() {
			d.logger.Error("ack failed")
			return errAckFailed
		}
		d.logger.Info("acked")
		return nil
	})
}

func (d *delivery) Nack() error {
	return wrapRmqPanicAsErr(func() error {
		if !d.orig.Push() {
			d.logger.Error("nack failed")
			return errNackFailed
		}
		d.logger.Info("nacked")
		return nil
	})
}

func (d *delivery) Reject() error {
	return wrapRmqPanicAsErr(func() error {
		if !d.orig.Reject() {
			d.logger.Error("reject failed")
			return errRejectFailed
		}
		d.logger.Info("reject")
		return nil
	})
}

func (d *delivery) GetHeader(name string) (header string, ok bool) {
	header, ok = d.headers[name]
	return
}

// message
type message struct {
	Resource string            `json:"resource"`
	Action   string            `json:"action"`
	ID       string            `json:"id"`
	Payload  json.RawMessage   `json:"payload"`
	Headers  map[string]string `json:"headers"`
}

type outMessage struct {
	Resource string            `json:"resource"`
	Action   string            `json:"action"`
	ID       string            `json:"id"`
	Payload  interface{}       `json:"payload"`
	Headers  map[string]string `json:"headers"`
}

func (c *Broker) Consume(resource, action string, consumer broker.ConsumeFunc) error {
	return wrapRmqPanicAsErr(func() error {
		queueName := fmt.Sprintf(queueNamePattern, resource, action)

		alreadyConsuming := false
		func() {
			c.guard.RLock()
			defer c.guard.RUnlock()

			_, alreadyConsuming = c.queues[queueName]
		}()
		if alreadyConsuming {
			return errors.New("redismq: already consuming")
		}

		queue := c.Connection.OpenQueue(queueName)
		queue.SetPushQueue(queue)
		if !queue.StartConsuming(1, time.Second/2) {
			return errConsumeFailed
		}
		queue.AddConsumerFunc(queueName, func(d rmq.Delivery) {
			msg := message{}
			err := json.Unmarshal([]byte(d.Payload()), &msg)
			if err != nil {
				c.logger.WithError(err).Error("failed to unmarshal delivery, rejecting")
				d.Reject()
				return
			}

			ident := broker.Identifier{
				Resource: msg.Resource, Action: msg.Action, ID: msg.ID,
			}

			c.logger.WithField("identify", ident).WithField("data", string(msg.Payload)).Infof("message received")

			defer func() {
				p := recover()
				if p != nil {
					d.Push()
					c.logger.WithField("panic", p).Error("panic occurs while consuming message")
					panic(p)
				}
			}()

			err = consumer(
				c,
				&delivery{
					c.logger.WithField("module", "broker.redismq.delivery").WithField("ident", ident),
					d, ident, []byte(msg.Payload), msg.Headers,
				},
			)
			if err != nil {
				c.logger.WithError(err).Error("error occurs while calling handler")
			}
		})

		c.logger.WithField("path", fmt.Sprintf("%s.%s.*", resource, action)).Info("consuming started")

		c.guard.Lock()
		defer c.guard.Unlock()
		c.queues[queueName] = queue

		return nil
	})
}

func (c *Broker) StopConsumer(resource, action string) error {
	return nil
}

func (c *Broker) Publish(identifier broker.Identifier, payload interface{}) error {
	return c.PublishCtx(context.Background(), identifier, payload)
}

func (c *Broker) PublishCtx(ctx context.Context, identifier broker.Identifier, payload interface{}) error {
	c.guard.Lock()
	defer c.guard.Unlock()

	return wrapRmqPanicAsErr(func() error {
		queueName := fmt.Sprintf(queueNamePattern, identifier.Resource, identifier.Action)

		l := c.logger.WithField(
			"identify", identifier,
		).WithField(
			"data", payload,
		).WithField(
			"queue", queueName,
		)

		queue, ok := c.queues[queueName]
		if !ok {
			queue = c.Connection.OpenQueue(queueName)
			queue.SetPushQueue(queue)
			c.queues[queueName] = queue
		}

		l.Infof("publishing message...")

		msg := outMessage{
			identifier.Resource,
			identifier.Action,
			identifier.ID,
			payload,
			nil,
		}
		bytes, err := json.Marshal(&msg)
		if err != nil {
			l.WithError(err).Error("message marshalling error")
			return err
		}

		if !queue.Publish(string(bytes)) {
			l.Error("message publishing failed")
			return errPubFailed
		}
		l.Info("message successfully published")
		return nil
	})
}

func (c *Broker) Start() error {
	c.logger.Info("starting listening")
	return nil
}

func (c *Broker) Stop() error {
	c.guard.Lock()
	defer c.guard.Unlock()

	c.logger.Info("stopping consuming")

	for _, q := range c.queues {
		q.StopConsuming()
	}

	return nil
}

func wrapRmqPanicAsErr(wrappable func() error) (err error) {
	defer func() {
		p := recover()
		if e, ok := p.(error); p != nil && ok {
			if e.Error()[:len(rmqPanicErrPrepend)] == rmqPanicErrPrepend {
				err = e
				return
			}
		}
		if p != nil {
			panic(p)
		}
	}()

	return wrappable()
}
