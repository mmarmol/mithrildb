package services

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	metricsMu   sync.Mutex
	getCount    int
	putCount    int
	deleteCount int
)

func WithMetrics(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsMu.Lock()
		switch r.URL.Path {
		case "/get":
			getCount++
		case "/put":
			putCount++
		case "/delete":
			deleteCount++
		}
		metricsMu.Unlock()
		h(w, r)
	}
}

func HandleMetrics(w http.ResponseWriter, r *http.Request) {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	fmt.Fprintf(w, "get_requests %d\nput_requests %d\ndelete_requests %d\n", getCount, putCount, deleteCount)
}
