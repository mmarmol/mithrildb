package events

import (
	"fmt"
	"mithrildb/db"

	"github.com/linxGnu/grocksdb"
)

var eventCF *grocksdb.ColumnFamilyHandle

func InitEventQueue(db *db.DB) error {
	handle, err := db.EnsureSystemColumnFamily(EventQueueCF)
	if err != nil {
		return fmt.Errorf("failed to init event queue CF: %w", err)
	}
	eventCF = handle
	return nil
}
