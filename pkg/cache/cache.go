package cache

import (
	"HydraServer/pkg/redisbackend"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"reflect"
	"time"
)

type Cache struct {
	client redis.UniversalClient
	g      singleflight.Group
	log    logrus.FieldLogger
}

func NewCache(redisBackend *redisbackend.RedisBackend, log logrus.FieldLogger) *Cache {
	roleDBMgr := &Cache{
		client: redisBackend.Client(),
		g:      singleflight.Group{},
		log:    log,
	}
	return roleDBMgr
}

// GetOrSet 通用的获取数据函数，支持缓存和数据库查询
func (c *Cache) GetOrSet(ctx context.Context, key interface{}, table interface{}, query func() (interface{}, error)) (interface{}, error) {
	cacheKey := genCacheKey(key, table)
	// 使用 singleflight 防止缓存击穿
	val, err, _ := c.g.Do(cacheKey, func() (interface{}, error) {
		// 先尝试从缓存获取
		result, err := c.client.Get(ctx, cacheKey).Result()
		if errors.Is(err, redis.Nil) {
			// 缓存未命中，执行查询
			value, err := query()
			if err != nil {
				return nil, err
			}

			b, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}

			// 将结果存入缓存
			err = c.client.Set(ctx, cacheKey, b, redis.KeepTTL).Err()
			if err != nil {
				c.log.Error("Failed to set cache", err)
			}

			return value, nil
		} else if err != nil {
			return nil, fmt.Errorf("cache get failed: %v", err)
		}

		// 从缓存获取数据
		err = json.Unmarshal([]byte(result), table)
		return table, err
	})

	return val, err
}

// Del 删除缓存
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Update 更新缓存
func (c *Cache) Update(ctx context.Context, key interface{}, table interface{}) error {
	cacheKey := genCacheKey(key, table)

	b, err := json.Marshal(table)
	if err != nil {
		return err
	}

	// 将结果存入缓存
	err = c.client.Set(ctx, cacheKey, b, time.Minute*5).Err()
	if err != nil {
		c.log.Error("Failed to set cache", err)
	}
	return err
}

// genCacheKey 生成缓存的键，使用反射调用 TableName 方法
func genCacheKey(key interface{}, table interface{}) string {
	val := reflect.ValueOf(table)

	// 确保传入的是指针类型
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		// 获取 TableName 方法
		method := val.MethodByName("TableName")
		if method.IsValid() {
			// 调用方法并获取返回值
			result := method.Call([]reflect.Value{})
			if len(result) > 0 {
				return fmt.Sprintf("%v-%v", key, result[0].String())
			}
		}
	}

	// 如果没有找到方法或者不是有效的指针，返回默认的格式
	return fmt.Sprintf("%v-%v", key, table)
}
