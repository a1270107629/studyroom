package ratelimit

import "golang.org/x/net/context"

//go:generate mockgen -source=./types.go -package=limitmocks -destination=mocks/limiter.mock.go Limiter
type Limiter interface {
	// Limit 要不要限流
	// 这是一种最简单的定义方式
	Limit(ctx context.Context, key string) (bool, error)
}
