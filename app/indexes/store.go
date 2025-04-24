package indexes

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mithrildb/db"
	"mithrildb/model"

	"github.com/linxGnu/grocksdb"
)

// ValidateIndexName ensures the index name produces a valid system CF.
func ValidateIndexName(name string) error {
	cfName := model.IndexCFPrefix + name
	if !db.IsValidSystemCF(cfName) {
		return fmt.Errorf("invalid index name: must produce a valid system CF: %q", cfName)
	}
	return nil
}

func SaveIndex(database *db.DB, def *model.IndexDefinition) error {
	if err := ValidateIndexName(def.Name); err != nil {
		return err
	}

	def.CreatedAt = time.Now().UnixMilli()
	def.LastUpdatedAt = def.CreatedAt
	def.TotalIndexed = 0

	key := []byte(def.Name)
	value, err := json.Marshal(def)
	if err != nil {
		return fmt.Errorf("failed to serialize index: %w", err)
	}

	cf, err := database.EnsureSystemColumnFamily(model.CFIndexDefinitions)
	if err != nil {
		return fmt.Errorf("failed to ensure index definitions CF: %w", err)
	}

	return database.TransactionDB.PutCF(database.DefaultWriteOptions, cf, key, value)
}

func GetIndex(database *db.DB, name string) (*model.IndexDefinition, error) {
	if err := ValidateIndexName(name); err != nil {
		return nil, err
	}

	cf, ok := database.Families[model.CFIndexDefinitions]
	if !ok {
		return nil, db.ErrInvalidColumnFamily
	}

	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetFillCache(false)
	defer readOpts.Destroy()

	val, err := database.TransactionDB.GetCF(readOpts, cf, []byte(name))
	if err != nil {
		return nil, err
	}
	defer val.Free()

	if !val.Exists() {
		return nil, db.ErrKeyNotFound
	}

	var def model.IndexDefinition
	if err := json.Unmarshal(val.Data(), &def); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &def, nil
}

func DeleteIndex(database *db.DB, name string) error {
	if err := ValidateIndexName(name); err != nil {
		return err
	}

	cf, ok := database.Families[model.CFIndexDefinitions]
	if !ok {
		return db.ErrInvalidColumnFamily
	}

	return database.TransactionDB.DeleteCF(database.DefaultWriteOptions, cf, []byte(name))
}

func ListIndexes(database *db.DB) ([]*model.IndexDefinition, error) {
	cf, ok := database.Families[model.CFIndexDefinitions]
	if !ok {
		return nil, db.ErrInvalidColumnFamily
	}

	opts := grocksdb.NewDefaultReadOptions()
	opts.SetFillCache(false)
	defer opts.Destroy()

	iter := database.TransactionDB.NewIteratorCF(opts, cf)
	defer iter.Close()

	var result []*model.IndexDefinition
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		val := iter.Value()
		var def model.IndexDefinition
		if err := json.Unmarshal(val.Data(), &def); err == nil {
			result = append(result, &def)
		}
		val.Free()
	}

	return result, nil
}

func ListIndexesForCF(database *db.DB, cf string) ([]*model.IndexDefinition, error) {
	all, err := ListIndexes(database)
	if err != nil {
		return nil, err
	}

	var result []*model.IndexDefinition
	for _, def := range all {
		if strings.HasPrefix(def.Name, cf+".") {
			result = append(result, def)
		}
	}

	return result, nil
}
