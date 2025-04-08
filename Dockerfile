# ---------- Step 1: Compile ----------
FROM ubuntu:25.04 AS builder

RUN apt-get update && apt-get install -y \
  libgflags-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev \
  libsnappy-dev \
  git build-essential golang wget

# Install Go
ENV GO_VERSION=1.22.2
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:$PATH"

# Clone and compile RocksDB
RUN git clone --depth 1 --branch v9.11.2 https://github.com/facebook/rocksdb.git
WORKDIR /rocksdb
RUN make static_lib

# Copy go source
WORKDIR /app
COPY app/ .

# Compile CGo and RocksDB
ENV CGO_ENABLED=1 \
    CGO_CFLAGS="-I/rocksdb/include" \
    CGO_LDFLAGS="-L/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd"
RUN go mod tidy && go build -o mithril .

# ---------- Step 2: Final Image ----------
FROM ubuntu:25.04

RUN apt-get update && apt-get install -y \
  libgflags-dev zlib1g libbz2-1.0 liblz4-1 libzstd1 libsnappy1v5 && \
  rm -rf /var/lib/apt/lists/*

# Install Folder
WORKDIR /srv/mithril

# Copy binaries from builder
COPY --from=builder /app/mithril ./mithril
COPY --from=builder /rocksdb/librocksdb.a ./librocksdb.a
COPY --from=builder /rocksdb/include ./include

# Copy resources folder
COPY resources/ ./resources/

# Create data folder
RUN mkdir -p /data/db

EXPOSE 5126
CMD ["./mithril"]
