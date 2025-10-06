package main

import (
	"encoding/json"
	"fmt"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("🎯 CometBFT 실제 메시지 시뮬레이터")
	fmt.Println("==================================")

	// CometBFT 매퍼 생성
	mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")

	// 실제 CometBFT 메시지 패턴 시뮬레이션
	fmt.Println("\n🔄 실제 합의 프로세스 시뮬레이션...")

	height := int64(1000000) // 실제 Cosmos Hub 높이
	round := int32(0)

	// 1. NewRoundStep - 라운드 시작
	fmt.Println("\n📦 1. NewRoundStep 메시지")
	newRoundStep := createRealNewRoundStep(height, round)
	testMessageConversion(mapper, newRoundStep)

	// 2. Proposal - 블록 제안
	fmt.Println("\n📦 2. Proposal 메시지")
	proposal := createRealProposal(height, round)
	testMessageConversion(mapper, proposal)

	// 3. BlockPart - 블록 조각들
	fmt.Println("\n📦 3. BlockPart 메시지들")
	for i := 0; i < 3; i++ {
		blockPart := createRealBlockPart(height, round, uint32(i))
		testMessageConversion(mapper, blockPart)
	}

	// 4. Vote (Prevote) - 투표들
	fmt.Println("\n📦 4. Vote (Prevote) 메시지들")
	for i := 0; i < 4; i++ {
		vote := createRealVote(height, round, fmt.Sprintf("validator%d", i), "PrevoteType")
		testMessageConversion(mapper, vote)
	}

	// 5. Vote (Precommit) - 커밋 투표들
	fmt.Println("\n📦 5. Vote (Precommit) 메시지들")
	for i := 0; i < 4; i++ {
		vote := createRealVote(height, round, fmt.Sprintf("validator%d", i), "PrecommitType")
		testMessageConversion(mapper, vote)
	}

	// 6. NewValidBlock - 유효한 블록 알림
	fmt.Println("\n📦 6. NewValidBlock 메시지")
	newValidBlock := createRealNewValidBlock(height, round)
	testMessageConversion(mapper, newValidBlock)

	// 7. Commit - 커밋 메시지
	fmt.Println("\n📦 7. Commit 메시지")
	commit := createRealCommit(height, round)
	testMessageConversion(mapper, commit)

	fmt.Println("\n🎉 실제 메시지 시뮬레이션 완료!")
}

func testMessageConversion(mapper *cometbftAdapter.CometBFTMapper, msg abstraction.RawConsensusMessage) {
	fmt.Printf("   📤 %s 변환 테스트...\n", msg.MessageType)

	// Canonical로 변환
	canonical, err := mapper.ToCanonical(msg)
	if err != nil {
		fmt.Printf("      ❌ 변환 실패: %v\n", err)
		return
	}

	fmt.Printf("      ✅ 변환 성공!\n")
	fmt.Printf("         📊 Height: %v\n", canonical.Height)
	fmt.Printf("         📊 Round: %v\n", canonical.Round)
	fmt.Printf("         📊 Type: %s\n", canonical.Type)

	if canonical.Proposer != "" {
		fmt.Printf("         📊 Proposer: %s\n", canonical.Proposer)
	}
	if canonical.Validator != "" {
		fmt.Printf("         📊 Validator: %s\n", canonical.Validator)
	}
	if canonical.BlockHash != "" {
		fmt.Printf("         📊 BlockHash: %s\n", canonical.BlockHash)
	}
	if len(canonical.Extensions) > 0 {
		fmt.Printf("         📊 Extensions: %d개\n", len(canonical.Extensions))
		for k, v := range canonical.Extensions {
			fmt.Printf("            %s: %v\n", k, v)
		}
	}
}

func createRealNewRoundStep(height int64, round int32) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":                   height,
		"round":                    round,
		"step":                     1,
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
	// 실제 Cosmos Hub 블록 해시 패턴
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())

	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Proposal",
		"block_id": map[string]interface{}{
			"hash":      blockHash,
			"prev_hash": "0xabcdef1234567890",
			"part_set_header": map[string]interface{}{
				"total": 1,
				"hash":  []byte(blockHash),
			},
		},
		"proposer_address": "cosmos1abc123def456",
		"signature":        fmt.Sprintf("sig_proposal_%d_%d", height, round),
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
	partData := []byte(fmt.Sprintf("block_part_%d_%d_%d", height, round, partIndex))

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
		"signature":         fmt.Sprintf("%s_sig_%s_%d_%d", voteType, validator, height, round),
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

func createRealCommit(height int64, round int32) abstraction.RawConsensusMessage {
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())

	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Commit",
		"block_id": map[string]interface{}{
			"hash": blockHash,
		},
		"signatures": []map[string]interface{}{
			{
				"block_id_flag":     "BlockIDFlagCommit",
				"validator_address": "validator0",
				"timestamp":         time.Now().Format(time.RFC3339),
				"signature":         "commit_sig_0",
			},
			{
				"block_id_flag":     "BlockIDFlagCommit",
				"validator_address": "validator1",
				"timestamp":         time.Now().Format(time.RFC3339),
				"signature":         "commit_sig_1",
			},
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "Commit",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}
