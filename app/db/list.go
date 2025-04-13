package db

import (
	"github.com/linxGnu/grocksdb"
)

func (db *DB) ListKeys(cf string, prefix string, startAfter string, limit int, opts *grocksdb.ReadOptions) ([]string, error) {
	handle, ok := db.Families[cf]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	var keys []string

	// Si no se pasan opciones, usamos las predeterminadas
	if opts == nil {
		opts = db.DefaultReadOptions
	}

	it := db.TransactionDB.NewIteratorCF(opts, handle)
	defer it.Close()

	startKey := []byte(startAfter)
	prefixBytes := []byte(prefix)

	if len(startKey) > 0 {
		it.Seek(startKey)
		if it.Valid() && string(it.Key().Data()) == startAfter {
			it.Next()
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
