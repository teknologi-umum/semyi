package main

import "sync"

type Queue struct {
	sync.RWMutex
	Items []Response
}

func NewQueue() *Queue {
	return &Queue{
		Items: []Response{},
	}
}
