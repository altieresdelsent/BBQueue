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

// All constant values used throughout the code
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

// Init Initializer for the Redis queue
func (r *RedisInFlightStorageAndQueue) Init() error {
	r.Context = context.Background()
	r.Engine = redis.NewClient(&redis.Options{
		Addr:     RedisAddress,
		Password: "",
		DB:       0,
	})
	return r.Engine.Ping(r.Context).Err()
}

// Add Adds a new message to the queue
func (r *RedisInFlightStorageAndQueue) Add(message string, timeout time.Duration) uuid.UUID {
	key := uuid.New()
	r.addMessageInProcessing(message, timeout, key)
	return key
}

// addMessageInProcessing Private method for adding message
func (r *RedisInFlightStorageAndQueue) addMessageInProcessing(message string, timeout time.Duration, key uuid.UUID) {
	messageKey := fmt.Sprint(key.String(), TypeMessage)
	durationKey := fmt.Sprint(key.String(), TypeDuration)
	r.Engine.MSet(r.Context, messageKey, message, durationKey, time.Now().Add(timeout).Format(time.RFC3339Nano))
	r.Engine.Expire(r.Context, messageKey, DefaultMessageTimeout)
	r.Engine.Expire(r.Context, durationKey, DefaultMessageTimeout)
}

// Get Fetches a message by key
func (r *RedisInFlightStorageAndQueue) Get(key uuid.UUID) (string, time.Time, bool, error) {
	return r.getMessageInProcessing(key)
}

// Private method for fetching a message
func (r *RedisInFlightStorageAndQueue) getMessageInProcessing(key uuid.UUID) (string, time.Time, bool, error) {
	message, err := r.Engine.Get(r.Context, fmt.Sprint(key.String(), TypeMessage)).Result()
	timeoutStr, err2 := r.Engine.Get(r.Context, fmt.Sprint(key.String(), TypeDuration)).Result()
	if err == redis.Nil && err2 == redis.Nil {
		return "", time.Time{}, true, nil
	} else if err == redis.Nil || err2 == redis.Nil {
		return "", time.Time{}, false, errors.New("inconsistent processing messages")
	}

	timeout, err3 := time.Parse(time.RFC3339Nano, timeoutStr)
	if err3 != nil {
		return "", time.Time{}, false, err3
	}

	return message, timeout, false, nil
}

func (r *RedisInFlightStorageAndQueue) GetAndDeleteExpiredKeys() chan QueueMessage {
	results := make(chan QueueMessage)
	go r.scanKeysAndProcess(results)
	return results
}

func (r *RedisInFlightStorageAndQueue) scanKeysAndProcess(results chan<- QueueMessage) {
	allKeys, _ := r.Engine.Keys(r.Context, fmt.Sprint("*", TypeDuration)).Result()
	defer close(results)

	for _, key := range allKeys {
		r.processKey(key, results)
	}
}

func (r *RedisInFlightStorageAndQueue) processKey(key string, results chan<- QueueMessage) {
	keyUUID, err := uuid.Parse(strings.TrimSuffix(key, TypeDuration))
	if err != nil {
		fmt.Printf("Error parsing UUID: %v \n", err)
		return
	}

	timeoutAsString, err := r.Engine.Get(r.Context, key).Result()
	if err != nil {
		r.handleDeleteError(r.deleteByKey(keyUUID), keyUUID)
		return
	}

	messageTimeoutTime, err := time.Parse(time.RFC3339Nano, timeoutAsString)
	if err != nil {
		r.handleDeleteError(r.deleteByKey(keyUUID), keyUUID)
		return
	}

	if messageTimeoutTime.After(time.Now()) {
		return
	}

	message, err := r.Engine.Get(r.Context, fmt.Sprint(keyUUID.String(), TypeMessage)).Result()
	if err != nil {
		r.handleDeleteError(r.deleteByKey(keyUUID), keyUUID)
		return
	}

	r.handleDeleteError(r.deleteByKey(keyUUID), keyUUID)

	results <- QueueMessage{
		Key:     keyUUID,
		Message: message,
	}
}

func (r *RedisInFlightStorageAndQueue) handleDeleteError(err error, keyUUID uuid.UUID) {
	if err != nil {
		fmt.Printf("Failed to delete key: %v with error %v \n", keyUUID.String(), err)
	}
}

// Deletes a message from the queue
func (r *RedisInFlightStorageAndQueue) deleteByKey(key uuid.UUID) error {
	return r.Engine.Del(r.Context, fmt.Sprint(key.String(), TypeMessage), fmt.Sprint(key.String(), TypeDuration)).Err()
}

// Delete a message from the queue
func (r *RedisInFlightStorageAndQueue) Delete(key uuid.UUID) error {
	return r.deleteByKey(key)
}

// Push a new message onto the queue
func (r *RedisInFlightStorageAndQueue) Push(message string) error {
	return r.Engine.LPush(r.Context, QueueName, message).Err()
}

// Pop a message off of the queue
func (r *RedisInFlightStorageAndQueue) Pop() (string, bool) {
	message, err := r.Engine.RPop(r.Context, QueueName).Result()
	if err != nil {
		return "", true
	}
	return message, false
}
