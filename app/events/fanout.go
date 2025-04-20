package events

import (
	"encoding/json"
	"log"
	"time"
)

// StartFanout begins processing events from the queue and dispatching to all matching listeners.
func StartFanout(queue QueueReader) {
	go func() {
		for {
			key, val, err := queue.Next()
			if err != nil {
				log.Printf("[fanout] ⚠️ error reading event: %v", err)
				time.Sleep(time.Second)
				continue
			}
			if val == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			var event DocumentChangeEvent
			if err := json.Unmarshal(val, &event); err != nil {
				log.Printf("[fanout] ⚠️ invalid event skipped: %v", err)
				_ = queue.Ack(key)
				continue
			}

			matched := false
			for _, listener := range GetListeners() {
				if !listener.PreFilter(event) {
					continue
				}

				matched = true
				payload, err := json.Marshal(event)
				if err != nil {
					log.Printf("[fanout] ❌ serialize error for listener %s: %v", listener.Name, err)
					continue
				}

				if err := listener.Queue.Enqueue(payload); err != nil {
					log.Printf("[fanout] ❌ enqueue error for listener %s: %v", listener.Name, err)
					continue
				}
			}

			if !matched {
				log.Printf("[fanout] ℹ️ no listeners accepted event for key %s", string(event.Key))
			}

			if err := queue.Ack(key); err != nil {
				log.Printf("[fanout] ❌ failed to ack event: %v", err)
			}
		}
	}()
}
