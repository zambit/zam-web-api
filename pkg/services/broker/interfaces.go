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

// IBroker
type IBroker interface {
	Consume(resource, action string, consumer ConsumeFunc) error
	StopConsumer(resource, action string) error

	Publish(identifier Identifier, payload interface{}) error
	PublishCtx(ctx context.Context, identifier Identifier, payload interface{}) error

	Start() error
	Stop() error
}
