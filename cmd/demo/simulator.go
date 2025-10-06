package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	besuAdapter "codec/hyperledger/besu/adapter"
	kaiaAdapter "codec/kaia/adapter"
	"codec/message/abstraction"
)

// MessageSimulator simulates real-time message flow between chains
type MessageSimulator struct {
	mappers map[string]abstraction.Mapper
	height  int64
	round   int64
}

func NewMessageSimulator() *MessageSimulator {
	return &MessageSimulator{
		mappers: map[string]abstraction.Mapper{
			"cometbft": cometbftAdapter.NewCometBFTMapper("testnet-cometbft"),
			"besu":     besuAdapter.NewBesuMapper("testnet-besu"),
			"kaia":     kaiaAdapter.NewKaiaMapper("testnet-kaia"),
		},
		height: 1000,
		round:  1,
	}
}

func (ms *MessageSimulator) RunSimulation(duration time.Duration) {
	fmt.Println("🚀 실시간 메시지 시뮬레이션 시작")
	fmt.Println("================================")
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

func (ms *MessageSimulator) generateAndProcessMessage(count int) {
	// 랜덤하게 체인 선택
	chains := []string{"cometbft", "besu", "kaia"}
	sourceChain := chains[rand.Intn(len(chains))]

	// 메시지 타입 선택
	msgTypes := []string{"proposal", "vote", "prepare", "commit"}
	msgType := msgTypes[rand.Intn(len(msgTypes))]

	fmt.Printf("📨 메시지 #%d: %s에서 %s 메시지 생성\n", count, sourceChain, msgType)

	// 원본 메시지 생성
	rawMsg := ms.generateRawMessage(sourceChain, msgType)

	// Canonical로 변환
	canonical, err := ms.mappers[sourceChain].ToCanonical(rawMsg)
	if err != nil {
		fmt.Printf("   ❌ 변환 실패: %v\n", err)
		return
	}

	fmt.Printf("   🔄 Canonical: height=%v, type=%s\n", canonical.Height, canonical.Type)

	// 다른 체인으로 라우팅 (랜덤)
	targetChains := []string{}
	for chain := range ms.mappers {
		if chain != sourceChain {
			targetChains = append(targetChains, chain)
		}
	}

	if len(targetChains) > 0 {
		targetChain := targetChains[rand.Intn(len(targetChains))]

		targetRaw, err := ms.mappers[targetChain].FromCanonical(canonical)
		if err != nil {
			fmt.Printf("   ❌ %s로 변환 실패: %v\n", targetChain, err)
			return
		}

		fmt.Printf("   📤 %s로 라우팅: %s\n", targetChain, targetRaw.MessageType)
	}

	fmt.Println()

	// 높이 증가
	ms.height++
	if ms.height%10 == 0 {
		ms.round++
	}
}

func (ms *MessageSimulator) generateRawMessage(chain, msgType string) abstraction.RawConsensusMessage {
	baseMsg := map[string]interface{}{
		"height":    ms.height,
		"round":     ms.round,
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      msgType,
	}

	var payload []byte
	var chainType abstraction.ChainType

	switch chain {
	case "cometbft":
		chainType = abstraction.ChainTypeCometBFT
		baseMsg["block_hash"] = fmt.Sprintf("0x%x", rand.Int63())
		baseMsg["proposer"] = fmt.Sprintf("node%d", rand.Intn(10)+1)
		baseMsg["validator"] = fmt.Sprintf("validator%d", rand.Intn(10)+1)
		baseMsg["signature"] = fmt.Sprintf("sig_%d", rand.Int63())

	case "besu":
		chainType = abstraction.ChainTypeHyperledger
		baseMsg["block_number"] = ms.height
		baseMsg["round_number"] = ms.round
		baseMsg["block_hash"] = fmt.Sprintf("0x%x", rand.Int63())
		baseMsg["proposer"] = fmt.Sprintf("validator%d", rand.Intn(4)+1)
		baseMsg["gas_limit"] = 8000000
		baseMsg["gas_used"] = rand.Intn(4000000) + 1000000

	case "kaia":
		chainType = abstraction.ChainTypeKaia
		baseMsg["block_number"] = ms.height
		baseMsg["round_number"] = ms.round
		baseMsg["block_hash"] = fmt.Sprintf("0x%x", rand.Int63())
		baseMsg["proposer"] = fmt.Sprintf("validator%d", rand.Intn(21)+1)
		baseMsg["gas_limit"] = 8000000
		baseMsg["consensus_type"] = "istanbul"
		baseMsg["governance_id"] = "governance-1"
	}

	payload, _ = json.Marshal(baseMsg)

	return abstraction.RawConsensusMessage{
		ChainType:   chainType,
		ChainID:     fmt.Sprintf("testnet-%s", chain),
		MessageType: msgType,
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func main() {
	fmt.Println("🎮 Byzantine Message Bridge 실시간 시뮬레이터")
	fmt.Println("=============================================")

	simulator := NewMessageSimulator()

	// 30초간 시뮬레이션 실행
	simulator.RunSimulation(30 * time.Second)
}
