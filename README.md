# 🛡️ MithrilDB

**MithrilDB** is a lightweight, document-oriented database built on top of [RocksDB](https://github.com/facebook/rocksdb). It provides a simple HTTP API for interacting with key-value data stored as JSON, with a focus on performance, multi-threaded access, and future support for clustering and high availability.

---

## ⚙️ Requirements

- [Go 1.22+](https://go.dev/dl/)
- GCC / build-essential
- Compression libraries:
  - `zlib1g-dev`, `libbz2-dev`, `libsnappy-dev`, `liblz4-dev`, `libzstd-dev`
---

## 📦 Setup

### 1. Clone and build RocksDB

MithrilDB depends on a locally compiled version of RocksDB. To install it:

```bash
git clone --depth 1 --branch v9.11.2 https://github.com/facebook/rocksdb.git ~/dev/rocksdb
cd ~/dev/rocksdb
make static_lib
```

### 2. Link RocksDB build inside the mithrildb folder

```bash
ln -s ~/dev/rocksdb rocksdb
echo "rocksdb" >> .gitignore
```

### 3. Build proyect

```bash
./build_local.sh
```

### 4. Run proyect

```bash
./bin/mithril
```

### 4. Test proyect

```bash
./test_local.sh
```