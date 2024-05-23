package main_test

import (
	"fmt"
	"testing"

	main "semyi"
)

func TestMemoryBroker(t *testing.T) {
	b := main.NewBroker[string]()

	topic := "test"
	count := 10

	sub, err := b.Subscribe(topic, func(event main.BrokerEvent[string]) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error subscribing %v", err)
	}

	for i := 0; i < count; i++ {
		message := &main.BrokerMessage[string]{
			Header: map[string]string{
				"foo": "bar",
				"id":  fmt.Sprintf("%d", i),
			},
			Body: "Hello world",
		}

		if err := b.Publish(topic, message); err != nil {
			t.Fatalf("Unexpected error publishing %d", i)
		}
	}

	if err := sub.Unsubscribe(); err != nil {
		t.Fatalf("Unexpected error unsubscribing from %s: %v", topic, err)
	}
}
