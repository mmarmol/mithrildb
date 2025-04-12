#!/bin/bash

set -e

IMAGE_NAME="mithrildb"
CONTAINER_NAME="mithrildb-test"
PORT=5126

echo "🛠️  Build image..."
docker build --progress=plain -t $IMAGE_NAME .

echo "🚀 Executing Container..."
docker run -d --rm --name $CONTAINER_NAME -p $PORT:5126 $IMAGE_NAME

echo "⏳ Waiting for server..."
until curl -s "http://localhost:$PORT/ping" >/dev/null; do
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
if [ "$STATUS" -ne 404 ]; then
    echo "❌ Esperaba 404 después de borrar, pero recibí $STATUS"
    exit 1
fi

echo "🔹 Test PING"
curl -s "http://localhost:$PORT/ping"

echo -e "\n🔹 Test HEALTH"
HEALTH=$(curl -s "http://localhost:$PORT/health")
echo "Respuesta HEALTH: $HEALTH"
echo "$HEALTH" | grep -q '"healthy"' || { echo "❌ HEALTH no contiene 'healthy'"; exit 1; }

echo -e "\n🔹 Test STATS"
STATS=$(curl -s "http://localhost:$PORT/stats")
echo "Respuesta STATS: $STATS"
echo "$STATS" | grep -q '"uptime"' || { echo "❌ STATS no contiene 'uptime'"; exit 1; }
echo "$STATS" | grep -q '"db_path"' || { echo "❌ STATS no contiene 'db_path'"; exit 1; }

echo -e "\n🧹 Cleaning..."
docker stop $CONTAINER_NAME

echo "✅ All tests passed successfully."