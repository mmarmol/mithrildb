package metrics

import (
	"strconv"
	"strings"

	"github.com/linxGnu/grocksdb"
)

func GetRocksDBMetrics(db *grocksdb.TransactionDB) map[string]any {
	props := []string{
		"rocksdb.estimate-num-keys",
		"rocksdb.total-sst-files-size",
		"rocksdb.num-entries-active-mem-table",
		"rocksdb.num-live-versions",
		"rocksdb.num-running-compactions",
	}

	results := make(map[string]any)

	for _, prop := range props {
		val := db.GetProperty(prop)
		if val != "" {
			if n, err := strconv.ParseUint(val, 10, 64); err == nil {
				results[prop] = n
			} else {
				results[prop] = val
			}
		}
	}

	stats := db.GetProperty("rocksdb.stats")
	if stats != "" {
		lines := strings.Split(stats, "\n")
		results["rocksdb.stats_summary"] = strings.Join(lines[0:4], "\n")
	}

	return results
}
