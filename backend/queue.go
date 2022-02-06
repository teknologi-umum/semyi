package main

import (
	"errors"
	"sync"
)

var ErrEmptyQueue = errors.New("Queue is empty")

type Queue struct {
	sync.RWMutex
	Items []Response
}

func NewQueue() *Queue {
	return &Queue{
		Items: []Response{},
	}
}

func (q *Queue) Dequeue() (Response, error) {
	q.Lock()
	defer q.Unlock()

	if len(q.Items) == 0 {
		return Response{}, ErrEmptyQueue
	}

	item := q.Items[0]
	q.Items = q.Items[1:]

	return item, nil
}

func (q *Queue) LatestItem() (Response, error) {
	q.RLock()
	defer q.RUnlock()

	if len(q.Items) == 0 {
		return Response{}, ErrEmptyQueue
	}

	return q.Items[len(q.Items)-1], nil
}
