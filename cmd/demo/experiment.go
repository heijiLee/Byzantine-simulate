package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	besuAdapter "codec/hyperledger/besu/adapter"
	kaiaAdapter "codec/kaia/adapter"
	"codec/message/abstraction"
	"codec/message/abstraction/validator"
)

func main() {
	fmt.Println("🔬 Byzantine Message Bridge 실험")
	fmt.Println("=====================================")

	// 1. 기본 변환 테스트
	testBasicConversion()

	// 2. 크로스체인 변환 테스트
	testCrossChainConversion()

	// 3. 검증 테스트
	testValidation()

	// 4. 실제 메시지 시뮬레이션
	testRealWorldScenario()

	fmt.Println("\n✅ 모든 실험이 완료되었습니다!")
}

func testBasicConversion() {
	fmt.Println("\n📋 1. 기본 변환 테스트")
	fmt.Println("----------------------")

	// CometBFT 메시지 생성
	cometbftMapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	fmt.Printf("📤 원본 CometBFT 메시지:\n")
	printJSON(rawMsg)

	// Canonical로 변환
	canonical, err := cometbftMapper.ToCanonical(rawMsg)
	if err != nil {
		log.Printf("변환 실패: %v", err)
		return
	}

	fmt.Printf("\n🔄 Canonical 메시지로 변환:\n")
	printJSON(canonical)

	// 다시 CometBFT로 변환
	rawBack, err := cometbftMapper.FromCanonical(canonical)
	if err != nil {
		log.Printf("역변환 실패: %v", err)
		return
	}

	fmt.Printf("\n📥 다시 CometBFT로 변환:\n")
	printJSON(rawBack)
}

func testCrossChainConversion() {
	fmt.Println("\n🌉 2. 크로스체인 변환 테스트")
	fmt.Println("---------------------------")

	// CometBFT -> Besu 변환
	cometbftMapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")
	besuMapper := besuAdapter.NewBesuMapper("testnet-besu")

	// CometBFT 메시지
	cometbftRaw := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	fmt.Printf("📤 CometBFT 메시지:\n")
	printJSON(cometbftRaw)

	// Canonical로 변환
	canonical, err := cometbftMapper.ToCanonical(cometbftRaw)
	if err != nil {
		log.Printf("CometBFT -> Canonical 실패: %v", err)
		return
	}

	fmt.Printf("\n🔄 Canonical 메시지:\n")
	printJSON(canonical)

	// Besu로 변환
	besuRaw, err := besuMapper.FromCanonical(canonical)
	if err != nil {
		log.Printf("Canonical -> Besu 실패: %v", err)
		return
	}

	fmt.Printf("\n📥 Besu 메시지:\n")
	printJSON(besuRaw)

	// 검증: 다시 Canonical로 변환해서 높이가 같은지 확인
	besuCanonical, err := besuMapper.ToCanonical(*besuRaw)
	if err != nil {
		log.Printf("Besu -> Canonical 실패: %v", err)
		return
	}

	if canonical.Height.Cmp(besuCanonical.Height) == 0 {
		fmt.Printf("\n✅ 높이 보존 확인: %v\n", canonical.Height)
	} else {
		fmt.Printf("\n❌ 높이 불일치: %v != %v\n", canonical.Height, besuCanonical.Height)
	}
}

func testValidation() {
	fmt.Println("\n✅ 3. 검증 테스트")
	fmt.Println("------------------")

	validator := validator.NewValidator(abstraction.ChainTypeCometBFT)

	// 유효한 메시지
	validMsg := &abstraction.CanonicalMessage{
		ChainID:   "testnet-cometbft",
		Height:    big.NewInt(1000),
		Round:     big.NewInt(1),
		Timestamp: time.Now(),
		Type:      abstraction.MsgTypeProposal,
		Proposer:  "node1",
		Signature: "sig123",
	}

	fmt.Printf("📋 유효한 메시지 검증:\n")
	printJSON(validMsg)

	err := validator.Validate(validMsg)
	if err != nil {
		fmt.Printf("❌ 검증 실패: %v\n", err)
	} else {
		fmt.Printf("✅ 검증 성공!\n")
	}

	// 무효한 메시지 (필수 필드 누락)
	invalidMsg := &abstraction.CanonicalMessage{
		ChainID: "testnet-cometbft",
		// Height, Round, Timestamp, Type 누락
	}

	fmt.Printf("\n📋 무효한 메시지 검증 (필수 필드 누락):\n")
	printJSON(invalidMsg)

	err = validator.Validate(invalidMsg)
	if err != nil {
		fmt.Printf("✅ 예상대로 검증 실패: %v\n", err)
	} else {
		fmt.Printf("❌ 예상과 다르게 검증 성공\n")
	}
}

func testRealWorldScenario() {
	fmt.Println("\n🌍 4. 실제 시나리오 시뮬레이션")
	fmt.Println("----------------------------")

	// 여러 체인의 메시지를 동시에 처리하는 시나리오
	scenarios := []struct {
		name   string
		chain  string
		mapper abstraction.Mapper
		raw    abstraction.RawConsensusMessage
	}{
		{
			name:   "CometBFT Proposal",
			chain:  "cometbft",
			mapper: cometbftAdapter.NewCometBFTMapper("testnet-cometbft"),
			raw: abstraction.RawConsensusMessage{
				ChainType:   abstraction.ChainTypeCometBFT,
				ChainID:     "testnet-cometbft",
				MessageType: "Proposal",
				Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
				Encoding:    "json",
				Timestamp:   time.Now(),
			},
		},
		{
			name:   "Besu Proposal",
			chain:  "besu",
			mapper: besuAdapter.NewBesuMapper("testnet-besu"),
			raw: abstraction.RawConsensusMessage{
				ChainType:   abstraction.ChainTypeHyperledger,
				ChainID:     "testnet-besu",
				MessageType: "PROPOSAL",
				Payload:     []byte(`{"height":1000,"round":0,"block_hash":"0xdef456","signature":"0x123456","code":0}`),
				Encoding:    "rlp",
				Timestamp:   time.Now(),
			},
		},
		{
			name:   "Kaia Proposal",
			chain:  "kaia",
			mapper: kaiaAdapter.NewKaiaMapper("testnet-kaia"),
			raw: abstraction.RawConsensusMessage{
				ChainType:   abstraction.ChainTypeKaia,
				ChainID:     "testnet-kaia",
				MessageType: "PROPOSAL",
				Payload:     []byte(`{"block_number":1000,"round_number":1,"type":"PROPOSAL","block_hash":"0x456def","proposer":"validator1","gas_limit":8000000,"consensus_type":"istanbul","timestamp":"2024-01-01T00:00:00Z"}`),
				Encoding:    "json",
				Timestamp:   time.Now(),
			},
		},
	}

	fmt.Printf("🔄 다중 체인 메시지 처리 시뮬레이션:\n\n")

	for i, scenario := range scenarios {
		fmt.Printf("📋 시나리오 %d: %s\n", i+1, scenario.name)

		// 원본 메시지 출력
		fmt.Printf("   📤 원본: %s\n", scenario.raw.MessageType)

		// Canonical로 변환
		canonical, err := scenario.mapper.ToCanonical(scenario.raw)
		if err != nil {
			fmt.Printf("   ❌ 변환 실패: %v\n", err)
			continue
		}

		fmt.Printf("   🔄 Canonical: type=%s, height=%v\n", canonical.Type, canonical.Height)

		// 다른 체인으로 변환 (순환)
		nextIndex := (i + 1) % len(scenarios)
		nextMapper := scenarios[nextIndex].mapper

		otherRaw, err := nextMapper.FromCanonical(canonical)
		if err != nil {
			fmt.Printf("   ❌ 다른 체인 변환 실패: %v\n", err)
			continue
		}

		fmt.Printf("   📥 %s로 변환: %s\n", scenarios[nextIndex].name, otherRaw.MessageType)
		fmt.Println()
	}
}

func printJSON(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("JSON 변환 실패: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(jsonData))
}
