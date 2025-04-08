package rocks

/*
#cgo LDFLAGS: -L/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd
#cgo CFLAGS: -I/rocksdb/include
#include "rocks_wrapper.h"  // ✅ solo el header para los tipos de C
*/
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

type DB struct {
	ptr *C.rocksdb_t
}

func Open(path string) (*DB, error) {
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

	var err *C.char
	C.delete_value(db.ptr, ckey, C.size_t(len(key)), &err)
	if err != nil {
		return parseErr(err)
	}
	return nil
}

func (db *DB) Flush() {
	flushOpts := C.rocksdb_flushoptions_create()
	defer C.rocksdb_flushoptions_destroy(flushOpts)

	var err *C.char
	C.rocksdb_flush(db.ptr, flushOpts, &err)
	if err != nil {
		log.Printf("Error flushing RocksDB: %s", C.GoString(err))
	}
}

// Close safely closes the RocksDB instance
func (db *DB) Close() {
	if db.ptr != nil {
		C.rocksdb_close(db.ptr)
	}
}

func parseErr(cerr *C.char) error {
	defer C.free(unsafe.Pointer(cerr))
	return fmt.Errorf("rocksdb error: %s", C.GoString(cerr))
}
