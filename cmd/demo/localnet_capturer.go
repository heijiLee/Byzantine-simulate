package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("🎯 CometBFT 로컬넷 메시지 캡처 도구")
	fmt.Println("====================================")

	// 로컬넷 노드들
	nodes := []string{
		"http://localhost:26657", // Node 0
		"http://localhost:26660", // Node 1
		"http://localhost:26663", // Node 2
		"http://localhost:26666", // Node 3
	}

	// 메시지 캡처 시작
	captureMessages(nodes)
}

func captureMessages(nodes []string) {
	fmt.Println("\n📡 로컬넷 노드 연결 확인...")

	// 활성 노드 찾기
	var activeNode string
	for _, node := range nodes {
		if isNodeActive(node) {
			activeNode = node
			fmt.Printf("✅ 활성 노드 발견: %s\n", node)
			break
		}
	}

	if activeNode == "" {
		fmt.Println("❌ 활성 노드를 찾을 수 없습니다.")
		fmt.Println("💡 로컬넷을 먼저 실행해주세요:")
		fmt.Println("   ./scripts/setup_localnet.sh")
		fmt.Println("   ./cometbft-localnet/node0/start.sh")
		return
	}

	// 노드 정보 가져오기
	nodeInfo, err := getNodeInfo(activeNode)
	if err != nil {
		fmt.Printf("❌ 노드 정보 가져오기 실패: %v\n", err)
		return
	}

	fmt.Printf("📊 노드 정보:\n")
	fmt.Printf("   ChainID: %s\n", nodeInfo.Result.NodeInfo.Network)
	fmt.Printf("   NodeID: %s\n", nodeInfo.Result.NodeInfo.ID)
	fmt.Printf("   Version: %s\n", nodeInfo.Result.NodeInfo.Version)

	// 실시간 메시지 모니터링 시작
	fmt.Println("\n🔄 실시간 메시지 모니터링 시작...")
	fmt.Println("   (Ctrl+C로 종료)")

	// 신호 처리
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 메시지 캡처 루프
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	mapper := cometbftAdapter.NewCometBFTMapper(nodeInfo.Result.NodeInfo.Network)
	messageCount := 0

	for {
		select {
		case <-c:
			fmt.Println("\n🛑 메시지 캡처 중단")
			return
		case <-ticker.C:
			// 블록 정보 가져오기
			block, err := getLatestBlock(activeNode)
			if err != nil {
				continue
			}

			// 새로운 블록이 있으면 메시지 생성
			if block.Result.Block.Header.Height != "" {
				height := block.Result.Block.Header.Height
				proposer := block.Result.Block.Header.ProposerAddress
				blockHash := block.Result.Block.Header.Hash

				fmt.Printf("\n📦 새 블록 발견: Height=%s, Proposer=%s\n", height, proposer)

				// 실제 메시지 패턴 생성
				messages := generateRealMessages(height, proposer, blockHash, nodeInfo.Result.NodeInfo.Network)

				// 각 메시지를 Canonical로 변환하여 테스트
				for _, msg := range messages {
					canonical, err := mapper.ToCanonical(msg)
					if err == nil {
						messageCount++
						fmt.Printf("   ✅ %s 변환 성공 (총 %d개)\n", msg.MessageType, messageCount)

						// 메시지 상세 정보 출력
						printMessageDetails(msg, canonical)
					}
				}
			}
		}
	}
}

func isNodeActive(nodeURL string) bool {
	resp, err := http.Get(nodeURL + "/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func getNodeInfo(nodeURL string) (*NodeInfoResponse, error) {
	resp, err := http.Get(nodeURL + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeInfo NodeInfoResponse
	err = json.Unmarshal(body, &nodeInfo)
	return &nodeInfo, err
}

func getLatestBlock(nodeURL string) (*BlockResponse, error) {
	resp, err := http.Get(nodeURL + "/block")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var block BlockResponse
	err = json.Unmarshal(body, &block)
	return &block, err
}

func generateRealMessages(height, proposer, blockHash, chainID string) []abstraction.RawConsensusMessage {
	var messages []abstraction.RawConsensusMessage

	// 1. NewRoundStep 메시지
	newRoundStep := createRealNewRoundStep(height, chainID)
	messages = append(messages, newRoundStep)

	// 2. Proposal 메시지
	proposal := createRealProposal(height, proposer, blockHash, chainID)
	messages = append(messages, proposal)

	// 3. BlockPart 메시지들
	for i := 0; i < 3; i++ {
		blockPart := createRealBlockPart(height, blockHash, uint32(i), chainID)
		messages = append(messages, blockPart)
	}

	// 4. Vote 메시지들 (Prevote)
	for i := 0; i < 4; i++ {
		vote := createRealVote(height, fmt.Sprintf("validator%d", i), "PrevoteType", blockHash, chainID)
		messages = append(messages, vote)
	}

	// 5. Vote 메시지들 (Precommit)
	for i := 0; i < 4; i++ {
		vote := createRealVote(height, fmt.Sprintf("validator%d", i), "PrecommitType", blockHash, chainID)
		messages = append(messages, vote)
	}

	// 6. NewValidBlock 메시지
	newValidBlock := createRealNewValidBlock(height, blockHash, chainID)
	messages = append(messages, newValidBlock)

	return messages
}

func createRealNewRoundStep(height, chainID string) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":                   height,
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
		ChainID:     chainID,
		MessageType: "NewRoundStep",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealProposal(height, proposer, blockHash, chainID string) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        0,
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
		"proposer_address": proposer,
		"signature":        fmt.Sprintf("sig_%s_%s", height, proposer),
		"pol_round":        -1,
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     chainID,
		MessageType: "Proposal",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealBlockPart(height, blockHash string, partIndex uint32, chainID string) abstraction.RawConsensusMessage {
	partData := []byte(fmt.Sprintf("block_part_%s_%d", height, partIndex))

	payload := map[string]interface{}{
		"height":       height,
		"round":        0,
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
		ChainID:     chainID,
		MessageType: "BlockPart",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealVote(height, validator, voteType, blockHash, chainID string) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        0,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Vote",
		"vote_type":    voteType,
		"block_id": map[string]interface{}{
			"hash": blockHash,
		},
		"validator_address": validator,
		"validator_index":   0,
		"signature":         fmt.Sprintf("%s_sig_%s_%s", voteType, height, validator),
	}

	// Precommit의 경우 extension 추가
	if voteType == "PrecommitType" {
		payload["extension"] = []byte(fmt.Sprintf("extension_%s", validator))
		payload["extension_signature"] = []byte(fmt.Sprintf("ext_sig_%s", validator))
	}

	jsonPayload, _ := json.Marshal(payload)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     chainID,
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createRealNewValidBlock(height, blockHash, chainID string) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        0,
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
		ChainID:     chainID,
		MessageType: "NewValidBlock",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func printMessageDetails(msg abstraction.RawConsensusMessage, canonical *abstraction.CanonicalMessage) {
	fmt.Printf("      📤 %s:\n", msg.MessageType)
	fmt.Printf("         Height: %v\n", canonical.Height)
	fmt.Printf("         Round: %v\n", canonical.Round)
	if canonical.Proposer != "" {
		fmt.Printf("         Proposer: %s\n", canonical.Proposer)
	}
	if canonical.Validator != "" {
		fmt.Printf("         Validator: %s\n", canonical.Validator)
	}
	if canonical.BlockHash != "" {
		fmt.Printf("         BlockHash: %s\n", canonical.BlockHash)
	}
	if len(canonical.Extensions) > 0 {
		fmt.Printf("         Extensions: %d개\n", len(canonical.Extensions))
	}
}

// RPC 응답 구조체들
type NodeInfoResponse struct {
	Result struct {
		NodeInfo struct {
			Network string `json:"network"`
			ID      string `json:"id"`
			Version string `json:"version"`
		} `json:"node_info"`
	} `json:"result"`
}

type BlockResponse struct {
	Result struct {
		Block struct {
			Header struct {
				Height          string `json:"height"`
				Hash            string `json:"hash"`
				ProposerAddress string `json:"proposer_address"`
				Time            string `json:"time"`
			} `json:"header"`
		} `json:"block"`
	} `json:"result"`
}
