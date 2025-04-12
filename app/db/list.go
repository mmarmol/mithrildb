package db

import (
	"github.com/linxGnu/grocksdb"
)

func (db *DB) ListKeys(prefix string, startAfter string, limit int) ([]string, error) {
	var keys []string

	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetFillCache(false) // avoid polluting the read cache
	defer readOpts.Destroy()

	it := db.TransactionDB.NewIterator(readOpts)
	defer it.Close()

	startKey := []byte(startAfter)
	prefixBytes := []byte(prefix)

	// Start iteration
	if len(startKey) > 0 {
		it.Seek(startKey)
		if it.Valid() && string(it.Key().Data()) == startAfter {
			it.Next() // skip the startAfter key itself
		}
	} else if len(prefixBytes) > 0 {
		it.Seek(prefixBytes)
	} else {
		it.SeekToFirst()
	}

	for ; it.Valid(); it.Next() {
		key := it.Key()
		defer key.Free()

		k := string(key.Data())

		// Optional: enforce prefix filtering even if we Seek by prefix
		if len(prefixBytes) > 0 && !hasPrefix(k, prefix) {
			break
		}

		keys = append(keys, k)
		if len(keys) >= limit {
			break
		}
	}

	return keys, nil
}

func hasPrefix(s, prefix string) bool {
	return len(prefix) == 0 || (len(s) >= len(prefix) && s[:len(prefix)] == prefix)
}
