package queue

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"strings"
	"time"
)

type RedisInFlightStorageAndQueue struct {
	Context context.Context
	Engine  *redis.Client
}

func (r RedisInFlightStorageAndQueue) Add(message string, timeout time.Duration) uuid.UUID {
	key := uuid.New()
	r.Engine.Set(r.Context, fmt.Sprint(key.String(), "_message"), message, time.Hour*48)
	r.Engine.Set(r.Context, fmt.Sprint(key.String(), "_duration"), time.Now().Add(timeout).String(), time.Hour*48)
	return key
}

func (r RedisInFlightStorageAndQueue) GetAndDeleteExpiredKeys() chan QueueMessage {
	results := make(chan QueueMessage)
	go func() {
		allKeys, _ := r.Engine.Keys(r.Context, "*_duration").Result()
		defer close(results)
		for _, key := range allKeys {
			keyUUID, err := uuid.Parse(strings.TrimSuffix(key, "_duration"))

			timeoutAsString, err := r.Engine.Get(r.Context, key).Bytes()
			if err == redis.Nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), "_message"), fmt.Sprint(keyUUID.String(), "_duration"))
				continue
			}
			if err != nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), "_message"), fmt.Sprint(keyUUID.String(), "_duration"))
				continue
			}

			messageTimeoutTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", string(timeoutAsString))

			if messageTimeoutTime.After(time.Now()) {
				continue
			}

			message, err := r.Engine.Get(r.Context, fmt.Sprint(keyUUID.String(), "_message")).Bytes()
			if err == redis.Nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), "_message"), fmt.Sprint(keyUUID.String(), "_duration"))
				continue
			}
			if err != nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), "_message"), fmt.Sprint(keyUUID.String(), "_duration"))
				continue
			}

			r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), "_message"), fmt.Sprint(keyUUID.String(), "_duration"))
			results <- QueueMessage{
				Key:     keyUUID,
				Message: string(message),
			}
		}

	}()
	return results
}

func (r RedisInFlightStorageAndQueue) Delete(key uuid.UUID) {
	r.Engine.Del(r.Context, fmt.Sprint(key.String(), "_message"), fmt.Sprint(key.String(), "_duration"))
}

func (r RedisInFlightStorageAndQueue) Push(message string) {
	r.Engine.LPush(r.Context, message)
}

func (r RedisInFlightStorageAndQueue) Pop() (message string, empty bool) {
	message, err := r.Engine.RPop(r.Context, message).Result()
	if err != nil {
		return "", true
	}
	return message, false
}
