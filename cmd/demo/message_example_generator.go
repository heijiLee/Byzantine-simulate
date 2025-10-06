package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("💾 CometBFT 메시지 예제 저장기")
	fmt.Println("===============================")

	// 예제 메시지들 생성
	messages := generateExampleMessages()

	// JSON 파일로 저장
	saveMessagesToJSON(messages)

	// 저장된 메시지 읽기 테스트
	testSavedMessages()

	fmt.Println("\n🎉 메시지 예제 저장 및 테스트 완료!")
}

func generateExampleMessages() []abstraction.RawConsensusMessage {
	fmt.Println("\n📦 예제 메시지 생성 중...")

	var messages []abstraction.RawConsensusMessage
	height := int64(1000000)
	round := int32(0)

	// 1. NewRoundStep
	messages = append(messages, createExampleNewRoundStep(height, round))

	// 2. Proposal
	messages = append(messages, createExampleProposal(height, round))

	// 3. BlockPart (3개)
	for i := 0; i < 3; i++ {
		messages = append(messages, createExampleBlockPart(height, round, uint32(i)))
	}

	// 4. Vote (Prevote) - 4개 validator
	for i := 0; i < 4; i++ {
		messages = append(messages, createExampleVote(height, round, fmt.Sprintf("validator%d", i), "PrevoteType"))
	}

	// 5. Vote (Precommit) - 4개 validator
	for i := 0; i < 4; i++ {
		messages = append(messages, createExampleVote(height, round, fmt.Sprintf("validator%d", i), "PrecommitType"))
	}

	// 6. NewValidBlock
	messages = append(messages, createExampleNewValidBlock(height, round))

	// 7. Commit
	messages = append(messages, createExampleCommit(height, round))

	fmt.Printf("✅ %d개 메시지 생성 완료\n", len(messages))
	return messages
}

func saveMessagesToJSON(messages []abstraction.RawConsensusMessage) {
	fmt.Println("\n💾 JSON 파일로 저장 중...")

	// examples 디렉토리 생성
	examplesDir := "examples/cometbft"
	os.MkdirAll(examplesDir, 0755)

	// 각 메시지 타입별로 개별 파일 저장
	messageGroups := make(map[string][]abstraction.RawConsensusMessage)
	for _, msg := range messages {
		messageGroups[msg.MessageType] = append(messageGroups[msg.MessageType], msg)
	}

	for msgType, msgs := range messageGroups {
		filename := fmt.Sprintf("%s/%s.json", examplesDir, msgType)
		saveMessageGroup(filename, msgs)
		fmt.Printf("   📄 %s: %d개 메시지 저장\n", filename, len(msgs))
	}

	// 전체 메시지를 하나의 파일로도 저장
	allMessagesFile := fmt.Sprintf("%s/all_messages.json", examplesDir)
	saveMessageGroup(allMessagesFile, messages)
	fmt.Printf("   📄 %s: %d개 메시지 저장\n", allMessagesFile, len(messages))

	// 메시지 타입별 샘플도 저장
	saveSampleMessages(examplesDir, messages)
}

func saveMessageGroup(filename string, messages []abstraction.RawConsensusMessage) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ 파일 생성 실패: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(messages)
}

func saveSampleMessages(examplesDir string, messages []abstraction.RawConsensusMessage) {
	fmt.Println("\n📋 샘플 메시지 저장 중...")

	// 각 타입별로 첫 번째 메시지만 샘플로 저장
	seenTypes := make(map[string]bool)
	var samples []abstraction.RawConsensusMessage

	for _, msg := range messages {
		if !seenTypes[msg.MessageType] {
			samples = append(samples, msg)
			seenTypes[msg.MessageType] = true
		}
	}

	samplesFile := fmt.Sprintf("%s/samples.json", examplesDir)
	saveMessageGroup(samplesFile, samples)
	fmt.Printf("   📄 %s: %d개 샘플 메시지 저장\n", samplesFile, len(samples))
}

func testSavedMessages() {
	fmt.Println("\n🧪 저장된 메시지 읽기 테스트...")

	// 매퍼 생성
	mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")

	// 샘플 메시지 읽기
	samplesFile := "examples/cometbft/samples.json"
	messages, err := loadMessagesFromJSON(samplesFile)
	if err != nil {
		fmt.Printf("❌ 샘플 메시지 읽기 실패: %v\n", err)
		return
	}

	fmt.Printf("📖 %d개 샘플 메시지 로드 완료\n", len(messages))

	// 각 메시지 변환 테스트
	successCount := 0
	for i, msg := range messages {
		fmt.Printf("\n📦 메시지 %d: %s\n", i+1, msg.MessageType)

		canonical, err := mapper.ToCanonical(msg)
		if err != nil {
			fmt.Printf("   ❌ 변환 실패: %v\n", err)
			continue
		}

		successCount++
		fmt.Printf("   ✅ 변환 성공!\n")
		fmt.Printf("      📊 Height: %v\n", canonical.Height)
		fmt.Printf("      📊 Round: %v\n", canonical.Round)
		fmt.Printf("      📊 Type: %s\n", canonical.Type)
		if canonical.Proposer != "" {
			fmt.Printf("      📊 Proposer: %s\n", canonical.Proposer)
		}
		if canonical.Validator != "" {
			fmt.Printf("      📊 Validator: %s\n", canonical.Validator)
		}
		if canonical.BlockHash != "" {
			fmt.Printf("      📊 BlockHash: %s\n", canonical.BlockHash)
		}
		if len(canonical.Extensions) > 0 {
			fmt.Printf("      📊 Extensions: %d개\n", len(canonical.Extensions))
		}
	}

	fmt.Printf("\n📊 테스트 결과: %d/%d 성공\n", successCount, len(messages))
}

func loadMessagesFromJSON(filename string) ([]abstraction.RawConsensusMessage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []abstraction.RawConsensusMessage
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&messages)
	return messages, err
}

// 예제 메시지 생성 함수들
func createExampleNewRoundStep(height int64, round int32) abstraction.RawConsensusMessage {
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
		Metadata: map[string]interface{}{
			"source": "example_generator",
			"height": height,
			"round":  round,
		},
	}
}

func createExampleProposal(height int64, round int32) abstraction.RawConsensusMessage {
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
		"proposer_address": "cosmos1abc123def456ghi789",
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
		Metadata: map[string]interface{}{
			"source": "example_generator",
			"height": height,
			"round":  round,
		},
	}
}

func createExampleBlockPart(height int64, round int32, partIndex uint32) abstraction.RawConsensusMessage {
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
		Metadata: map[string]interface{}{
			"source": "example_generator",
			"height": height,
			"round":  round,
		},
	}
}

func createExampleVote(height int64, round int32, validator, voteType string) abstraction.RawConsensusMessage {
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
		Metadata: map[string]interface{}{
			"source": "example_generator",
			"height": height,
			"round":  round,
		},
	}
}

func createExampleNewValidBlock(height int64, round int32) abstraction.RawConsensusMessage {
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
		Metadata: map[string]interface{}{
			"source": "example_generator",
			"height": height,
			"round":  round,
		},
	}
}

func createExampleCommit(height int64, round int32) abstraction.RawConsensusMessage {
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
		Metadata: map[string]interface{}{
			"source": "example_generator",
			"height": height,
			"round":  round,
		},
	}
}
