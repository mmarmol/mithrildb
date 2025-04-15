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
DOC=$(curl -s "http://localhost:$PORT/documents?cf=logs&key=foo")
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
    -d "$MULTIPUT_PAYLOAD" "http://localhost:$PORT/documents/bulk/put?cf=logs")
[ "$MPUT" = "200" ] && echo "✅ Multiput succeeded" || (echo "❌ Multiput failed"; exit 1)

echo
echo "🔹 MULTIGET"
REQ='{"keys":["k1","k2","k3"]}'
RESP=$(curl -s -X POST -H "Content-Type: application/json" -d "$REQ" \
    "http://localhost:$PORT/documents/bulk/get?cf=logs")
echo "MultiGet: $RESP"
echo "$RESP" | grep -q '"k1":' && echo "✅ k1 ok" || echo "❌ k1 missing"
echo "$RESP" | grep -q '"k2":' && echo "✅ k2 ok" || echo "❌ k2 missing"
echo "$RESP" | grep -q '"k3":null' && echo "✅ k3 null" || echo "❌ k3 unexpected"

# -----------------------------------
# INSERT
# -----------------------------------

echo
echo "🔹 Test INSERT document (should succeed)"
INSERT_BODY='{"value":"initial"}'
INSERT_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null -X POST \
  -H "Content-Type: application/json" \
  -d "$INSERT_BODY" \
  "http://localhost:$PORT/documents/insert?cf=logs&key=insert-key")

if [ "$INSERT_RESPONSE" = "200" ]; then
  echo "✅ Insert succeeded"
else
  echo "❌ Insert failed with status $INSERT_RESPONSE"
  exit 1
fi

echo
echo "🔹 Test INSERT again with same key (should fail)"
INSERT_CONFLICT=$(curl -s -w "%{http_code}" -o /dev/null -X POST \
  -H "Content-Type: application/json" \
  -d "$INSERT_BODY" \
  "http://localhost:$PORT/documents/insert?cf=logs&key=insert-key")

if [ "$INSERT_CONFLICT" = "409" ]; then
  echo "✅ Insert conflict correctly detected"
else
  echo "❌ Insert conflict not handled (expected 409, got $INSERT_CONFLICT)"
  exit 1
fi

echo
echo "🔹 GET inserted document"
INSERTED_DOC=$(curl -s "http://localhost:$PORT/documents?cf=logs&key=insert-key")
echo "Response: $INSERTED_DOC"
echo "$INSERTED_DOC" | grep -q '"value":"initial"' && echo "✅ Value is correct" || (echo "❌ Incorrect value"; exit 1)

# -----------------------------------
# LIST KEYS
# -----------------------------------
echo
echo "🔹 LIST KEYS"
KEYS=$(curl -s "http://localhost:$PORT/documents/keys?cf=logs&prefix=k")
echo "KEYS: $KEYS"

# -----------------------------------
# LIST DOCUMENTS
# -----------------------------------
echo
echo "🔹 LIST DOCUMENTS"
DOCS=$(curl -s "http://localhost:$PORT/documents/list?cf=logs&prefix=k")
echo "DOCS: $DOCS"
echo "$DOCS" | grep -q '"k1":' && echo "✅ k1 present" || echo "❌ k1 missing"
echo "$DOCS" | grep -q '"meta":' && echo "✅ metadata found" || echo "❌ metadata missing"

# -----------------------------------
# METRICS
# -----------------------------------
echo
echo "🔹 METRICS"
curl -s "http://localhost:$PORT/metrics"

# -----------------------------------
# COUNTER
# -----------------------------------
echo
echo "🔹 Create counter document"
curl -s -X POST "http://localhost:$PORT/documents?cf=logs&key=mycounter&type=counter" \
  -H "Content-Type: application/json" -d '{"value": 10}' >/dev/null
echo "✅ Counter document created with value 10"

echo
echo "🔹 Increment counter by 5"
INC_RESPONSE=$(curl -s -X POST "http://localhost:$PORT/documents/counters/delta?cf=logs&key=mycounter" \
  -H "Content-Type: application/json" -d '{"delta": 5}')
echo "Response: $INC_RESPONSE"
echo "$INC_RESPONSE" | grep -q '"new":15' && echo "✅ Counter incremented to 15" || (echo "❌ Counter increment failed"; exit 1)

echo
echo "🔹 Decrement counter by 2"
DEC_RESPONSE=$(curl -s -X POST "http://localhost:$PORT/documents/counters/delta?cf=logs&key=mycounter" \
  -H "Content-Type: application/json" -d '{"delta": -2}')
echo "Response: $DEC_RESPONSE"
echo "$DEC_RESPONSE" | grep -q '"new":13' && echo "✅ Counter decremented to 13" || (echo "❌ Counter decrement failed"; exit 1)

# -----------------------------------
# REPLACE
# -----------------------------------
echo
echo "🔹 Replace existing document"
REPLACE_OK=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$PORT/documents/replace?cf=logs&key=k1" \
  -H "Content-Type: application/json" -d '{"value": "replaced"}')
[ "$REPLACE_OK" = "200" ] && echo "✅ Replace succeeded for existing key" || (echo "❌ Replace failed"; exit 1)

echo
echo "🔹 Replace on non-existing key (should fail)"
REPLACE_FAIL=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$PORT/documents/replace?cf=logs&key=doesnotexist" \
  -H "Content-Type: application/json" -d '{"value": "something"}')
[ "$REPLACE_FAIL" = "404" ] && echo "✅ Replace correctly failed on missing key" || (echo "❌ Replace should have failed"; exit 1)


# -----------------------------------
# LIST
# -----------------------------------
echo
echo "🔹 Test List Operations"

# Limpieza por si ya existía
curl -s -X POST "http://localhost:$PORT/documents?cf=logs&key=mylist&type=list" \
     -H "Content-Type: application/json" -d '{"value": []}' >/dev/null

echo "➡️ Push values to the list"
for val in "a" "b" "c"; do
	curl -s -X POST "http://localhost:$PORT/documents/lists/push?cf=logs&key=mylist" \
	     -H "Content-Type: application/json" -d "{\"element\": \"$val\"}" >/dev/null
done

echo "➡️ Unshift value to the list"
curl -s -X POST "http://localhost:$PORT/documents/lists/unshift?cf=logs&key=mylist" \
     -H "Content-Type: application/json" -d '{"element": "x"}' >/dev/null

echo "➡️ List range (0 to end)"
RANGE=$(curl -s "http://localhost:$PORT/documents/lists/range?cf=logs&key=mylist&start=0&end=-1")
echo "List after push/unshift: $RANGE"
echo "$RANGE" | grep -q '"x"' && echo "✅ 'x' is at start" || echo "❌ 'x' missing"
echo "$RANGE" | grep -q '"c"' && echo "✅ 'c' is at end" || echo "❌ 'c' missing"

echo "➡️ Shift value"
SHIFTED=$(curl -s -X POST "http://localhost:$PORT/documents/lists/shift?cf=logs&key=mylist")
echo "Shifted value: $SHIFTED"
echo "$SHIFTED" | grep -q '"element":"x"' && echo "✅ Shifted out 'x'" || echo "❌ Unexpected shift result"

echo "➡️ Pop value"
POPPED=$(curl -s -X POST "http://localhost:$PORT/documents/lists/pop?cf=logs&key=mylist")
echo "Popped value: $POPPED"
echo "$POPPED" | grep -q '"element":"c"' && echo "✅ Popped 'c'" || echo "❌ Unexpected pop result"

# Verificar estado final de la lista
FINAL=$(curl -s "http://localhost:$PORT/documents/lists/range?cf=logs&key=mylist&start=0&end=-1")
echo "Final list: $FINAL"

# -----------------------------------
# SET
# -----------------------------------
echo
echo "🔹 Test Set Operations"

# Crear documento tipo set vacío (por si ya existía)
curl -s -X POST "http://localhost:$PORT/documents?cf=logs&key=myset&type=set" \
     -H "Content-Type: application/json" -d '{"value": []}' >/dev/null

echo "➡️ Add values to set"
for val in "red" "green" "blue"; do
    curl -s -X POST "http://localhost:$PORT/documents/sets/add?cf=logs&key=myset" \
         -H "Content-Type: application/json" -d "{\"element\": \"$val\"}" >/dev/null
done

echo "➡️ Check 'green' in set"
RESP=$(curl -s "http://localhost:$PORT/documents/sets/contains?cf=logs&key=myset&element=green")
echo "Contains green? $RESP"
echo "$RESP" | grep -q '"contains":true' && echo "✅ 'green' found" || echo "❌ 'green' missing"

echo "➡️ Remove 'green' from set"
curl -s -X POST "http://localhost:$PORT/documents/sets/remove?cf=logs&key=myset" \
     -H "Content-Type: application/json" -d '{"element": "green"}' >/dev/null

echo "➡️ Check 'green' again"
RESP=$(curl -s "http://localhost:$PORT/documents/sets/contains?cf=logs&key=myset&element=green")
echo "Contains green? $RESP"
echo "$RESP" | grep -q '"contains":false' && echo "✅ 'green' removed" || echo "❌ 'green' still present"


echo
echo "✅ All tests completed successfully."
