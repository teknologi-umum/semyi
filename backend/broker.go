package main

import (
	"sync"

	"github.com/google/uuid"
)

// Code acquired from https://github.com/go-micro/go-micro/blob/v1.18.0/broker/memory/memory.go Apache-2.0 license.
// The code was modified to fit the project's codebase.
//
//    Copyright 2015 Asim Aslam.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

// BrokerCallbackHandler is used to process messages via a subscription of a topic.
// The handler is passed a publication interface which contains the
// message and optional Ack method to acknowledge receipt of the message.
type BrokerCallbackHandler[T any] func(BrokerEvent[T]) error

type BrokerMessage[T any] struct {
	Header map[string]string
	Body   T
}

// BrokerEvent is given to a subscription handler for processing
type BrokerEvent[T any] interface {
	Topic() string
	Message() *BrokerMessage[T]
	Ack() error
}

type Broker[T any] struct {
	sync.RWMutex
	Subscribers map[string][]*BrokerSubscriber[T]
}

type memoryEvent[T any] struct {
	topic   string
	message *BrokerMessage[T]
}

type BrokerSubscriber[T any] struct {
	id      string
	topic   string
	exit    chan bool
	handler BrokerCallbackHandler[T]
}

func (m *Broker[T]) Publish(topic string, message *BrokerMessage[T]) error {
	m.RLock()

	subs, ok := m.Subscribers[topic]
	m.RUnlock()
	if !ok {
		return nil
	}

	p := &memoryEvent[T]{
		topic:   topic,
		message: message,
	}

	for _, sub := range subs {
		if err := sub.handler(p); err != nil {
			return err
		}
	}

	return nil
}

func (m *Broker[T]) Subscribe(topic string, callback BrokerCallbackHandler[T]) (*BrokerSubscriber[T], error) {
	sub := &BrokerSubscriber[T]{
		id:      uuid.New().String(),
		topic:   topic,
		exit:    make(chan bool, 1),
		handler: callback,
	}

	m.Lock()
	m.Subscribers[topic] = append(m.Subscribers[topic], sub)
	m.Unlock()

	go func() {
		<-sub.exit
		m.Lock()
		var newSubscribers []*BrokerSubscriber[T]
		for _, sb := range m.Subscribers[topic] {
			if sb.id == sub.id {
				continue
			}
			newSubscribers = append(newSubscribers, sb)
		}
		m.Subscribers[topic] = newSubscribers
		m.Unlock()
	}()

	return sub, nil
}

func (m *memoryEvent[T]) Topic() string {
	return m.topic
}

func (m *memoryEvent[T]) Message() *BrokerMessage[T] {
	return m.message
}

func (m *memoryEvent[T]) Ack() error {
	return nil
}

func (m *BrokerSubscriber[T]) Topic() string {
	return m.topic
}

func (m *BrokerSubscriber[T]) Unsubscribe() error {
	m.exit <- true
	return nil
}

func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		Subscribers: make(map[string][]*BrokerSubscriber[T]),
	}
}
