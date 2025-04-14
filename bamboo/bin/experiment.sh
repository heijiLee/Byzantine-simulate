#!/usr/bin/env bash

# === Config ===
TARGET_BLOCK=30
LOG_FILE="./log/server.log"
SERVER_PID_FILE="server.pid"
CLIENT_PID_FILE="client.pid"
ALGORITHM="hotstuff"
LOG_LEVEL="debug"

# === Functions ===

start_server() {
    if [ -f "$SERVER_PID_FILE" ]; then
        echo "[Server] Already running."
        return
    fi

    echo "[Server] Starting..."
    go build -o server_exec ../server/
    mkdir -p log
    ./server_exec -sim=true -log_level=${LOG_LEVEL} -algorithm=${ALGORITHM} > ${LOG_FILE} 2>&1 &
    echo $! > "${SERVER_PID_FILE}"
    echo "[Server] Started with PID $(cat ${SERVER_PID_FILE})"
}

start_client() {
    if [ -f "$CLIENT_PID_FILE" ]; then
        echo "[Client] Already running."
        return
    fi

    echo "[Client] Starting..."
    go build -o client_exec ../client/
    ./client_exec &
    echo $! > "${CLIENT_PID_FILE}"
    echo "[Client] Started with PID $(cat ${CLIENT_PID_FILE})"
}

stop_client() {
    if [ ! -f "$CLIENT_PID_FILE" ]; then
        echo "[Client] Not running."
    else
        while read -r pid; do
            if [ -n "$pid" ]; then
                kill -15 "$pid"
                echo "[Client] Killed PID $pid"
            fi
        done < "$CLIENT_PID_FILE"
        rm -f "$CLIENT_PID_FILE"
    fi
}

stop_server() {
    if [ ! -f "$SERVER_PID_FILE" ]; then
        echo "[Server] Not running."
    else
        while read -r pid; do
            if [ -n "$pid" ]; then
                kill -15 "$pid"
                echo "[Server] Killed PID $pid"
            fi
        done < "$SERVER_PID_FILE"
        rm -f "$SERVER_PID_FILE"
    fi
}

monitor_blocks() {
    echo "[Monitor] Watching for block #${TARGET_BLOCK}..."
    tail -Fn0 "${LOG_FILE}" | \
    while read -r line; do
        if echo "$line" | grep -q "${TARGET_BLOCK}th block is commited"; then
            echo "[Monitor] Target block #${TARGET_BLOCK} reached."
            stop_client
            stop_server
            exit 0
        fi
    done
}

# === Main Execution ===

start_server
sleep 1
start_client
monitor_blocks
