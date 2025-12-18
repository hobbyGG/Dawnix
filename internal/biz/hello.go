package biz

import "context"

type HelloRepo interface {
	Hello(ctx context.Context, data interface{}) error
}
