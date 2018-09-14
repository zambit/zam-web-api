package broker

import (
	"context"
	"fmt"
)

// Identifier
type Identifier struct {
	Resource string
	Action   string
	ID       string
}

// String
func (i Identifier) String() string {
	return fmt.Sprintf("%s.%s.%s", i.Resource, i.Action, i.ID)
}

// Delivery
type Delivery interface {
	Identifier() Identifier
	Payload() []byte

	Ack() error
	Nack() error
	Reject() error

	GetHeader(name string) (header string, ok bool)
}

// ConsumeFunc
type ConsumeFunc func(IBroker, Delivery) error

// MiddlewareFunc
type MiddlewareFunc func(b IBroker, d Delivery, next ConsumeFunc) error

// IBroker
type IBroker interface {
	AddMiddleware(middleware MiddlewareFunc)
	Consume(resource, action string, consumer ConsumeFunc) error
	StopConsumer(resource, action string) error

	Publish(identifier Identifier, payload interface{}) error
	PublishCtx(ctx context.Context, identifier Identifier, payload interface{}) error

	Start() error
	Stop() error
}
