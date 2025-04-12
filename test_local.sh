#!/bin/bash

set -e

IMAGE_NAME="mithrildb"
CONTAINER_NAME="mithrildb-test"
PORT=5126

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
# Create column family
# -----------------------------------
echo
echo "🔹 Ensure column family 'logs' exists"
CF_CREATE_PAYLOAD='{"name": "logs"}'
CF_CREATE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" \
    -d "$CF_CREATE_PAYLOAD" "http://localhost:$PORT/families")
if [ "$CF_CREATE_RESPONSE" = "201" ]; then
    echo "✅ Column family 'logs' created"
else
    echo "❌ Failed to create column family 'logs' (status $CF_CREATE_RESPONSE)"
    exit 1
fi

# -----------------------------------
# PUT
# -----------------------------------
echo
echo "🔹 PUT document"
PUT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"value":"bar"}' \
    "http://localhost:$PORT/documents?cf=logs&key=foo")
if [ "$PUT_RESPONSE" = "200" ]; then
    echo "✅ PUT successful"
else
    echo "❌ PUT failed with status $PUT_RESPONSE"
    exit 1
fi

# -----------------------------------
# GET
# -----------------------------------
echo
echo "🔹 GET document"
DOC=$(curl -s "http://localhost:$PORT/documents/foo?cf=logs")
echo "Response: $DOC"
echo "$DOC" | grep -q '"value":"bar"' && echo "✅ Value is bar" || (echo "❌ Value incorrect"; exit 1)
CAS=$(echo "$DOC" | grep -o '"rev":"[^"]*' | cut -d':' -f2 | tr -d '"')

# -----------------------------------
# CAS Check
# -----------------------------------
echo
echo "🔹 Test CAS (should succeed)"
CAS_OK_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"value":"newval"}' \
    "http://localhost:$PORT/documents?cf=logs&key=foo&cas=$CAS")
[ "$CAS_OK_RESPONSE" = "200" ] && echo "✅ CAS update succeeded" || (echo "❌ CAS update failed"; exit 1)

echo
echo "🔹 Test CAS (should fail)"
CAS_FAIL_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"value":"bad-update"}' \
    "http://localhost:$PORT/documents?cf=logs&key=foo&cas=123")
[ "$CAS_FAIL_RESPONSE" = "412" ] && echo "✅ CAS conflict detected" || (echo "❌ CAS test failed"; exit 1)

# -----------------------------------
# MULTIPUT / MULTIGET
# -----------------------------------
echo
echo "🔹 MULTIPUT"
MULTIPUT_PAYLOAD='{
  "k1": { "value": "hello", "type": "json" },
  "k2": { "value": [1, 2, 3], "type": "list" }
}'
MPUT=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "$MULTIPUT_PAYLOAD" "http://localhost:$PORT/documents/bulk?cf=logs")
[ "$MPUT" = "200" ] && echo "✅ Multiput succeeded" || (echo "❌ Multiput failed"; exit 1)

echo
echo "🔹 MULTIGET"
REQ='{"keys":["k1","k2","k3"]}'
RESP=$(curl -s -X POST -H "Content-Type: application/json" -d "$REQ" \
    "http://localhost:$PORT/documents/get?cf=logs")
echo "MultiGet: $RESP"
echo "$RESP" | grep -q '"k1":' && echo "✅ k1 ok" || echo "❌ k1 missing"
echo "$RESP" | grep -q '"k2":' && echo "✅ k2 ok" || echo "❌ k2 missing"
echo "$RESP" | grep -q '"k3":null' && echo "✅ k3 null" || echo "❌ k3 unexpected"

# -----------------------------------
# LIST KEYS
# -----------------------------------
echo
echo "🔹 LIST KEYS"
KEYS=$(curl -s "http://localhost:$PORT/keys?cf=logs&prefix=k")
echo "KEYS: $KEYS"

# -----------------------------------
# LIST DOCUMENTS
# -----------------------------------
echo
echo "🔹 LIST DOCUMENTS"
DOCS=$(curl -s "http://localhost:$PORT/documents?cf=logs&prefix=k")
echo "DOCS: $DOCS"
echo "$DOCS" | grep -q '"k1":' && echo "✅ k1 present" || echo "❌ k1 missing"
echo "$DOCS" | grep -q '"meta":' && echo "✅ metadata found" || echo "❌ metadata missing"

# -----------------------------------
# METRICS
# -----------------------------------
echo
echo "🔹 METRICS"
curl -s "http://localhost:$PORT/metrics"

echo
echo "✅ All tests completed successfully."
