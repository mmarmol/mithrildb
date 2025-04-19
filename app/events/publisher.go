package events

import (
	"encoding/json"
	"fmt"
	"time"

	"mithrildb/model"

	"github.com/google/uuid"
	"github.com/linxGnu/grocksdb"
)

const EventQueueCF = "system.eventqueue"

// DocumentChangeEvent represents a change made to a document.
type DocumentChangeEvent struct {
	Operation    string          `json:"operation"`
	CF           string          `json:"cf"`
	Key          string          `json:"key"`
	Timestamp    int64           `json:"timestamp"`
	Document     *model.Document `json:"document,omitempty"`
	PreviousRev  *string         `json:"previous_rev,omitempty"`
	PreviousMeta *model.Metadata `json:"previous_meta,omitempty"`
}

// ChangeEventOptions encapsulates the data required to enqueue a document change event.
type ChangeEventOptions struct {
	Txn          *grocksdb.Transaction
	CFName       string
	Key          string
	Document     *model.Document
	Operation    string
	PreviousRev  *string
	PreviousMeta *model.Metadata
}

// PublishChangeEvent creates and writes a change event to the event queue CF within the provided transaction.
func PublishChangeEvent(opts ChangeEventOptions) error {
	event := DocumentChangeEvent{
		Operation:    opts.Operation,
		CF:           opts.CFName,
		Key:          opts.Key,
		Timestamp:    time.Now().UnixMilli(),
		Document:     opts.Document,
		PreviousRev:  opts.PreviousRev,
		PreviousMeta: opts.PreviousMeta,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize DocumentChangeEvent: %w", err)
	}

	if eventCF == nil {
		return fmt.Errorf("event queue column family not initialized")
	}

	keyBytes := generateEventKey()
	return opts.Txn.PutCF(eventCF, keyBytes, data)
}

func generateEventKey() []byte {
	ts := time.Now().UnixNano()
	uid := uuid.New().String()
	return []byte(fmt.Sprintf("%020d-%s", ts, uid))
}
