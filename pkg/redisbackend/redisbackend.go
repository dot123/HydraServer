package redisbackend

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var (
	rs     *redsync.Redsync
	client redis.UniversalClient
)

type RedisLocker interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) (bool, error)
}

type RedisBackend struct {
	mutex *redsync.Mutex
}

func NewRedisBackend(config *redis.UniversalOptions) *RedisBackend {
	client = redis.NewUniversalClient(config)
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(pong)

	pool := goredis.NewPool(client)
	rs = redsync.New(pool)

	return new(RedisBackend)
}

func (rb *RedisBackend) NewMutex(key string) RedisLocker {
	m := rs.NewMutex(fmt.Sprintf("lock/%s", key), redsync.WithExpiry(30*time.Second))
	return &RedisBackend{mutex: m}
}

func (rb *RedisBackend) Lock(ctx context.Context) error {
	return rb.mutex.LockContext(ctx)
}

func (rb *RedisBackend) Unlock(ctx context.Context) (bool, error) {
	return rb.mutex.UnlockContext(ctx)
}

func (rb *RedisBackend) Client() redis.UniversalClient {
	return client
}

// FlushDB 清空当前数据库
func (rb *RedisBackend) FlushDB(ctx context.Context) {
	if err := client.FlushDB(ctx).Err(); err != nil {
		log.Println("Failed to clear the current database:", err)
	} else {
		log.Println("Current database has been cleared")
	}
}

// FlushAll 清空所有数据库
func (rb *RedisBackend) FlushAll(ctx context.Context) {
	if err := client.FlushAll(ctx).Err(); err != nil {
		log.Println("Failed to clear all databases:", err)
	} else {
		log.Println("All databases have been cleared")
	}
}
