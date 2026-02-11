package repository

import (
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	PoolSize     int
	MinIdleConns int
}

func NewRedisClient(cfg RedisConfig) (*goredis.Client, error) {
	options := &goredis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	}

	client := goredis.NewClient(options)

	return client, nil
}
