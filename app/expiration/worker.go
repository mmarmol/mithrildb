package expiration

import (
	"encoding/json"
	"log"
	"time"

	"mithrildb/db"
	"mithrildb/events"
)

type Listener struct {
	DB    *db.DB
	Queue *events.RocksQueue
}

func NewListener(db *db.DB, q *events.RocksQueue) *Listener {
	return &Listener{
		DB:    db,
		Queue: q,
	}
}

func (l *Listener) Start() {
	go func() {
		for {
			key, value, err := l.Queue.Next()
			if err != nil {
				log.Printf("[ttl] ⚠️ error reading from ttl queue: %v", err)
				time.Sleep(time.Second)
				continue
			}
			if value == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			var evt events.DocumentChangeEvent
			if err := json.Unmarshal(value, &evt); err != nil {
				log.Printf("[ttl] ❌ invalid TTL event: %v", err)
				_ = l.Queue.Ack(key)
				continue
			}

			switch {
			case evt.Operation == events.OpDelete:
				if err := l.DB.ClearAllTTL(evt.CF, evt.Key); err != nil {
					log.Printf("[ttl] ⚠️ failed to clear TTL for deleted key %s:%s: %v", evt.CF, evt.Key, err)
				}

			case evt.Document != nil && evt.Document.Meta.Expiration > 0:
				if err := l.DB.ReplaceTTL(evt.CF, evt.Key, evt.Document.Meta.Expiration); err != nil {
					log.Printf("[ttl] ⚠️ failed to update TTL for key %s:%s: %v", evt.CF, evt.Key, err)
				}
			}

			if err := l.Queue.Ack(key); err != nil {
				log.Printf("[ttl] ❌ failed to ack TTL event: %v", err)
			}
		}
	}()
}
