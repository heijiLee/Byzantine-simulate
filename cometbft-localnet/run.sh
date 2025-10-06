#!/bin/bash
echo "🚀 CometBFT 로컬넷 실행"
echo "======================="

# 환경변수 설정
export CMTHOME=$(pwd)/node0

echo "📁 CMTHOME: $CMTHOME"
echo "📡 노드 시작 중..."

# 노드 실행
cometbft start --home ./node0 &
NODE_PID=$!

echo "✅ 노드 시작됨 (PID: $NODE_PID)"
echo "📖 WAL 파일: $CMTHOME/data/cs.wal/wal"
echo ""
echo "🛑 종료하려면: kill $NODE_PID"

# WAL 파일이 생성될 때까지 대기
echo "⏳ WAL 파일 생성 대기 중..."
while [ ! -f "$CMTHOME/data/cs.wal/wal" ]; do
    sleep 1
done

echo "✅ WAL 파일 생성 완료!"
echo "🎯 메시지 캡처 도구 실행:"
echo "   go run ../cmd/demo/wal_capturer.go"

# 백그라운드에서 계속 실행
wait $NODE_PID
