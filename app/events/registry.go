package events

import (
	"sync"
)

// Queue defines an interface for enqueuing event messages.
type Queue interface {
	Enqueue(value []byte) error
}

type Listener struct {
	Name      string
	Queue     Queue
	PreFilter func(DocumentChangeEvent) bool
}

var (
	listenersMu sync.RWMutex
	listeners   []Listener
)

// RegisterListener adds a new listener to receive fanout events.
func RegisterListener(l Listener) {
	listenersMu.Lock()
	defer listenersMu.Unlock()
	listeners = append(listeners, l)
}

// GetListeners returns a snapshot of all currently registered listeners.
func GetListeners() []Listener {
	listenersMu.RLock()
	defer listenersMu.RUnlock()
	return append([]Listener(nil), listeners...) // defensive copy
}
