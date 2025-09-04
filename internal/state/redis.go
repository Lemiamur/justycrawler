package state

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisTimeout = 5 * time.Second
)

// RedisState реализует интерфейс crawler.State с использованием Redis.
type RedisState struct {
	client *redis.Client
	setKey string
}

// NewRedisState создает новый экземпляр RedisState.
func NewRedisState(ctx context.Context, addr, password string, db int, setKey string) (*RedisState, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	pingCtx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}

	return &RedisState{
		client: rdb,
		setKey: setKey,
	}, nil
}

// Add атомарно добавляет URL в множество в Redis.
func (s *RedisState) Add(ctx context.Context, url string) (bool, error) {
	result, err := s.client.SAdd(ctx, s.setKey, url).Result()
	if err != nil {
		return false, fmt.Errorf("ошибка выполнения команды SADD в Redis: %w", err)
	}
	// Если result == 1, значит, элемент был новым.
	return result > 0, nil
}

// Clear удаляет ключ состояния из Redis.
func (s *RedisState) Clear(ctx context.Context) error {
	if err := s.client.Del(ctx, s.setKey).Err(); err != nil {
		return fmt.Errorf("ошибка выполнения команды DEL в Redis для ключа %s: %w", s.setKey, err)
	}
	return nil
}

// Close закрывает соединение с Redis.
func (s *RedisState) Close() error {
	return s.client.Close()
}
