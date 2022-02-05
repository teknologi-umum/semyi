package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/allegro/bigcache/v3"
)

type Subscriber struct {
	url   []string
	db    *sql.DB
	cache *bigcache.BigCache
	queue *Queue
}

func (d *Deps) NewSubscriber(url ...string) (*Subscriber, error) {
	if len(url) == 0 {
		return &Subscriber{}, errors.New("no url provided")
	}
	return &Subscriber{
		url:   url,
		db:    d.DB,
		cache: d.Cache,
		queue: d.Queue,
	}, nil
}

func (s *Subscriber) Listen(ctx context.Context) <-chan Response {
	ch := make(chan Response)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// listen for changes in s.queue.Items
				item, err := s.queue.LatestItem()
				if err != nil {
					if errors.Is(err, ErrEmptyQueue) {
						continue
					}
					log.Println("Error dequeueing item from queue:", err)
					continue
				}

				if contains(item.URL, s.url) {
					// send item to ch
					ch <- item
				}

			}
		}
	}()

	return ch
}

func contains(item string, items []string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}
