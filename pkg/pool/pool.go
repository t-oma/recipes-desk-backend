package pool

import "context"

type Pool[T any] interface {
	Get(ctx context.Context) (T, error)
	Put(item T)
	Close(ctx context.Context) error
}

type Stats[T any] interface {
	Stats() T
}
