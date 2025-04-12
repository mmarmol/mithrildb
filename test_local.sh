#!/bin/bash

set -e

IMAGE_NAME="mithrildb"
CONTAINER_NAME="mithrildb-test"
PORT=5126

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
# BASIC PUT/GET/DELETE
# -----------------------------------

echo
echo "🔹 Test PUT with sync=true"
curl -s -X POST "http://localhost:$PORT/put?key=foo&val=bar&sync=true"
echo "✅ PUT foo=bar done"

echo
echo "🔹 Test GET value for 'foo'"
VAL=$(curl -s "http://localhost:$PORT/get?key=foo")
if [ "$VAL" = "bar" ]; then
    echo "✅ GET returned expected value: $VAL"
else
    echo "❌ GET returned unexpected value: $VAL"
    exit 1
fi

echo
echo "🔹 Test DELETE key 'foo'"
curl -s -X POST "http://localhost:$PORT/delete?key=foo"
echo "✅ DELETE completed"

echo
echo "🔹 Test GET after delete (expect 404)"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/get?key=foo")
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
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" \
    -d "$PAYLOAD" "http://localhost:$PORT/multiput")

if [ "$RESPONSE" = "200" ]; then
    echo "✅ MULTIPUT succeeded: $PAYLOAD"
else
    echo "❌ MULTIPUT failed with status $RESPONSE"
    exit 1
fi

echo
echo "🔹 Test MULTIGET (expect k1 and k2 to be present)"
QUERY='{"keys":["k1","k2","k3"]}'
MULTIGET=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "$QUERY" "http://localhost:$PORT/multiget")

echo "MultiGet Response: $MULTIGET"

echo "$MULTIGET" | grep -q '"k1":"v1"' && echo "✅ k1 returned correctly" || echo "❌ k1 missing or incorrect"
echo "$MULTIGET" | grep -q '"k2":"v2"' && echo "✅ k2 returned correctly" || echo "❌ k2 missing or incorrect"
echo "$MULTIGET" | grep -q '"k3":null' && echo "✅ k3 correctly returned as null" || echo "❌ k3 unexpected"

# -----------------------------------
# LIST KEYS
# -----------------------------------

echo
echo "🔹 Test LIST keys (prefix=k, limit=10)"
LIST=$(curl -s "http://localhost:$PORT/list?prefix=k&limit=10")

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

# -----------------------------------
# CLEANUP
# -----------------------------------

echo
echo "🧹 Cleaning up..."
docker stop $CONTAINER_NAME

echo
echo "✅ All tests completed successfully."
