package db

// ListDocumentKeys returns keys from a column family that match the prefix and pagination options.
func (db *DB) ListDocumentKeys(opts KeyListOptions) ([]string, error) {
	handle, ok := db.Families[opts.ColumnFamily]
	if !ok {
		return nil, ErrInvalidColumnFamily
	}

	var keys []string

	if opts.ReadOptions == nil {
		opts.ReadOptions = db.DefaultReadOptions
	}

	it := db.TransactionDB.NewIteratorCF(opts.ReadOptions, handle)
	defer it.Close()

	startKey := []byte(opts.StartAfter)
	prefixBytes := []byte(opts.Prefix)

	if len(startKey) > 0 {
		it.Seek(startKey)
		if it.Valid() && string(it.Key().Data()) == opts.StartAfter {
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

		if len(prefixBytes) > 0 && !hasPrefix(k, opts.Prefix) {
			break
		}

		keys = append(keys, k)
		if len(keys) >= opts.Limit {
			break
		}
	}

	if keys == nil {
		keys = []string{}
	}
	return keys, nil
}

func hasPrefix(s, prefix string) bool {
	return len(prefix) == 0 || (len(s) >= len(prefix) && s[:len(prefix)] == prefix)
}
