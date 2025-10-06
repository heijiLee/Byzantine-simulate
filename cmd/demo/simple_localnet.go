package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("🏗️ CometBFT 간단 로컬넷 생성기")
	fmt.Println("===============================")

	// CometBFT 설치 확인
	if !isCometBFTInstalled() {
		fmt.Println("📦 CometBFT 설치 중...")
		if err := installCometBFT(); err != nil {
			fmt.Printf("❌ CometBFT 설치 실패: %v\n", err)
			return
		}
	}

	fmt.Println("✅ CometBFT 설치 확인 완료")

	// 로컬넷 생성
	fmt.Println("🌐 로컬넷 생성 중...")
	if err := createLocalnet(); err != nil {
		fmt.Printf("❌ 로컬넷 생성 실패: %v\n", err)
		return
	}

	fmt.Println("✅ 로컬넷 생성 완료!")
	fmt.Println("")
	fmt.Println("🚀 로컬넷 실행:")
	fmt.Println("   cd cometbft-localnet")
	fmt.Println("   cometbft start --home ./node0")
	fmt.Println("")
	fmt.Println("📖 WAL 메시지 캡처:")
	fmt.Println("   go run ../cmd/demo/wal_capturer.go")
}

func isCometBFTInstalled() bool {
	_, err := exec.LookPath("cometbft")
	return err == nil
}

func installCometBFT() error {
	cmd := exec.Command("go", "install", "github.com/cometbft/cometbft/cmd/cometbft@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createLocalnet() error {
	// 기존 디렉토리 정리
	os.RemoveAll("cometbft-localnet")

	// 로컬넷 생성
	cmd := exec.Command("cometbft", "testnet", "--v", "1", "--o", "cometbft-localnet", "--starting-ip-address", "192.168.10.2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// 환경변수 설정을 위한 스크립트 생성
	return createRunScript()
}

func createRunScript() error {
	scriptContent := `#!/bin/bash
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
`

	return os.WriteFile("cometbft-localnet/run.sh", []byte(scriptContent), 0755)
}
