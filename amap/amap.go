package amap

import (
	"github.com/go-estar/logger"
	"golang.org/x/time/rate"
)

func New(key string, logger2 logger.Logger) *AMap {
	if key == "" {
		panic("amap key must set")
	}
	if logger2 == nil {
		panic("amap logger must set")
	}
	return &AMap{
		Key:     key,
		Limiter: rate.NewLimiter(1, 1),
		Logger:  logger2,
	}
}

type AMap struct {
	Key string
	*rate.Limiter
	logger.Logger
}
