package db

import (
	"fmt"

	"github.com/linxGnu/grocksdb"
)

func (db *DB) MultiGet(cf string, keys []string, opts *grocksdb.ReadOptions) (map[string]*string, error) {
	handle, ok := db.Families[cf]
	if !ok {
		return nil, fmt.Errorf("column family '%s' does not exist", cf)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	byteKeys := make([][]byte, len(keys))
	for i, k := range keys {
		byteKeys[i] = []byte(k)
	}

	values, err := db.TransactionDB.MultiGetWithCF(opts, handle, byteKeys...)
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
