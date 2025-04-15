package db

import (
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
