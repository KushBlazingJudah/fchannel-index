package util

import (
	"container/list"
	"errors"
	"fmt"
)

type Queue struct {
	Queue *list.List
}

func (q Queue) Enqueue(value string) {
	q.Queue.PushBack(value)
}

func (q Queue) Dequeue() (string, error) {
	var value string

	if q.Queue.Len() == 0 {
		return value, errors.New("Queue is empty")
	}

	element := q.Queue.Front()
	q.Queue.Remove(element)
	value = fmt.Sprintf("%v", element.Value)

	return value, nil
}

func (q Queue) Len() int {
	return q.Queue.Len()
}
