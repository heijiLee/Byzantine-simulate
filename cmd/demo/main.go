package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("🎮 CometBFT Byzantine Message Bridge 데모")
	fmt.Println("=======================================")
	fmt.Println()

	// 사용 가능한 데모 목록
	fmt.Println("📋 사용 가능한 데모:")
	fmt.Println("   1. 메시지 시뮬레이션")
	fmt.Println("   2. Vote 변환 테스트")
	fmt.Println("   3. Consensus State 파싱")
	fmt.Println("   4. Consensus State 포맷터")
	fmt.Println("   5. WAL 파일 분석")
	fmt.Println("   6. 로컬넷 설정")
	fmt.Println("   7. 설정 테스트")
	fmt.Println()

	// 간단한 메시지 시뮬레이션 실행
	fmt.Println("🚀 CometBFT 메시지 시뮬레이션 시작...")
	fmt.Println()

	simulator := NewCometBFTMessageSimulator()
	simulator.RunSimulation(10 * time.Second)

	fmt.Println()
	fmt.Println("🧪 Vote 변환 테스트 실행...")
	RunVoteConversionTest()

	fmt.Println()
	fmt.Println("🔍 Consensus State 파싱 실행...")
	RunConsensusStateParser()

	fmt.Println()
	fmt.Println("📊 Consensus State 포맷터 실행...")
	RunDetailedConsensusFormatter()

	fmt.Println()
	fmt.Println("🔧 설정 테스트 실행...")
	RunSetupTest()
}
