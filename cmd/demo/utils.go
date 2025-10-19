package main

import (
	"encoding/json"
	"fmt"

	"codec/message/abstraction"
)

// printRawMessage prints a RawConsensusMessage in JSON format
func printRawMessage(raw abstraction.RawConsensusMessage) {
	fmt.Printf("      RawCometBFT Message:\n")

	// JSON으로 예쁘게 출력
	jsonData, err := json.MarshalIndent(raw, "         ", "  ")
	if err != nil {
		fmt.Printf("         Error marshaling: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(jsonData))
}

// printCanonicalMessage prints a CanonicalMessage in JSON format
func printCanonicalMessage(canonical *abstraction.CanonicalMessage) {
	fmt.Printf("      Canonical Message:\n")

	// JSON으로 예쁘게 출력
	jsonData, err := json.MarshalIndent(canonical, "         ", "  ")
	if err != nil {
		fmt.Printf("         Error marshaling: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(jsonData))
}

func RunSetupTest() {
	fmt.Println("🔧 Byzantine Message Bridge 설정 테스트")
	fmt.Println("=====================================")

	// 기본 타입 테스트
	fmt.Println("✅ ChainTypeCometBFT:", abstraction.ChainTypeCometBFT)
	fmt.Println("✅ ChainTypeHyperledger:", abstraction.ChainTypeHyperledger)
	fmt.Println("✅ ChainTypeKaia:", abstraction.ChainTypeKaia)

	fmt.Println("✅ MsgTypeProposal:", abstraction.MsgTypeProposal)
	fmt.Println("✅ MsgTypeVote:", abstraction.MsgTypeVote)
	fmt.Println("✅ MsgTypeBlock:", abstraction.MsgTypeBlock)

	fmt.Println("\n🎉 모든 기본 타입이 정상적으로 로드되었습니다!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
