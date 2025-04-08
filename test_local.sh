#!/bin/bash

set -e

IMAGE_NAME="mithrildb"
CONTAINER_NAME="mithrildb-test"
PORT=8080

echo "🛠️  Build image..."
docker build --progress=plain -t $IMAGE_NAME .

echo "🚀 Executing Container..."
docker run -d --rm --name $CONTAINER_NAME -p $PORT:8080 $IMAGE_NAME

echo "⏳ Waiting for server..."
until curl -s "http://localhost:$PORT/metrics" >/dev/null; do
    sleep 0.5
done

echo "✅ Server active. Executing tests..."

echo "🔹 Test PUT"
curl -s -X POST "http://localhost:$PORT/put?key=foo&val=bar"

echo -e "\n🔹 Test GET"
VAL=$(curl -s "http://localhost:$PORT/get?key=foo")
echo "Valor recibido: $VAL"

echo "🔹 Test DELETE"
curl -s -X POST "http://localhost:$PORT/delete?key=foo"

echo "🔹 Test GET post-delete (espera error)"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/get?key=foo")
echo "Código de estado esperado: $STATUS"

echo "🔹 Test dynamic configuration (runtime)"
curl -s -X POST "http://localhost:$PORT/config/runtime" \
     -H "Content-Type: application/json" \
     -d '{"some_field":"some_value"}'

echo "🔹 Get static configuration"
curl -s "http://localhost:$PORT/config/static"

echo "🔹 Show metrics"
curl -s "http://localhost:$PORT/metrics"

echo -e "\n🧹 Cleaning..."
docker stop $CONTAINER_NAME

echo "✅ Test finished."

