package metrics

import (
	"mithrildb/events"
	"mithrildb/expiration"
)

func GetExpirationMetrics(s *expiration.Service) *ExpirationMetrics {
	stats := s.Stats()

	return &ExpirationMetrics{
		LastRunUnix:      stats.LastRunTime.Unix(),
		LastDurationMs:   stats.LastDuration.Milliseconds(),
		LastDeleted:      stats.LastDeleted,
		TotalDeleted:     stats.TotalDeleted,
		TotalRuns:        stats.TotalRuns,
		LastErrorMessage: formatError(stats.LastError),
	}
}

func formatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func GetEventSystemMetrics() *EventSystemMetrics {
	var queues []QueueMetrics

	// Agregar métricas de listeners
	for _, l := range events.GetListeners() {
		if l.Queue != nil {
			head, tail, depth := l.Queue.Metrics()
			queues = append(queues, QueueMetrics{
				Name:  l.Name,
				Head:  head,
				Tail:  tail,
				Depth: depth,
			})
		}
	}

	// Agregar métricas de la cola de eventos
	if q := events.GetEventQueue(); q != nil {
		head, tail, depth := q.Metrics()
		queues = append(queues, QueueMetrics{
			Name:  "system.eventqueue",
			Head:  head,
			Tail:  tail,
			Depth: depth,
		})
	}

	return &EventSystemMetrics{Queues: queues}
}
