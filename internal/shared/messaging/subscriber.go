package messaging

import (
	"context"
	"io"
)

type Topic string

type BackMessage interface {
	GetPayload() Payload
	Nack() bool
	Ack() bool
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic Topic) (<-chan BackMessage, error)
}

func Listen[T Message](
	ctx context.Context,
	s Subscriber,
	topic Topic,
	h Handler[T],
) (cancel func() error, err error) {
	messages, err := s.Subscribe(ctx, topic)
	if err != nil {
		return nil, err
	}

	subCtx, cancelCtx := context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-subCtx.Done():
				return
			case m, ok := <-messages:
				if !ok {
					return
				}

				msg, err := MessageFromPayload[T](m.GetPayload())
				if err != nil {
					m.Nack()
					continue
				}

				err = h.Handle(subCtx, *msg)
				if err != nil {
					m.Nack()
					continue
				}

				m.Ack()
			}
		}
	}()

	cancelFunc := func() error {
		cancelCtx()
		if closer, ok := any(s).(io.Closer); ok {
			return closer.Close()
		}
		return nil
	}

	return cancelFunc, nil
}
