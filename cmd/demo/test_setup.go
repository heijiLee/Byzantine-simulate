package main

import (
	"byzantine-message-bridge/message/abstraction"
	"fmt"
)

func main() {
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
