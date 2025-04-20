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

func (db *DB) WriteTTL(cfName, key string, expiration int64) error {
	if expiration <= 0 {
		return nil // no TTL to write
	}

	handle, err := db.EnsureSystemColumnFamily(CFSystemExpiration)
	if err != nil {
		return err
	}

	ttlKey := fmt.Sprintf("%d:%s:%s", expiration, cfName, key)
	expStr := strconv.FormatInt(expiration, 10)

	return db.TransactionDB.PutCF(db.DefaultWriteOptions, handle, []byte(ttlKey), []byte(expStr))
}

func (db *DB) ClearAllTTL(cfName, key string) error {
	handle, err := db.EnsureSystemColumnFamily(CFSystemExpiration)
	if err != nil {
		return err
	}

	readOpts := grocksdb.NewDefaultReadOptions()
	defer readOpts.Destroy()
	readOpts.SetPrefixSameAsStart(true)
	readOpts.SetFillCache(false)

	iter := db.TransactionDB.NewIteratorCF(readOpts, handle)
	defer iter.Close()

	prefix := fmt.Sprintf("0:%s:%s", cfName, key)

	for iter.Seek([]byte(prefix)); iter.Valid(); iter.Next() {
		k := iter.Key()
		if !bytes.Contains(k.Data(), []byte(fmt.Sprintf(":%s:%s", cfName, key))) {
			k.Free()
			break
		}
		if err := db.TransactionDB.DeleteCF(db.DefaultWriteOptions, handle, k.Data()); err != nil {
			k.Free()
			return fmt.Errorf("failed to delete TTL entry for key=%q: %w", key, err)
		}
		k.Free()
	}
	return nil
}

func (db *DB) ReplaceTTL(cfName, key string, expiration int64) error {
	if err := db.ClearAllTTL(cfName, key); err != nil {
		return err
	}
	return db.WriteTTL(cfName, key, expiration)
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

	iter.Seek([]byte(fmt.Sprintf("%010d:", 0)))

	count := 0

	for ; iter.Valid(); iter.Next() {
		if count >= limit {
			break
		}

		k := iter.Key()
		ttlKey := append([]byte{}, k.Data()...)
		k.Free()

		parts := bytes.SplitN(ttlKey, []byte(":"), 3)
		if len(parts) != 3 {
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
			}
			continue
		}

		if doc.Meta.Expiration != ts {
			continue
		}

		err = db.DeleteDocument(DocumentDeleteOptions{
			ColumnFamily: cfName,
			Key:          docKey,
			WriteOptions: db.DefaultWriteOptions,
		})
		if err != nil {
			log.Printf("[expiration] failed to delete expired %s:%s: %v", cfName, docKey, err)
			continue
		}

		count++
	}

	return count, nil
}
