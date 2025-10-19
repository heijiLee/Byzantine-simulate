package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

// CometBFTMessageSimulator simulates CometBFT consensus messages
type CometBFTMessageSimulator struct {
	mapper *cometbftAdapter.CometBFTMapper
	height int64
	round  int64
}

func NewCometBFTMessageSimulator() *CometBFTMessageSimulator {
	return &CometBFTMessageSimulator{
		mapper: cometbftAdapter.NewCometBFTMapper("cosmos-hub-4"),
		height: 1000,
		round:  1,
	}
}

func (ms *CometBFTMessageSimulator) RunSimulation(duration time.Duration) {
	fmt.Println("🚀 CometBFT 실시간 메시지 시뮬레이션 시작")
	fmt.Println("=====================================")
	fmt.Printf("⏱️  실행 시간: %v\n", duration)
	fmt.Println()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(duration)
	messageCount := 0

	for {
		select {
		case <-ticker.C:
			messageCount++
			ms.generateAndProcessMessage(messageCount)
		case <-timeout:
			fmt.Printf("\n✅ 시뮬레이션 완료! 총 %d개 메시지 처리\n", messageCount)
			return
		}
	}
}

func (ms *CometBFTMessageSimulator) generateAndProcessMessage(count int) {
	// CometBFT 메시지 타입 선택
	msgTypes := []string{"proposal", "prevote", "precommit", "new_round_step"}
	msgType := msgTypes[rand.Intn(len(msgTypes))]

	fmt.Printf("📨 메시지 #%d: CometBFT %s 메시지 생성\n", count, msgType)

	// 원본 메시지 생성
	rawMsg := ms.generateRawMessage(msgType)

	// RawCometBFT 메시지 출력
	fmt.Printf("   📋 RawCometBFT 메시지:\n")
	printRawMessage(rawMsg)

	// Canonical로 변환
	fmt.Printf("   🔄 RawCometBFT → Canonical 변환 중...\n")
	canonical, err := ms.mapper.ToCanonical(rawMsg)
	if err != nil {
		fmt.Printf("   ❌ 변환 실패: %v\n", err)
		return
	}

	// Canonical 메시지 출력
	fmt.Printf("   📋 Canonical 메시지:\n")
	printCanonicalMessage(canonical)

	// 다시 RawCometBFT로 변환
	fmt.Printf("   🔄 Canonical → RawCometBFT 변환 중...\n")
	targetRaw, err := ms.mapper.FromCanonical(canonical)
	if err != nil {
		fmt.Printf("   ❌ RawCometBFT 변환 실패: %v\n", err)
		return
	}

	// 변환된 RawCometBFT 메시지 출력
	fmt.Printf("   📋 변환된 RawCometBFT 메시지:\n")
	printRawMessage(*targetRaw)

	fmt.Printf("   ✅ 변환 완료: %s\n", targetRaw.MessageType)
	fmt.Println()

	// 높이 증가
	ms.height++
	if ms.height%10 == 0 {
		ms.round++
	}
}

func (ms *CometBFTMessageSimulator) generateRawMessage(msgType string) abstraction.RawConsensusMessage {
	// 메시지 타입을 숫자로 변환
	var typeNum int32
	switch msgType {
	case "proposal":
		typeNum = 32 // Proposal 타입
	case "prevote":
		typeNum = 1 // Prevote 타입
	case "precommit":
		typeNum = 2 // Precommit 타입
	case "new_round_step":
		typeNum = 0 // NewRoundStep 타입
	default:
		typeNum = 0
	}

	baseMsg := map[string]interface{}{
		"height":    fmt.Sprintf("%d", ms.height), // 문자열로 변환
		"round":     fmt.Sprintf("%d", ms.round),  // 문자열로 변환
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      typeNum, // 숫자로 변환
	}

	// CometBFT 특화 필드 추가 (mapper가 기대하는 필드명 사용)
	baseMsg["block_id"] = map[string]interface{}{
		"hash": fmt.Sprintf("0x%x", rand.Int63()),
		"parts": map[string]interface{}{
			"total": 1,
			"hash":  fmt.Sprintf("0x%x", rand.Int63()),
		},
	}
	baseMsg["proposer_address"] = fmt.Sprintf("node%d", rand.Intn(10)+1)
	baseMsg["validator_address"] = fmt.Sprintf("validator%d", rand.Intn(10)+1)
	baseMsg["signature"] = fmt.Sprintf("sig_%d", rand.Int63())

	payload, _ := json.Marshal(baseMsg)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: msgType,
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}
