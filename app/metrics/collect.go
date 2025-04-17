package metrics

import "mithrildb/expiration"

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
