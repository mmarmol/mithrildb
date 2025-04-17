package db

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/linxGnu/grocksdb"
)

const CFSystemExpiration = "system.expiration"

func (db *DB) WriteTTLInTxn(txn *grocksdb.Transaction, cfName, key string, expiration int64) error {
	if expiration <= 0 {
		return nil // no TTL to write
	}

	handle, err := db.GetOrCreateSystemIndex(CFSystemExpiration)
	if err != nil {
		return err
	}

	ttlKey := fmt.Sprintf("%d:%s:%s", expiration, cfName, key)
	expStr := strconv.FormatInt(expiration, 10)

	if err := txn.PutCF(handle, []byte(ttlKey), []byte(expStr)); err != nil {
		return fmt.Errorf("failed to write TTL entry: %w", err)
	}
	return nil
}

func (db *DB) ClearAllTTLInTxn(txn *grocksdb.Transaction, cfName, key string) error {
	handle, err := db.GetOrCreateSystemIndex(CFSystemExpiration)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("0:%s:%s", cfName, key)

	readOpts := grocksdb.NewDefaultReadOptions()
	defer readOpts.Destroy()
	readOpts.SetPrefixSameAsStart(true)
	readOpts.SetFillCache(false)

	iter := txn.NewIteratorCF(readOpts, handle)
	defer iter.Close()

	for iter.Seek([]byte(prefix)); iter.Valid(); iter.Next() {
		k := iter.Key()
		if !bytes.Contains(k.Data(), []byte(fmt.Sprintf(":%s:%s", cfName, key))) {
			break
		}
		if err := txn.DeleteCF(handle, k.Data()); err != nil {
			k.Free()
			return fmt.Errorf("failed to delete TTL entry for key=%q: %w", key, err)
		}
		k.Free()
	}
	return nil
}

func (db *DB) ReplaceTTLInTxn(txn *grocksdb.Transaction, cfName, key string, expiration int64) error {
	if err := db.ClearAllTTLInTxn(txn, cfName, key); err != nil {
		return err
	}
	if err := db.WriteTTLInTxn(txn, cfName, key, expiration); err != nil {
		return err
	}
	return nil
}
