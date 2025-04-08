#!/bin/bash
set -e

APP_DIR="./app"
ROCKSDB_DIR="./rocksdb"

if [ ! -f "$ROCKSDB_DIR/include/rocksdb/c.h" ]; then
  echo "❌ Error: $ROCKSDB_DIR/include/rocksdb/c.h not found."
  exit 1
fi

if [ ! -f "$ROCKSDB_DIR/librocksdb.a" ]; then
  echo "❌ Error: $ROCKSDB_DIR/librocksdb.a not found."
  exit 1
fi

echo "🔧 Building ..."

export CGO_CFLAGS="-I$(realpath $ROCKSDB_DIR/include)"
export CGO_LDFLAGS="-L$(realpath $ROCKSDB_DIR) -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd"

cd $APP_DIR
go build -buildvcs=false -o ../bin/mithril .
cd ..

echo "✅ Build successful"
