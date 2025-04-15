package metrics

import (
	"os"
	"runtime"
	"syscall"
	"time"

	"mithrildb/config"
)

func getDiskStats(path string) (DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskInfo{}, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return DiskInfo{
		Path:       path,
		TotalBytes: total,
		UsedBytes:  used,
		FreeBytes:  free,
	}, nil
}

func GetServerMetrics(cfg *config.AppConfig, startTime time.Time) (ServerMetrics, error) {
	hostname, _ := os.Hostname()
	mountDisk, err := getDiskStats(".")
	if err != nil {
		return ServerMetrics{}, err
	}
	dbDisk, err := getDiskStats(cfg.RocksDB.DBPath)
	if err != nil {
		return ServerMetrics{}, err
	}

	return ServerMetrics{
		UptimeSeconds: int64(time.Since(startTime).Seconds()),
		Hostname:      hostname,
		PID:           os.Getpid(),
		Port:          cfg.Server.Port,
		GoVersion:     runtime.Version(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		MountDisk:     mountDisk,
		DBDisk:        dbDisk,
	}, nil
}
