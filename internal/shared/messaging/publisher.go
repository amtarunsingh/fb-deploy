package messaging

import (
	"github.com/google/uuid"
	"reflect"
)

type Payload []byte

type Message interface {
	GetId() uuid.UUID
	GetPayload() Payload
	Load(payload Payload) error
}

func MessageFromPayload[T Message](payload Payload) (*T, error) {
	var t T

	rv := reflect.ValueOf(&t).Elem()
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	err := t.Load(payload)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type Publisher interface {
	Publish(topic Topic, message Message) error
}
