package db

import (
	"github.com/linxGnu/grocksdb"
)

func (db *DB) MultiGet(keys []string, opts *grocksdb.ReadOptions) (map[string]*string, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	byteKeys := make([][]byte, len(keys))
	for i, k := range keys {
		byteKeys[i] = []byte(k)
	}

	values, err := db.TransactionDB.MultiGet(opts, byteKeys...)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, val := range values {
			if val != nil {
				val.Free()
			}
		}
	}()

	result := make(map[string]*string, len(keys))
	for i, val := range values {
		if val == nil || val.Size() == 0 {
			result[keys[i]] = nil
		} else {
			s := string(val.Data())
			result[keys[i]] = &s
		}
	}

	return result, nil
}
