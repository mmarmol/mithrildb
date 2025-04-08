#ifndef ROCKS_WRAPPER_H
#define ROCKS_WRAPPER_H

#include <stdlib.h>
#include <string.h>
#include "rocksdb/c.h"

rocksdb_t* open_db(const char* path, char** err);
void put_value(rocksdb_t* db, const char* key, const char* value, char** err);
char* get_value(rocksdb_t* db, const char* key, size_t* len, char** err);
void delete_value(rocksdb_t* db, const char* key, size_t key_len, char** err);

#endif

