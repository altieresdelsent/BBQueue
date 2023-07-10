package queue

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	defaultTimeout = time.Duration(5) * time.Minute
)

type ExtraQueue struct {
	internalQueue NormalQueue
	processing    InFlightStorage
	Timeout       time.Duration
}

type QueueMessage struct {
	Key     uuid.UUID
	Message string
}

func NewExtraQueue(internalQueue NormalQueue, processing InFlightStorage) *ExtraQueue {
	return &ExtraQueue{
		internalQueue: internalQueue,
		processing:    processing,
		Timeout:       defaultTimeout,
	}
}

func (receiver *ExtraQueue) Add(message string) error {
	return receiver.internalQueue.Push(message)
}

func (receiver *ExtraQueue) Get(timeout time.Duration) (key uuid.UUID, message string, empty bool) {
	message, empty = receiver.internalQueue.Pop()
	if !empty {
		key = receiver.processing.Add(message, timeout)
	} else {
		key = uuid.Nil
	}

	return key, message, empty
}

func (receiver *ExtraQueue) GetInProcessing(key uuid.UUID) (message string, timeout time.Time, empty bool, er error) {
	return receiver.processing.Get(key)
}

func (receiver *ExtraQueue) Delete(key uuid.UUID) error {
	return receiver.processing.Delete(key)
}

func (receiver *ExtraQueue) ReAddExpiredKeys() {
	items := receiver.processing.GetAndDeleteExpiredKeys()

	for item := range items {
		err := receiver.Add(item.Message)
		if err != nil {
			fmt.Println("error re-adding expired message to queue")
		}
	}
}

type NormalQueue interface {
	Push(message string) error
	Pop() (message string, empty bool)
}

type InFlightStorage interface {
	Add(message string, timeout time.Duration) uuid.UUID
	GetAndDeleteExpiredKeys() chan QueueMessage
	Delete(key uuid.UUID) error
	Get(key uuid.UUID) (string, time.Time, bool, error)
}
