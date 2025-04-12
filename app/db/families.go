package db

import (
	"errors"
	"sort"

	"github.com/linxGnu/grocksdb"
)

var ErrFamilyExists = errors.New("column family already exists")

// ListFamilyNames returns the list of column family names currently loaded in memory.
func (db *DB) ListFamilyNames() []string {
	names := make([]string, 0, len(db.Families))
	for name := range db.Families {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (db *DB) CreateFamily(name string) error {
	if _, exists := db.Families[name]; exists {
		return ErrFamilyExists
	}

	opts := grocksdb.NewDefaultOptions()
	defer opts.Destroy()

	handle, err := db.TransactionDB.CreateColumnFamily(opts, name)
	if err != nil {
		return err
	}

	db.Families[name] = handle
	return nil
}
