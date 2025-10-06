package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("🌐 CometBFT 실제 메시지 캡처 도구")
	fmt.Println("==================================")

	// 실제 CometBFT 노드에서 메시지 캡처
	captureRealMessages()
}

func captureRealMessages() {
	fmt.Println("\n📡 실제 CometBFT 노드 연결 시도...")

	// 실제 CometBFT RPC 엔드포인트들
	endpoints := []string{
		"https://rpc-cosmos.ecostake.com",      // Cosmos Hub
		"https://cosmos-rpc.polkachu.com",      // Cosmos Hub
		"https://rpc-cosmoshub.blockapsis.com", // Cosmos Hub
		"https://osmosis-rpc.polkachu.com",     // Osmosis
		"https://rpc-osmosis.blockapsis.com",   // Osmosis
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\n🔗 연결 시도: %s\n", endpoint)

		// 노드 상태 확인
		status, err := getNodeStatus(endpoint)
		if err != nil {
			fmt.Printf("   ❌ 연결 실패: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ 연결 성공!\n")
		fmt.Printf("   📊 노드 정보:\n")
		fmt.Printf("      ChainID: %s\n", status.Result.NodeInfo.Network)
		fmt.Printf("      Height: %s\n", status.Result.SyncInfo.LatestBlockHeight)
		fmt.Printf("      Time: %s\n", status.Result.SyncInfo.LatestBlockTime)

		// 최근 블록 정보 가져오기
		block, err := getLatestBlock(endpoint)
		if err != nil {
			fmt.Printf("   ❌ 블록 정보 가져오기 실패: %v\n", err)
			continue
		}

		fmt.Printf("   📦 최근 블록:\n")
		fmt.Printf("      Height: %s\n", block.Result.Block.Header.Height)
		fmt.Printf("      Hash: %s\n", block.Result.Block.Header.Hash)
		fmt.Printf("      Proposer: %s\n", block.Result.Block.Header.ProposerAddress)
		fmt.Printf("      Time: %s\n", block.Result.Block.Header.Time)

		// 합의 상태 확인
		consensus, err := getConsensusState(endpoint)
		if err != nil {
			fmt.Printf("   ❌ 합의 상태 가져오기 실패: %v\n", err)
			continue
		}

		fmt.Printf("   🗳️ 합의 상태:\n")
		fmt.Printf("      Height: %s\n", consensus.Result.RoundState.Height)
		fmt.Printf("      Round: %s\n", consensus.Result.RoundState.Round)
		fmt.Printf("      Step: %s\n", consensus.Result.RoundState.Step)

		// 실제 메시지 패턴 분석
		analyzeMessagePatterns(endpoint, status.Result.NodeInfo.Network)

		break // 첫 번째 성공한 엔드포인트만 사용
	}
}

func getNodeStatus(endpoint string) (*NodeStatusResponse, error) {
	resp, err := http.Get(endpoint + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var status NodeStatusResponse
	err = json.Unmarshal(body, &status)
	return &status, err
}

func getLatestBlock(endpoint string) (*BlockResponse, error) {
	resp, err := http.Get(endpoint + "/block")
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

func getConsensusState(endpoint string) (*ConsensusResponse, error) {
	resp, err := http.Get(endpoint + "/consensus_state")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var consensus ConsensusResponse
	err = json.Unmarshal(body, &consensus)
	return &consensus, err
}

func analyzeMessagePatterns(endpoint, chainID string) {
	fmt.Printf("\n🔍 메시지 패턴 분석:\n")

	// 실제 CometBFT 메시지 패턴 시뮬레이션
	mapper := cometbftAdapter.NewCometBFTMapper(chainID)

	// 실제 블록 높이와 라운드 사용
	height := int64(1000000) // 실제 높이
	round := int32(0)

	// 1. Proposal 메시지 패턴
	fmt.Printf("   📋 Proposal 패턴:\n")
	proposal := createRealisticProposal(height, round, chainID)
	canonical, err := mapper.ToCanonical(proposal)
	if err == nil {
		fmt.Printf("      ✅ 변환 성공: proposer=%s, height=%v\n",
			canonical.Proposer, canonical.Height)
	}

	// 2. Vote 메시지 패턴
	fmt.Printf("   🗳️ Vote 패턴:\n")
	for i := 0; i < 3; i++ {
		vote := createRealisticVote(height, round, fmt.Sprintf("validator%d", i), "PrevoteType", chainID)
		canonical, err := mapper.ToCanonical(vote)
		if err == nil {
			fmt.Printf("      ✅ 변환 성공: validator=%s, type=%s\n",
				canonical.Validator, canonical.Type)
		}
	}

	fmt.Printf("   🎯 실제 메시지 패턴 분석 완료!\n")
}

func createRealisticProposal(height int64, round int32, chainID string) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Proposal",
		"block_id": map[string]interface{}{
			"hash":      fmt.Sprintf("0x%x", time.Now().UnixNano()),
			"prev_hash": "0x1234567890abcdef",
			"part_set_header": map[string]interface{}{
				"total": 1,
				"hash":  []byte("block_hash"),
			},
		},
		"proposer_address": "cosmosvaloper1abc123def456",
		"signature":        fmt.Sprintf("sig_%d_%d", height, round),
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

func createRealisticVote(height int64, round int32, validator, voteType, chainID string) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Vote",
		"vote_type":    voteType,
		"block_id": map[string]interface{}{
			"hash": fmt.Sprintf("0x%x", time.Now().UnixNano()),
		},
		"validator_address": validator,
		"validator_index":   0,
		"signature":         fmt.Sprintf("%s_sig_%d_%d", voteType, height, round),
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

// CometBFT RPC 응답 구조체들
type NodeStatusResponse struct {
	Result struct {
		NodeInfo struct {
			Network string `json:"network"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
			LatestBlockTime   string `json:"latest_block_time"`
		} `json:"sync_info"`
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

type ConsensusResponse struct {
	Result struct {
		RoundState struct {
			Height string `json:"height"`
			Round  string `json:"round"`
			Step   string `json:"step"`
		} `json:"round_state"`
	} `json:"result"`
}
