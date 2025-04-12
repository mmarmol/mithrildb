#!/bin/bash

set -e

IMAGE_NAME="mithrildb"
CONTAINER_NAME="mithrildb-test"
PORT=5126

# Trap to make sure we clean up Docker container even if tests fail
trap "echo 🧹 Cleaning up...; docker stop $CONTAINER_NAME" EXIT

echo "🛠️  Building Docker image..."
docker build --progress=plain -t $IMAGE_NAME .

echo "🚀 Starting container..."
docker run -d --rm --name $CONTAINER_NAME -p $PORT:5126 $IMAGE_NAME

echo "⏳ Waiting for server to become available..."
until curl -s "http://localhost:$PORT/ping" >/dev/null; do
    sleep 0.5
done

echo
echo "✅ Server is up. Running tests..."

# -----------------------------------
# CREATE COLUMN FAMILY (if needed)
# -----------------------------------

echo
echo "🔹 Ensure column family 'logs' exists"
CF_CREATE_PAYLOAD='{"name": "logs"}'  # Correct JSON structure
CF_CREATE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" \
    -d "$CF_CREATE_PAYLOAD" "http://localhost:$PORT/families")

# Check if the column family creation was successful
if [ "$CF_CREATE_RESPONSE" = "201" ]; then
    echo "✅ Column family 'logs' created or already exists."
else
    echo "❌ Failed to create or access column family 'logs' with status $CF_CREATE_RESPONSE"
    exit 1
fi

# -----------------------------------
# BASIC PUT/GET/DELETE
# -----------------------------------

echo
echo "🔹 Test PUT with sync=true to 'logs' CF"
PUT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$PORT/put?cf=logs&key=foo&val=bar&sync=true")
if [ "$PUT_RESPONSE" = "200" ]; then
    echo "✅ PUT foo=bar in 'logs' CF done"
else
    echo "❌ PUT failed with status $PUT_RESPONSE"
    exit 1
fi

echo
echo "🔹 Test GET value for 'foo' from 'logs' CF"
VAL=$(curl -s "http://localhost:$PORT/get?cf=logs&key=foo")
if [ "$VAL" = "bar" ]; then
    echo "✅ GET returned expected value: $VAL"
else
    echo "❌ GET returned unexpected value: $VAL"
    exit 1
fi

echo
echo "🔹 Test DELETE key 'foo' from 'logs' CF"
DELETE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$PORT/delete?cf=logs&key=foo")
if [ "$DELETE_RESPONSE" = "200" ]; then
    echo "✅ DELETE completed"
else
    echo "❌ DELETE failed with status $DELETE_RESPONSE"
    exit 1
fi

echo
echo "🔹 Test GET after delete from 'logs' CF (expect 404)"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/get?cf=logs&key=foo")
if [ "$STATUS" = "404" ]; then
    echo "✅ Correctly received 404 for deleted key"
else
    echo "❌ Expected 404, got $STATUS"
    exit 1
fi

# -----------------------------------
# MULTIPUT / MULTIGET
# -----------------------------------

echo
echo "🔹 Test MULTIPUT"
PAYLOAD='{"k1":"v1","k2":"v2"}'
MULTIPUT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" \
    -d "$PAYLOAD" "http://localhost:$PORT/multiput?cf=logs")

if [ "$MULTIPUT_RESPONSE" = "200" ]; then
    echo "✅ MULTIPUT succeeded: $PAYLOAD"
else
    echo "❌ MULTIPUT failed with status $MULTIPUT_RESPONSE"
    exit 1
fi

echo
echo "🔹 Test MULTIGET from 'logs' CF (expect k1 and k2 to be present)"
QUERY='{"keys":["k1","k2","k3"]}'
MULTIGET_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "$QUERY" "http://localhost:$PORT/multiget?cf=logs")

echo "MultiGet Response: $MULTIGET_RESPONSE"

echo "$MULTIGET_RESPONSE" | grep -q '"k1":"v1"' && echo "✅ k1 returned correctly" || echo "❌ k1 missing or incorrect"
echo "$MULTIGET_RESPONSE" | grep -q '"k2":"v2"' && echo "✅ k2 returned correctly" || echo "❌ k2 missing or incorrect"
echo "$MULTIGET_RESPONSE" | grep -q '"k3":null' && echo "✅ k3 correctly returned as null" || echo "❌ k3 unexpected"

# -----------------------------------
# LIST KEYS
# -----------------------------------

echo
echo "🔹 Test LIST keys from 'logs' CF (prefix=k, limit=10)"
LIST=$(curl -s "http://localhost:$PORT/list?cf=logs&prefix=k&limit=10")

echo "LIST response: $LIST"

# Count entries (number of commas + 1)
KEY_COUNT=$(echo "$LIST" | tr -cd ',' | wc -c)
KEY_COUNT=$((KEY_COUNT + 1))

echo "🔸 Keys returned: $KEY_COUNT"

# Print each key (manually, since no jq)
echo "$LIST" | sed 's/[][]//g' | tr ',' '\n' | tr -d '"' | sed '/^$/d' | nl

# Check expected keys
echo "$LIST" | grep -q '"k1"' && echo "✅ k1 appears in list" || echo "❌ k1 missing in list"
echo "$LIST" | grep -q '"k2"' && echo "✅ k2 appears in list" || echo "❌ k2 missing in list"

# -----------------------------------
# METRICS
# -----------------------------------

echo
echo "🔹 Test METRICS"
METRICS=$(curl -s "http://localhost:$PORT/metrics")

echo "$METRICS" | grep -q '"server"' && echo "✅ 'server' block found" || echo "❌ 'server' block missing"
echo "$METRICS" | grep -q '"rocksdb"' && echo "✅ 'rocksdb' block found" || echo "❌ 'rocksdb' block missing"

# Optional: show server uptime
UPTIME=$(echo "$METRICS" | grep -o '"uptime_seconds":[0-9]*' | cut -d: -f2)
echo "Server uptime: ${UPTIME}s"

echo
echo "✅ All tests completed successfully."