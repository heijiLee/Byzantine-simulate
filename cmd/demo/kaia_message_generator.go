package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	kaiaAdapter "codec/kaia/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("💾 Kaia IBFT 메시지 예제 저장기")
	fmt.Println("===============================")

	// 예제 메시지들 생성
	messages := generateKaiaMessages()

	// JSON 파일로 저장
	saveKaiaMessagesToJSON(messages)

	// 저장된 메시지 읽기 테스트
	testSavedKaiaMessages()

	fmt.Println("\n🎉 Kaia 메시지 예제 저장 및 테스트 완료!")
}

func generateKaiaMessages() []abstraction.RawConsensusMessage {
	fmt.Println("\n📦 Kaia IBFT 메시지 생성 중...")

	var messages []abstraction.RawConsensusMessage
	height := int64(1000000)
	round := int32(0)

	// 1. Preprepare (IBFT 3-phase의 1단계)
	messages = append(messages, createKaiaPreprepare(height, round))

	// 2. Prepare (IBFT 3-phase의 2단계) - 여러 validator
	for i := 0; i < 4; i++ {
		messages = append(messages, createKaiaPrepare(height, round, fmt.Sprintf("validator%d", i)))
	}

	// 3. Commit (IBFT 3-phase의 3단계) - 여러 validator
	for i := 0; i < 4; i++ {
		messages = append(messages, createKaiaCommit(height, round, fmt.Sprintf("validator%d", i)))
	}

	// 4. RoundChange (라운드 변경)
	messages = append(messages, createKaiaRoundChange(height, round+1))

	fmt.Printf("✅ %d개 Kaia IBFT 메시지 생성 완료\n", len(messages))
	return messages
}

func saveKaiaMessagesToJSON(messages []abstraction.RawConsensusMessage) {
	fmt.Println("\n💾 JSON 파일로 저장 중...")

	// examples 디렉토리 생성
	examplesDir := "examples/kaia"
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

func testSavedKaiaMessages() {
	fmt.Println("\n🧪 저장된 Kaia 메시지 읽기 테스트...")

	// 매퍼 생성
	mapper := kaiaAdapter.NewKaiaMapper("kaia-mainnet")

	// 샘플 메시지 읽기
	samplesFile := "examples/kaia/samples.json"
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

// Kaia IBFT 메시지 생성 함수들
func createKaiaPreprepare(height int64, round int32) abstraction.RawConsensusMessage {
	// Kaia IBFT Preprepare 메시지 구조
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	parentHash := fmt.Sprintf("0x%x", time.Now().UnixNano()-1000)

	// View 구조 (Round, Sequence)
	view := map[string]interface{}{
		"round":    round,
		"sequence": height,
	}

	// Proposal 구조 (블록 헤더 기반)
	proposal := map[string]interface{}{
		"number":      height,
		"hash":        blockHash,
		"parent_hash": parentHash,
		"timestamp":   time.Now().Unix(),
		"gas_limit":   30000000,
		"gas_used":    15000000,
		"extra_data":  []byte("kaia-ibft-consensus"),
		"mix_hash":    fmt.Sprintf("0x%x", time.Now().UnixNano()),
		"nonce":       []byte{0, 0, 0, 0, 0, 0, 0, 0},
		"base_fee":    big.NewInt(25000000000).String(), // 25 Gwei
	}

	// Preprepare 페이로드 (RLP 구조 시뮬레이션)
	payload := map[string]interface{}{
		"message_type": "Preprepare",
		"view":         view,
		"proposal":     proposal,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	// RLP 인코딩 시뮬레이션 (실제로는 RLP로 인코딩됨)
	rlpData := fmt.Sprintf("preprepare_%d_%d_%s", height, round, blockHash)

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeKaia,
		ChainID:     "kaia-mainnet",
		MessageType: "Preprepare",
		Payload:     jsonPayload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"source":   "kaia_example_generator",
			"height":   height,
			"round":    round,
			"rlp_data": rlpData,
			"consensus_msg": map[string]interface{}{
				"prev_hash": parentHash,
				"payload":   rlpData,
			},
		},
	}
}

func createKaiaPrepare(height int64, round int32, validator string) abstraction.RawConsensusMessage {
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	parentHash := fmt.Sprintf("0x%x", time.Now().UnixNano()-1000)

	// View 구조
	view := map[string]interface{}{
		"round":    round,
		"sequence": height,
	}

	// Subject 구조 (Prepare/Commit/RoundChange 공통)
	subject := map[string]interface{}{
		"view":      view,
		"digest":    blockHash,  // 제안 블록 해시
		"prev_hash": parentHash, // 부모 블록 해시
	}

	payload := map[string]interface{}{
		"message_type": "Prepare",
		"subject":      subject,
		"validator":    validator,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	rlpData := fmt.Sprintf("prepare_%d_%d_%s_%s", height, round, validator, blockHash)

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeKaia,
		ChainID:     "kaia-mainnet",
		MessageType: "Prepare",
		Payload:     jsonPayload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"source":    "kaia_example_generator",
			"height":    height,
			"round":     round,
			"validator": validator,
			"rlp_data":  rlpData,
			"consensus_msg": map[string]interface{}{
				"prev_hash": parentHash,
				"payload":   rlpData,
			},
		},
	}
}

func createKaiaCommit(height int64, round int32, validator string) abstraction.RawConsensusMessage {
	blockHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	parentHash := fmt.Sprintf("0x%x", time.Now().UnixNano()-1000)

	// View 구조
	view := map[string]interface{}{
		"round":    round,
		"sequence": height,
	}

	// Subject 구조
	subject := map[string]interface{}{
		"view":      view,
		"digest":    blockHash,
		"prev_hash": parentHash,
	}

	// CommittedSeal (커밋 서명)
	committedSeal := fmt.Sprintf("committed_seal_%s_%d_%d", validator, height, round)

	payload := map[string]interface{}{
		"message_type":   "Commit",
		"subject":        subject,
		"validator":      validator,
		"committed_seal": committedSeal,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	rlpData := fmt.Sprintf("commit_%d_%d_%s_%s", height, round, validator, blockHash)

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeKaia,
		ChainID:     "kaia-mainnet",
		MessageType: "Commit",
		Payload:     jsonPayload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"source":         "kaia_example_generator",
			"height":         height,
			"round":          round,
			"validator":      validator,
			"rlp_data":       rlpData,
			"committed_seal": committedSeal,
			"consensus_msg": map[string]interface{}{
				"prev_hash": parentHash,
				"payload":   rlpData,
			},
		},
	}
}

func createKaiaRoundChange(height int64, round int32) abstraction.RawConsensusMessage {
	parentHash := fmt.Sprintf("0x%x", time.Now().UnixNano()-1000)

	// View 구조 (새 라운드)
	view := map[string]interface{}{
		"round":    round,
		"sequence": height,
	}

	// Subject 구조 (RoundChange용, Digest는 보통 비움)
	subject := map[string]interface{}{
		"view":      view,
		"digest":    "", // RoundChange에서는 보통 비움
		"prev_hash": parentHash,
	}

	payload := map[string]interface{}{
		"message_type": "RoundChange",
		"subject":      subject,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	rlpData := fmt.Sprintf("roundchange_%d_%d", height, round)

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeKaia,
		ChainID:     "kaia-mainnet",
		MessageType: "RoundChange",
		Payload:     jsonPayload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"source":   "kaia_example_generator",
			"height":   height,
			"round":    round,
			"rlp_data": rlpData,
			"consensus_msg": map[string]interface{}{
				"prev_hash": parentHash,
				"payload":   rlpData,
			},
		},
	}
}
