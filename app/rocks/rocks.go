package rocks

/*
#cgo LDFLAGS: -L/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd
#cgo CFLAGS: -I/rocksdb/include
#include "rocks_wrapper.h"  // ✅ solo el header para los tipos de C
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type DB struct {
	ptr *C.rocksdb_t
}

type Config struct {
	DBPath                         string
	Port                           int
	CreateIfMissing                bool
	ErrorIfExists                  bool
	ParanoidChecks                 bool
	WriteBufferSize                int
	MaxWriteBufferNumber           int
	MinWriteBufferNumberToMerge    int
	MaxBackgroundCompactions       int
	MaxBackgroundFlushes           int
	Compression                    string
	CompactionStyle                string
	Level0FileNumCompactionTrigger int
	Level0SlowdownWritesTrigger    int
	Level0StopWritesTrigger        int
	MaxCompactionBytes             int
	BlockSize                      int
	BlockCacheSize                 int
	BlockCacheCompressedSize       int
	BloomFilterPolicy              bool
	WriteSync                      bool
	DisableAutoCompactions         bool
	MaxOpenFiles                   int
	MaxLogFileSize                 int
	LogFileTimeToRoll              int
	ReuseLogs                      bool
	Statistics                     bool
	DisableMemtableActivity        bool
	MaxTotalWALSize                int
	WalTTLSeconds                  int
	WalSizeLimitMB                 int
	ManifestPreallocationSize      int
}

func DefaultConfig() *Config {
	return &Config{
		DBPath:                         "/data/db",
		Port:                           5126,
		CreateIfMissing:                true,
		ErrorIfExists:                  false,
		ParanoidChecks:                 false,
		WriteBufferSize:                67108864,
		MaxWriteBufferNumber:           3,
		MinWriteBufferNumberToMerge:    1,
		MaxBackgroundCompactions:       2,
		MaxBackgroundFlushes:           1,
		Compression:                    "snappy",
		CompactionStyle:                "level",
		Level0FileNumCompactionTrigger: 4,
		Level0SlowdownWritesTrigger:    8,
		Level0StopWritesTrigger:        12,
		MaxCompactionBytes:             67108864,
		BlockSize:                      16384,
		BlockCacheSize:                 2147483648,
		BlockCacheCompressedSize:       104857600,
		BloomFilterPolicy:              true,
		WriteSync:                      false,
		DisableAutoCompactions:         false,
		MaxOpenFiles:                   1000,
		MaxLogFileSize:                 1073741824,
		LogFileTimeToRoll:              3600000,
		ReuseLogs:                      true,
		Statistics:                     true,
		DisableMemtableActivity:        false,
		MaxTotalWALSize:                536870912,
		WalTTLSeconds:                  86400,
		WalSizeLimitMB:                 1024,
		ManifestPreallocationSize:      4194304,
	}
}

func Open(path string, cfg *Config) (*DB, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var err *C.char
	db := C.open_db(cpath, &err)
	if err != nil {
		return nil, parseErr(err)
	}
	return &DB{ptr: db}, nil
}

func (db *DB) ApplyConfig(cfg Config) error {
	// Traduce y aplica la configuración usando funciones C y RocksDB.
	// Ejemplo de cómo aplicar el `WriteBufferSize`:

	return nil // Actualizar para reflejar el éxito o error de la operación
}

func (db *DB) Get(key string) (string, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var err *C.char
	var clen C.size_t
	val := C.get_value(db.ptr, ckey, &clen, &err)
	if err != nil {
		return "", parseErr(err)
	}
	defer C.free(unsafe.Pointer(val))

	return C.GoStringN(val, C.int(clen)), nil
}

func (db *DB) Put(key, value string) error {
	ckey := C.CString(key)
	cval := C.CString(value)
	defer C.free(unsafe.Pointer(ckey))
	defer C.free(unsafe.Pointer(cval))

	var err *C.char
	C.put_value(db.ptr, ckey, cval, &err)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (db *DB) Delete(key string) error {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	wo := C.rocksdb_writeoptions_create()
	defer C.rocksdb_writeoptions_destroy(wo)

	var err *C.char
	C.rocksdb_delete(db.ptr, wo, ckey, C.size_t(len(key)), &err)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (db *DB) Close() {
	C.rocksdb_close(db.ptr)
}

func parseErr(cerr *C.char) error {
	defer C.free(unsafe.Pointer(cerr))
	return fmt.Errorf("rocksdb error: %s", C.GoString(cerr))
}
