package messaging

import "context"

type Handler[T Message] interface {
	Handle(ctx context.Context, message T) error
}
