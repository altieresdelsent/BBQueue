package queue

import (
	"github.com/google/uuid"
	"time"
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

func (receiver ExtraQueue) Add(message string) {
	receiver.internalQueue.Push(message)
}

func (receiver ExtraQueue) Get() (key uuid.UUID, message string) {
	message = receiver.internalQueue.Pop()
	key = receiver.processing.Add(message, receiver.Timeout)
	return key, message
}

func (receiver ExtraQueue) Delete(key uuid.UUID) {
	receiver.processing.Delete(key)
}

func (receiver ExtraQueue) ReAddExpiredKeys(key uuid.UUID) {
	items := receiver.processing.GetAndDeleteExpiredKeys()

	for item := range items {
		receiver.Add(item.Message)
	}
}

type NormalQueue interface {
	Push(message string)
	Pop() (message string)
}

type InFlightStorage interface {
	Add(message string, timeout time.Duration) uuid.UUID
	GetAndDeleteExpiredKeys() chan QueueMessage
	Delete(key uuid.UUID)
}
