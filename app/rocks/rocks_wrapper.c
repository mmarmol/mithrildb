#include "rocks_wrapper.h"

rocksdb_t* open_db(const char* path, char** err) {
    rocksdb_options_t* options = rocksdb_options_create();
    rocksdb_options_set_create_if_missing(options, 1);
    return rocksdb_open(options, path, err);
}

void put_value(rocksdb_t* db, const char* key, const char* value, char** err) {
    rocksdb_writeoptions_t* write_opts = rocksdb_writeoptions_create();
    rocksdb_put(db, write_opts, key, strlen(key), value, strlen(value), err);
}

char* get_value(rocksdb_t* db, const char* key, size_t* len, char** err) {
    rocksdb_readoptions_t* read_opts = rocksdb_readoptions_create();
    return rocksdb_get(db, read_opts, key, strlen(key), len, err);
}

void delete_value(rocksdb_t* db, const char* key, size_t key_len, char** err) {
    rocksdb_writeoptions_t* write_opts = rocksdb_writeoptions_create();
    rocksdb_delete(db, write_opts, key, key_len, err);
    rocksdb_writeoptions_destroy(write_opts);
}