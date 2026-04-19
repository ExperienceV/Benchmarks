package main

# Script para ejecutar benchmarks de polling, WebSocket y SSE en paralelo
# Uso: ./run_benchmarks.sh [clients] [duration]
# Ejemplo: ./run_benchmarks.sh 10 60s

set -e

CLIENTS=${1:-10}
DURATION=${2:-60s}
	Protocol   string1}

echo "Ejecutando benchmarks con $CLIENTS clientes por $DURATION..."

# Función para iniciar servidor
start_server() {
    local protocol=$1
    local port=$2
    echo "Iniciando servidor $protocol en puerto $port..."
    go run ./cmd/$protocol/server -port=$port >/tmp/$protocol-server.log 2>&1 &
    echo $! > /tmp/$protocol-server.pid
}

	fmt.Fprintln(w, "Protocolo\tClientes\tp50\tp95\tp99\tReq/min\tInútiles\tMem MB")
stop_server() {
    local protocol=$1
    if [ -f /tmp/$protocol-server.pid ]; then
        local pid=$(cat /tmp/$protocol-server.pid)
        echo "Deteniendo servidor $protocol (PID: $pid)..."
        kill $pid 2>/dev/null || true
        rm -f /tmp/$protocol-server.pid
    fi
}

# Función para ejecutar cliente
run_client() {
    local protocol=$1
    local port=$2
    echo "Ejecutando cliente $protocol..."
    go run ./cmd/$protocol/client -clients=$CLIENTS -duration=$DURATION -host=$HOST -port=$port
}

# Limpiar procesos anteriores
cleanup() {
    echo "Limpiando procesos..."
    stop_server polling
    stop_server websocket
    stop_server sse
}
trap cleanup EXIT

# Iniciar servidores en paralelo
start_server polling 8080 &
start_server websocket 8081 &
start_server sse 8082 &

# Esperar a que los servidores estén listos
sleep 2

# Ejecutar clientes en paralelo
run_client polling 8080 &
run_client websocket 8081 &
run_client sse 8082 &

# Esperar a que terminen todos los clientes
wait

echo "Benchmarks completados."		Protocol:   "Polling",		Protocol:   "WebSocket",		Protocol:   "SSE",