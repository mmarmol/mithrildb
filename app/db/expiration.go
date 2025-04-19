package db

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/linxGnu/grocksdb"
)

const CFSystemExpiration = "system.expiration"

func (db *DB) WriteTTLInTxn(txn *grocksdb.Transaction, cfName, key string, expiration int64) error {
	if expiration <= 0 {
		return nil // no TTL to write
	}

	handle, err := db.EnsureSystemColumnFamily(CFSystemExpiration)
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
	handle, err := db.EnsureSystemColumnFamily(CFSystemExpiration)
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

func (db *DB) ProcessExpiredBatch(now int64, limit int) (int, error) {
	handle, ok := db.Families[CFSystemExpiration]
	if !ok {
		return 0, fmt.Errorf("expiration index not found")
	}

	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetFillCache(false)
	defer readOpts.Destroy()

	iter := db.TransactionDB.NewIteratorCF(readOpts, handle)
	defer iter.Close()

	// Empieza a leer desde la clave más baja posible (timestamp=0)
	prefix := fmt.Sprintf("%010d:", 0)
	iter.Seek([]byte(prefix))

	count := 0

	for ; iter.Valid(); iter.Next() {
		if count >= limit {
			break
		}

		keySlice := iter.Key()
		ttlKey := append([]byte{}, keySlice.Data()...)
		keySlice.Free()

		parts := bytes.SplitN(ttlKey, []byte(":"), 3)
		if len(parts) < 3 {
			continue
		}

		ts, err := strconv.ParseInt(string(parts[0]), 10, 64)
		if err != nil || ts > now {
			break
		}

		cfName := string(parts[1])
		docKey := string(parts[2])

		doc, err := db.GetDocument(DocumentReadOptions{
			ColumnFamily: cfName,
			Key:          docKey,
			ReadOptions:  readOpts,
		})
		if err != nil {
			if !errors.Is(err, ErrKeyNotFound) {
				log.Printf("[expiration] failed to read %s:%s: %v", cfName, docKey, err)
				continue
			}
		} else {
			if doc.Meta.Expiration != ts {
				continue
			}
		}

		if err := db.DeleteDocument(DocumentDeleteOptions{
			ColumnFamily: cfName,
			Key:          docKey,
			WriteOptions: db.DefaultWriteOptions,
		}); err != nil {
			log.Printf("[expiration] failed to delete %s:%s: %v", cfName, docKey, err)
			continue
		}

		count++
	}

	return count, nil
}
