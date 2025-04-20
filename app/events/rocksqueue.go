package events

import (
	"encoding/binary"
	"sync"

	"github.com/linxGnu/grocksdb"
)

const (
	metaHeadKey = "__meta:head"
	metaTailKey = "__meta:tail"
)

type RocksQueue struct {
	DB        *grocksdb.TransactionDB
	CF        *grocksdb.ColumnFamilyHandle
	WriteOpts *grocksdb.WriteOptions
	ReadOpts  *grocksdb.ReadOptions

	mu   sync.Mutex
	head uint64
	tail uint64
}

func NewRocksQueue(db *grocksdb.TransactionDB, cf *grocksdb.ColumnFamilyHandle) (*RocksQueue, error) {
	q := &RocksQueue{
		DB:        db,
		CF:        cf,
		WriteOpts: grocksdb.NewDefaultWriteOptions(),
		ReadOpts:  grocksdb.NewDefaultReadOptions(),
		head:      0,
		tail:      0,
	}

	if err := q.loadPointers(); err != nil {
		return nil, err
	}

	return q, nil
}

func encodeSeq(seq uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, seq)
	return buf
}

func (q *RocksQueue) loadPointers() error {
	head, err := q.DB.GetCF(q.ReadOpts, q.CF, []byte(metaHeadKey))
	if err != nil {
		return err
	}
	defer head.Free()
	if head.Exists() {
		q.head = binary.BigEndian.Uint64(head.Data())
	}

	tail, err := q.DB.GetCF(q.ReadOpts, q.CF, []byte(metaTailKey))
	if err != nil {
		return err
	}
	defer tail.Free()
	if tail.Exists() {
		q.tail = binary.BigEndian.Uint64(tail.Data())
	}

	return nil
}

func (q *RocksQueue) savePointer(key string, value uint64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, value)
	return q.DB.PutCF(q.WriteOpts, q.CF, []byte(key), buf)
}

func (q *RocksQueue) Enqueue(data []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	key := encodeSeq(q.tail)
	if err := q.DB.PutCF(q.WriteOpts, q.CF, key, data); err != nil {
		return err
	}

	q.tail++
	return q.savePointer(metaTailKey, q.tail)
}

func (q *RocksQueue) Next() ([]byte, []byte, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	it := q.DB.NewIteratorCF(q.ReadOpts, q.CF)
	defer it.Close()

	it.Seek(encodeSeq(q.head))
	if !it.Valid() {
		return nil, nil, nil
	}

	key := it.Key()
	value := it.Value()
	defer key.Free()
	defer value.Free()

	return append([]byte{}, key.Data()...), append([]byte{}, value.Data()...), nil
}

func (q *RocksQueue) Ack(key []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if err := q.DB.DeleteCF(q.WriteOpts, q.CF, key); err != nil {
		return err
	}
	q.head++
	return q.savePointer(metaHeadKey, q.head)
}

func (q *RocksQueue) EnqueueWithTxn(txn *grocksdb.Transaction, value []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	key := encodeSeq(q.tail)
	if err := txn.PutCF(q.CF, key, value); err != nil {
		return err
	}

	q.tail++
	return q.savePointer(metaTailKey, q.tail)
}
