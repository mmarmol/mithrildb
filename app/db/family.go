package db

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/linxGnu/grocksdb"
)

// ListColumnFamilies returns the loaded column family names, split into user/system categories.
func (db *DB) ListColumnFamilies() (userCFs, systemCFs []string) {
	names := make([]string, 0, len(db.Families))
	for name := range db.Families {
		names = append(names, name)
	}
	sort.Strings(names)
	return splitCFNamesByType(names)
}

// CreateColumnFamily creates a new column family with default RocksDB options.
func (db *DB) CreateColumnFamily(name string) error {
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

// EnsureSystemColumnFamily ensures a system-level column family exists, creating it if necessary.
func (db *DB) EnsureSystemColumnFamily(name string) (*grocksdb.ColumnFamilyHandle, error) {
	if handle, ok := db.Families[name]; ok {
		return handle, nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.CreateColumnFamily(name); err != nil {
		if !errors.Is(err, ErrFamilyExists) {
			return nil, err
		}
	}

	if handle, ok := db.Families[name]; ok {
		return handle, nil
	}
	return nil, fmt.Errorf("column family %q not available after creation", name)
}

var userCFRegex = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
var systemCFRegex = regexp.MustCompile(`^[a-z0-9_.-]{1,64}$`)

// IsValidUserCF checks whether a column family name is valid for user data.
func IsValidUserCF(name string) bool {
	return userCFRegex.MatchString(name)
}

// IsValidSystemCF checks whether a column family name is valid for internal system data.
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
			systemCFs = append(systemCFs, name)
		}
	}
	return
}
