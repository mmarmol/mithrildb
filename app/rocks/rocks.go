package rocks

/*
#cgo LDFLAGS: -L/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd
#cgo CFLAGS: -I/rocksdb/include
#include "rocks_wrapper.h"  // ✅ solo el header
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type DB struct {
	ptr *C.rocksdb_t
}

type RuntimeConfig struct{}
type StaticConfig struct{}

type Config struct {
	RuntimeConfig RuntimeConfig
	StaticConfig  StaticConfig
}

func DefaultConfig() *Config {
	return &Config{
		RuntimeConfig: RuntimeConfig{},
		StaticConfig:  StaticConfig{},
	}
}

func (cfg *Config) ApplyRuntimeConfig(newCfg RuntimeConfig) {
	cfg.RuntimeConfig = newCfg
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

func (db *DB) ApplyRuntimeConfig(cfg RuntimeConfig) {}

func parseErr(cerr *C.char) error {
	defer C.free(unsafe.Pointer(cerr))
	return fmt.Errorf("rocksdb error: %s", C.GoString(cerr))
}
