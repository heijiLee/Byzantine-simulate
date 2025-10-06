package main

import (
	"encoding/json"
	"fmt"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("🔍 CometBFT 실제 메시지 샘플 생성기")
	fmt.Println("=====================================")

	// 실제 CometBFT 메시지 패턴 시뮬레이션
	simulateRealConsensusFlow()
}

func simulateRealConsensusFlow() {
	fmt.Println("\n📋 실제 합의 플로우 시뮬레이션")
	fmt.Println("-------------------------------")

	mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")
	height := int64(1000000) // 실제 Cosmos Hub 높이
	round := int32(0)

	// 1. NewRoundStep - 라운드 시작
	fmt.Println("\n1️⃣ NewRoundStep (라운드 시작)")
	newRoundStep := createRealNewRoundStep(height, round)
	printMessage("NewRoundStep", newRoundStep)

	canonical, _ := mapper.ToCanonical(newRoundStep)
	fmt.Printf("   → Canonical 변환: height=%v, step=%v\n",
		canonical.Height, canonical.Extensions["step"])

	// 2. Proposal - 블록 제안
	fmt.Println("\n2️⃣ Proposal (블록 제안)")
	proposal := createRealProposal(height, round)
	printMessage("Proposal", proposal)

	canonical, _ = mapper.ToCanonical(proposal)
	fmt.Printf("   → Canonical 변환: proposer=%s, block_hash=%s\n",
		canonical.Proposer, canonical.BlockHash)

	// 3. BlockPart - 블록 조각 전송
	fmt.Println("\n3️⃣ BlockPart (블록 조각)")
	for i := 0; i < 3; i++ {
		blockPart := createRealBlockPart(height, round, uint32(i))
		printMessage(fmt.Sprintf("BlockPart[%d]", i), blockPart)
	}

	// 4. Vote (Prevote) - 투표
	fmt.Println("\n4️⃣ Vote (Prevote)")
	for i := 0; i < 5; i++ {
		vote := createRealVote(height, round, fmt.Sprintf("validator%d", i), "PrevoteType")
		printMessage(fmt.Sprintf("Prevote[%d]", i), vote)
	}

	// 5. Vote (Precommit) - 커밋 투표
	fmt.Println("\n5️⃣ Vote (Precommit)")
	for i := 0; i < 5; i++ {
		vote := createRealVote(height, round, fmt.Sprintf("validator%d", i), "PrecommitType")
		printMessage(fmt.Sprintf("Precommit[%d]", i), vote)
	}

	// 6. NewValidBlock - 유효한 블록 알림
	fmt.Println("\n6️⃣ NewValidBlock")
	newValidBlock := createRealNewValidBlock(height, round)
	printMessage("NewValidBlock", newValidBlock)

	fmt.Println("\n✅ 실제 합의 플로우 시뮬레이션 완료!")
}

func createRealNewRoundStep(height int64, round int32) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":                   height,
		"round":                    round,
		"step":                     1, // NewHeight step
		"seconds_since_start_time": 0,
		"last_commit_round":        -1,
		"message_type":             "NewRoundStep",
		"timestamp":                time.Now().Format(time.RFC3339),
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "NewRoundStep",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealProposal(height int64, round int32) abstraction.RawConsensusMessage {
	// 실제 Cosmos Hub 블록 해시 패턴 시뮬레이션
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())

	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Proposal",
		"block_id": map[string]interface{}{
			"hash":      blockHash,
			"prev_hash": "0x1234567890abcdef",
			"part_set_header": map[string]interface{}{
				"total": 1,
				"hash":  []byte(blockHash),
			},
		},
		"proposer_address": "cosmosvaloper1abc123def456",
		"signature":        fmt.Sprintf("sig_%d_%d", height, round),
		"pol_round":        -1,
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "Proposal",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealBlockPart(height int64, round int32, partIndex uint32) abstraction.RawConsensusMessage {
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	partData := []byte(fmt.Sprintf("block_part_%d_data_%d", height, partIndex))

	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "BlockPart",
		"block_id": map[string]interface{}{
			"hash": blockHash,
		},
		"part_index": partIndex,
		"part_bytes": partData,
		"part_proof": []byte(fmt.Sprintf("merkle_proof_%d", partIndex)),
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "BlockPart",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealVote(height int64, round int32, validator, voteType string) abstraction.RawConsensusMessage {
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())

	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Vote",
		"vote_type":    voteType,
		"block_id": map[string]interface{}{
			"hash": blockHash,
		},
		"validator_address": validator,
		"validator_index":   0,
		"signature":         fmt.Sprintf("%s_sig_%d_%d", voteType, height, round),
	}

	// Precommit의 경우 extension 추가
	if voteType == "PrecommitType" {
		payload["extension"] = []byte(fmt.Sprintf("extension_%s", validator))
		payload["extension_signature"] = []byte(fmt.Sprintf("ext_sig_%s", validator))
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealNewValidBlock(height int64, round int32) abstraction.RawConsensusMessage {
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())

	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "NewValidBlock",
		"block_id": map[string]interface{}{
			"hash": blockHash,
			"part_set_header": map[string]interface{}{
				"total": 1,
				"hash":  []byte(blockHash),
			},
		},
		"is_commit":   true,
		"block_parts": []string{"part1", "part2", "part3"},
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "NewValidBlock",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func printMessage(msgType string, msg abstraction.RawConsensusMessage) {
	fmt.Printf("   📤 %s:\n", msgType)
	fmt.Printf("      ChainID: %s\n", msg.ChainID)
	fmt.Printf("      MessageType: %s\n", msg.MessageType)
	fmt.Printf("      Timestamp: %s\n", msg.Timestamp.Format(time.RFC3339))

	// Payload 일부만 출력 (너무 길면 생략)
	var payload map[string]interface{}
	json.Unmarshal(msg.Payload, &payload)
	if height, ok := payload["height"]; ok {
		fmt.Printf("      Height: %v\n", height)
	}
	if round, ok := payload["round"]; ok {
		fmt.Printf("      Round: %v\n", round)
	}
}
