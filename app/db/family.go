package db

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/linxGnu/grocksdb"
)

// ListFamilyNames returns the list of column family names currently loaded in memory.
func (db *DB) ListFamilyNames() (usr, sys []string) {
	names := make([]string, 0, len(db.Families))
	for name := range db.Families {
		names = append(names, name)
	}
	sort.Strings(names)
	return splitCFNamesByType(names)
}

func (db *DB) CreateFamily(cfName string) error {
	if _, exists := db.Families[cfName]; exists {
		return ErrFamilyExists
	}

	opts := grocksdb.NewDefaultOptions()
	defer opts.Destroy()

	handle, err := db.TransactionDB.CreateColumnFamily(opts, cfName)
	if err != nil {
		return err
	}

	db.Families[cfName] = handle
	return nil
}

// GetOrCreateSystemIndex ensures a system-level CF is available (created if missing).
func (db *DB) GetOrCreateSystemIndex(cfName string) (*grocksdb.ColumnFamilyHandle, error) {
	if handle, ok := db.Families[cfName]; ok {
		return handle, nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.CreateFamily(cfName); err != nil {
		if !errors.Is(err, ErrFamilyExists) {
			return nil, err
		}
	}

	if handle, ok := db.Families[cfName]; ok {
		return handle, nil
	}
	return nil, fmt.Errorf("column family %q not available after creation", cfName)
}

var userCFRegex = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
var systemCFRegex = regexp.MustCompile(`^[a-z0-9_.-]{1,64}$`)

func IsValidUserCF(name string) bool {
	return userCFRegex.MatchString(name)
}

func IsValidSystemCF(name string) bool {
	if !systemCFRegex.MatchString(name) {
		return false
	}
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return false
	}
	if !strings.Contains(name, ".") {
		return false
	}
	return true
}

func splitCFNamesByType(names []string) (userCFs, systemCFs []string) {
	for _, name := range names {
		if IsValidUserCF(name) {
			userCFs = append(userCFs, name)
		}
		if IsValidSystemCF(name) {
			systemCFs = append(userCFs, name)
		}
	}
	return
}
