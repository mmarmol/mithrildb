package events

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

const (
	OpPut     = "put"
	OpDelete  = "delete"
	OpMutate  = "mutate"
	OpReplace = "replace"
	OpInsert  = "insert"
	OpTouch   = "touch"
)

const EventQueueCF = "system.eventqueue"

var eventQueue *RocksQueue

func InitEventQueue(q *RocksQueue) {
	eventQueue = q
}

type DocumentChangeEvent struct {
	Operation          string          `json:"operation"`
	CF                 string          `json:"cf"`
	Key                string          `json:"key"`
	Timestamp          int64           `json:"timestamp"`
	Document           *model.Document `json:"document,omitempty"`
	PreviousMeta       *model.Metadata `json:"previous_meta,omitempty"`
	ExplicitExpiration *int64          `json:"explicit_expiration,omitempty"`
}

type ChangeEventOptions struct {
	Txn                *grocksdb.Transaction
	CFName             string
	Key                string
	Document           *model.Document
	Operation          string
	PreviousMeta       *model.Metadata
	ExplicitExpiration *int64
}

func PublishChangeEvent(opts ChangeEventOptions) error {
	if eventQueue == nil {
		return fmt.Errorf("event queue is not initialized")
	}

	event := DocumentChangeEvent{
		Operation:          opts.Operation,
		CF:                 opts.CFName,
		Key:                opts.Key,
		Timestamp:          time.Now().UnixMilli(),
		Document:           opts.Document,
		PreviousMeta:       opts.PreviousMeta,
		ExplicitExpiration: opts.ExplicitExpiration,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize DocumentChangeEvent: %w", err)
	}

	return eventQueue.EnqueueWithTxn(opts.Txn, data)
}
