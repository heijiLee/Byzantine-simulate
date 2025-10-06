package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("📖 CometBFT WAL 파일 직접 읽기 도구")
	fmt.Println("===================================")

	// WAL 파일 경로 찾기
	walPath := findWALFile()
	if walPath == "" {
		fmt.Println("❌ WAL 파일을 찾을 수 없습니다.")
		fmt.Println("💡 CometBFT 노드를 실행하거나 수동으로 경로를 지정해주세요.")
		return
	}

	fmt.Printf("📁 WAL 파일 경로: %s\n", walPath)

	// WAL 파일 읽기
	readWALFile(walPath)
}

func findWALFile() string {
	// 1. 환경변수에서 찾기
	if cmtHome := os.Getenv("CMTHOME"); cmtHome != "" {
		walPath := filepath.Join(cmtHome, "data", "cs.wal", "wal")
		if _, err := os.Stat(walPath); err == nil {
			return walPath
		}
	}

	// 2. 기본 경로들에서 찾기
	searchPaths := []string{
		"./cometbft-localnet/node0/data/cs.wal/wal",
		"./data/cs.wal/wal",
		"~/.cometbft/data/cs.wal/wal",
		"~/.gaia/data/cs.wal/wal",
		"~/.osmosis/data/cs.wal/wal",
	}

	for _, path := range searchPaths {
		expandedPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath
		}
	}

	return ""
}

func readWALFile(walPath string) {
	fmt.Println("\n📖 WAL 파일 읽기 시작...")

	// WAL 파일 열기
	file, err := os.Open(walPath)
	if err != nil {
		fmt.Printf("❌ WAL 파일 열기 실패: %v\n", err)
		return
	}
	defer file.Close()

	// 파일 정보
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("❌ 파일 정보 가져오기 실패: %v\n", err)
		return
	}

	fmt.Printf("📊 파일 크기: %d bytes\n", fileInfo.Size())
	fmt.Printf("📅 수정 시간: %s\n", fileInfo.ModTime())

	// WAL 파일은 바이너리 형식이므로 직접 읽기
	// 실제로는 CometBFT의 WAL 디코더가 필요하지만, 여기서는 파일 구조를 분석
	analyzeWALStructure(file)

	// 메시지 매퍼 생성
	mapper := cometbftAdapter.NewCometBFTMapper("test-chain")

	// 샘플 메시지 생성 및 테스트
	fmt.Println("\n🧪 샘플 메시지 변환 테스트...")
	testSampleMessages(mapper)
}

func analyzeWALStructure(file *os.File) {
	fmt.Println("\n🔍 WAL 파일 구조 분석...")

	// 파일의 처음 100바이트 읽기
	buffer := make([]byte, 100)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Printf("❌ 파일 읽기 실패: %v\n", err)
		return
	}

	fmt.Printf("📄 파일 헤더 (%d bytes):\n", n)
	fmt.Printf("   Hex: %x\n", buffer[:n])
	fmt.Printf("   Text: %s\n", string(buffer[:n]))

	// 파일 끝으로 이동하여 마지막 부분 확인
	file.Seek(-100, io.SeekEnd)
	n, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Printf("❌ 파일 끝 읽기 실패: %v\n", err)
		return
	}

	fmt.Printf("📄 파일 끝 (%d bytes):\n", n)
	fmt.Printf("   Hex: %x\n", buffer[:n])
	fmt.Printf("   Text: %s\n", string(buffer[:n]))
}

func testSampleMessages(mapper *cometbftAdapter.CometBFTMapper) {
	// 실제 CometBFT 메시지 패턴 생성
	messages := []abstraction.RawConsensusMessage{
		createSampleNewRoundStep(),
		createSampleProposal(),
		createSampleVote(),
		createSampleBlockPart(),
		createSampleNewValidBlock(),
	}

	successCount := 0
	for i, msg := range messages {
		fmt.Printf("\n📦 메시지 %d: %s\n", i+1, msg.MessageType)

		// Canonical로 변환
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

	fmt.Printf("\n📊 테스트 완료: %d/%d 성공\n", successCount, len(messages))
}

func createSampleNewRoundStep() abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":                   1,
		"round":                    0,
		"step":                     1,
		"seconds_since_start_time": 0,
		"last_commit_round":        -1,
		"message_type":             "NewRoundStep",
		"timestamp":                time.Now().Format(time.RFC3339),
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "test-chain",
		MessageType: "NewRoundStep",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createSampleProposal() abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       1,
		"round":        0,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Proposal",
		"block_id": map[string]interface{}{
			"hash":      "0x1234567890abcdef",
			"prev_hash": "0xabcdef1234567890",
			"part_set_header": map[string]interface{}{
				"total": 1,
				"hash":  []byte("0x1234567890abcdef"),
			},
		},
		"proposer_address": "validator0",
		"signature":        "sig_proposal_1",
		"pol_round":        -1,
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "test-chain",
		MessageType: "Proposal",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createSampleVote() abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       1,
		"round":        0,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Vote",
		"vote_type":    "PrevoteType",
		"block_id": map[string]interface{}{
			"hash": "0x1234567890abcdef",
		},
		"validator_address": "validator1",
		"validator_index":   1,
		"signature":         "sig_vote_1",
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "test-chain",
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createSampleBlockPart() abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       1,
		"round":        0,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "BlockPart",
		"block_id": map[string]interface{}{
			"hash": "0x1234567890abcdef",
		},
		"part_index": 0,
		"part_bytes": []byte("block_part_data"),
		"part_proof": []byte("merkle_proof"),
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "test-chain",
		MessageType: "BlockPart",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createSampleNewValidBlock() abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       1,
		"round":        0,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "NewValidBlock",
		"block_id": map[string]interface{}{
			"hash": "0x1234567890abcdef",
			"part_set_header": map[string]interface{}{
				"total": 1,
				"hash":  []byte("0x1234567890abcdef"),
			},
		},
		"is_commit":   true,
		"block_parts": []string{"part1", "part2"},
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "test-chain",
		MessageType: "NewValidBlock",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}
