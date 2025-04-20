package metrics

type DiskInfo struct {
	Path       string `json:"path"`
	TotalBytes uint64 `json:"total_bytes"`
	UsedBytes  uint64 `json:"used_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
}

type ServerMetrics struct {
	UptimeSeconds int64    `json:"uptime_seconds"`
	Hostname      string   `json:"hostname"`
	PID           int      `json:"pid"`
	Port          int      `json:"port"`
	GoVersion     string   `json:"go_version"`
	OS            string   `json:"os"`
	Arch          string   `json:"arch"`
	MountDisk     DiskInfo `json:"mount_disk"`
	DBDisk        DiskInfo `json:"db_disk"`
}

type FullMetrics struct {
	Server     ServerMetrics       `json:"server"`
	RocksDB    map[string]any      `json:"rocksdb"`
	Expiration *ExpirationMetrics  `json:"expiration,omitempty"`
	Events     *EventSystemMetrics `json:"events,omitempty"`
}

type ExpirationMetrics struct {
	LastRunUnix      int64  `json:"last_run_unix"`
	LastDurationMs   int64  `json:"last_duration_ms"`
	LastDeleted      int    `json:"last_deleted"`
	TotalDeleted     int    `json:"total_deleted"`
	TotalRuns        int    `json:"total_runs"`
	LastErrorMessage string `json:"last_error,omitempty"`
}

type QueueMetrics struct {
	Name  string `json:"name"`
	Head  uint64 `json:"head"`
	Tail  uint64 `json:"tail"`
	Depth uint64 `json:"depth"`
}

type EventSystemMetrics struct {
	Queues []QueueMetrics `json:"queues"`
}
