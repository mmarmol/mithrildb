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
	Server  ServerMetrics  `json:"server"`
	RocksDB map[string]any `json:"rocksdb"`
}
