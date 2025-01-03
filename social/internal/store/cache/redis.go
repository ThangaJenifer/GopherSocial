package cache

import "github.com/go-redis/redis/v8" //when using redis 6.x use v8 but using redis above 7.x use v9

func NewRedisClient(addr, pw string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})
}
