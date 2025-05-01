#!/usr/bin/env bash

PID_FILE=client.pid
QUERY_URL="127.0.0.1:8070/query" # Replica가 Query 요청을 처리하는 URL

# Query를 보내고 TPS와 Latency를 출력
function query_replica() {
    echo "Querying replica for TPS and Latency..."
    response=$(curl -s "${QUERY_URL}")
    if [ $? -eq 0 ]; then
        echo "Query Response:"
        echo "${response}"
    else
        echo "Failed to query replica."
    fi
}

if [ ! -f "${PID_FILE}" ]; then
    echo "No client is running."
else
    # Query replica before shutting down the client
    query_replica

    # Shutdown the client
    while read pid; do
        if [ -z "${pid}" ]; then
            echo "No client is running."
        else
            kill -15 "${pid}"
            echo "Client with PID ${pid} shutdown."
        fi
    done < "${PID_FILE}"
    rm "${PID_FILE}"
fi