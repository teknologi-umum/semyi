package main

import (
	"context"
	"errors"
	"fmt"
)

type Subscriber struct {
	subscribers []*BrokerSubscriber[MonitorHistorical]
	ch          chan MonitorHistorical
}

func NewSubscriber(centralBroker *Broker[MonitorHistorical], monitorIds ...string) (*Subscriber, error) {
	if len(monitorIds) == 0 {
		return &Subscriber{}, errors.New("no monitorIds provided")
	}

	ch := make(chan MonitorHistorical)
	var subscribers []*BrokerSubscriber[MonitorHistorical]
	// create a new BrokerSubscriber
	for _, monitorId := range monitorIds {
		subscriber, err := centralBroker.Subscribe(monitorId, func(event BrokerEvent[MonitorHistorical]) error {
			// send the event to the channel
			message := event.Message()
			ch <- message.Body
			return nil
		})
		if err != nil {
			return &Subscriber{}, fmt.Errorf("failed to subscribe to monitor %s: %w", monitorId, err)
		}

		subscribers = append(subscribers, subscriber)

	}
	return &Subscriber{
		subscribers: subscribers,
		ch:          ch,
	}, nil
}

func (s *Subscriber) Listen(ctx context.Context) <-chan MonitorHistorical {
	return s.ch
}
