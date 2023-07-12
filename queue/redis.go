package queue

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"strings"
	"time"
)

const (
	RedisAddress          = "127.0.0.1:6379"
	QueueName             = "q"
	TypeMessage           = "_message"
	TypeDuration          = "_duration"
	DefaultMessageTimeout = time.Hour * 48
)

type RedisInFlightStorageAndQueue struct {
	Context context.Context
	Engine  *redis.Client
}

func (r *RedisInFlightStorageAndQueue) Init() error {
	r.Context = context.Background()

	r.Engine = redis.NewClient(&redis.Options{
		Addr:     RedisAddress,
		Password: "",
		DB:       0,
	})

	return r.Engine.Ping(r.Context).Err()
}

func (r *RedisInFlightStorageAndQueue) Add(message string, timeout time.Duration) uuid.UUID {
	key := uuid.New()
	r.addMessageInProcessing(message, timeout, key)
	return key
}

func (r *RedisInFlightStorageAndQueue) addMessageInProcessing(message string, timeout time.Duration, key uuid.UUID) {
	r.Engine.Set(r.Context, fmt.Sprint(key.String(), TypeMessage), message, DefaultMessageTimeout)

	r.Engine.Set(r.Context, fmt.Sprint(key.String(), TypeDuration), time.Now().Add(timeout).Format(time.RFC3339Nano), DefaultMessageTimeout)
}

func (r *RedisInFlightStorageAndQueue) Get(key uuid.UUID) (message string, timeout time.Time, empty bool, er error) {
	return r.getMessageInProcessing(key)

}

func (r *RedisInFlightStorageAndQueue) getMessageInProcessing(key uuid.UUID) (message string, timeout time.Time, empty bool, err error) {
	message, err = r.Engine.Get(r.Context, fmt.Sprint(key.String(), TypeMessage)).Result()

	timeoutString, err2 := r.Engine.Get(r.Context, fmt.Sprint(key.String(), TypeDuration)).Result()

	if err == redis.Nil && err2 == redis.Nil {
		empty = true
		return
	} else if err == redis.Nil || err2 == redis.Nil {
		err = errors.New("inconsistent processing messages")
		return
	}

	if err2 != nil {
		err = err2
		return
	}

	timeout, err3 := time.Parse(time.RFC3339Nano, timeoutString)
	if err3 != nil {
		err = err3
		return
	}
	return
}

func (r *RedisInFlightStorageAndQueue) GetAndDeleteExpiredKeys() chan QueueMessage {
	results := make(chan QueueMessage)
	go func() {
		allKeys, _ := r.Engine.Keys(r.Context, fmt.Sprint("*", TypeDuration)).Result()
		defer close(results)
		for _, key := range allKeys {
			keyUUID, err := uuid.Parse(strings.TrimSuffix(key, TypeDuration))

			timeoutAsString, err := r.Engine.Get(r.Context, key).Bytes()
			if err == redis.Nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), TypeMessage), fmt.Sprint(keyUUID.String(), TypeDuration))
				continue
			}
			if err != nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), TypeMessage), fmt.Sprint(keyUUID.String(), TypeDuration))
				continue
			}

			messageTimeoutTime, err := time.Parse(time.RFC3339Nano, string(timeoutAsString))

			if messageTimeoutTime.After(time.Now()) {
				continue
			}

			message, err := r.Engine.Get(r.Context, fmt.Sprint(keyUUID.String(), TypeMessage)).Bytes()
			if err == redis.Nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), TypeMessage), fmt.Sprint(keyUUID.String(), TypeDuration))
				continue
			}
			if err != nil {
				r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), TypeMessage), fmt.Sprint(keyUUID.String(), TypeDuration))
				continue
			}

			r.Engine.Del(r.Context, fmt.Sprint(keyUUID.String(), TypeMessage), fmt.Sprint(keyUUID.String(), TypeDuration))
			results <- QueueMessage{
				Key:     keyUUID,
				Message: string(message),
			}
		}

	}()
	return results
}

func (r *RedisInFlightStorageAndQueue) Delete(key uuid.UUID) (bool, error) {
	err := r.Engine.Del(r.Context, fmt.Sprint(key.String(), TypeMessage), fmt.Sprint(key.String(), TypeDuration)).Err()
	if err == redis.Nil {
		return true, nil
	}
	return false, err
}

func (r *RedisInFlightStorageAndQueue) Push(message string) error {
	return r.Engine.LPush(r.Context, QueueName, message).Err()
}

func (r *RedisInFlightStorageAndQueue) Pop() (message string, empty bool) {
	message, err := r.Engine.RPop(r.Context, QueueName).Result()
	if err != nil {
		return "", true
	}
	return message, false
}
