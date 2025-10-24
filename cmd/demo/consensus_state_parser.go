package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

// ConsensusState represents the structure of CometBFT consensus state
type ConsensusState struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		RoundState struct {
			Height     string `json:"height"`
			Round      int    `json:"round"`
			Step       int    `json:"step"`
			StartTime  string `json:"start_time"`
			CommitTime string `json:"commit_time"`
			Validators struct {
				Validators []struct {
					Address string `json:"address"`
					PubKey  struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"pub_key"`
					VotingPower      string `json:"voting_power"`
					ProposerPriority string `json:"proposer_priority"`
				} `json:"validators"`
				Proposer struct {
					Address string `json:"address"`
					PubKey  struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"pub_key"`
					VotingPower      string `json:"voting_power"`
					ProposerPriority string `json:"proposer_priority"`
				} `json:"proposer"`
			} `json:"validators"`
			Proposal           interface{} `json:"proposal"`
			ProposalBlock      interface{} `json:"proposal_block"`
			ProposalBlockParts interface{} `json:"proposal_block_parts"`
			LockedRound        int         `json:"locked_round"`
			LockedBlock        interface{} `json:"locked_block"`
			LockedBlockParts   interface{} `json:"locked_block_parts"`
			ValidRound         int         `json:"valid_round"`
			ValidBlock         interface{} `json:"valid_block"`
			ValidBlockParts    interface{} `json:"valid_block_parts"`
			Votes              []struct {
				Round              int      `json:"round"`
				Prevotes           []string `json:"prevotes"`
				PrevotesBitArray   string   `json:"prevotes_bit_array"`
				Precommits         []string `json:"precommits"`
				PrecommitsBitArray string   `json:"precommits_bit_array"`
			} `json:"votes"`
			CommitRound int `json:"commit_round"`
			LastCommit  struct {
				Votes         []string `json:"votes"`
				VotesBitArray string   `json:"votes_bit_array"`
				PeerMaj23s    struct{} `json:"peer_maj_23s"`
			} `json:"last_commit"`
			LastValidators struct {
				Validators []struct {
					Address string `json:"address"`
					PubKey  struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"pub_key"`
					VotingPower      string `json:"voting_power"`
					ProposerPriority string `json:"proposer_priority"`
				} `json:"validators"`
				Proposer struct {
					Address string `json:"address"`
					PubKey  struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"pub_key"`
					VotingPower      string `json:"voting_power"`
					ProposerPriority string `json:"proposer_priority"`
				} `json:"proposer"`
			} `json:"last_validators"`
			TriggeredTimeoutPrecommit bool `json:"triggered_timeout_precommit"`
		} `json:"round_state"`
		Peers []struct {
			NodeAddress string `json:"node_address"`
			PeerState   struct {
				RoundState struct {
					Height                     string `json:"height"`
					Round                      int    `json:"round"`
					Step                       int    `json:"step"`
					StartTime                  string `json:"start_time"`
					Proposal                   bool   `json:"proposal"`
					ProposalBlockPartSetHeader struct {
						Total int    `json:"total"`
						Hash  string `json:"hash"`
					} `json:"proposal_block_part_set_header"`
					ProposalBlockParts interface{} `json:"proposal_block_parts"`
					ProposalPolRound   int         `json:"proposal_pol_round"`
					ProposalPol        string      `json:"proposal_pol"`
					Prevotes           string      `json:"prevotes"`
					Precommits         string      `json:"precommits"`
					LastCommitRound    int         `json:"last_commit_round"`
					LastCommit         string      `json:"last_commit"`
					CatchupCommitRound int         `json:"catchup_commit_round"`
					CatchupCommit      string      `json:"catchup_commit"`
				} `json:"round_state"`
				Stats struct {
					Votes      string `json:"votes"`
					BlockParts string `json:"block_parts"`
				} `json:"stats"`
			} `json:"peer_state"`
		} `json:"peers"`
	} `json:"result"`
}

func RunConsensusStateParser() {
	fmt.Println("🔍 CometBFT Consensus State 파서")
	fmt.Println("=================================")

	// curl.json 파일 읽기
	consensusData, err := readConsensusStateJSON()
	if err != nil {
		fmt.Printf("❌ Consensus State JSON 읽기 실패: %v\n", err)
		return
	}
	fmt.Println("✅ Consensus State JSON 파일 읽기 완료")

	// JSON 파싱
	var consensusState ConsensusState
	if err := json.Unmarshal(consensusData, &consensusState); err != nil {
		fmt.Printf("❌ JSON 파싱 실패: %v\n", err)
		return
	}

	// Consensus State 분석
	fmt.Println("\n📊 Consensus State 분석:")
	fmt.Printf("   Height: %s\n", consensusState.Result.RoundState.Height)
	fmt.Printf("   Round: %d\n", consensusState.Result.RoundState.Round)
	fmt.Printf("   Step: %d\n", consensusState.Result.RoundState.Step)
	fmt.Printf("   Start Time: %s\n", consensusState.Result.RoundState.StartTime)
	fmt.Printf("   Commit Time: %s\n", consensusState.Result.RoundState.CommitTime)
	fmt.Printf("   Validators: %d개\n", len(consensusState.Result.RoundState.Validators.Validators))
	fmt.Printf("   Peers: %d개\n", len(consensusState.Result.Peers))

	// Validator 정보 출력
	fmt.Println("\n👥 Validator 정보:")
	for i, validator := range consensusState.Result.RoundState.Validators.Validators {
		fmt.Printf("   [%d] Address: %s\n", i+1, validator.Address[:12]+"...")
		fmt.Printf("       Voting Power: %s\n", validator.VotingPower)
		fmt.Printf("       Proposer Priority: %s\n", validator.ProposerPriority)
	}

	// Proposer 정보
	fmt.Println("\n🎯 Current Proposer:")
	proposer := consensusState.Result.RoundState.Validators.Proposer
	fmt.Printf("   Address: %s\n", proposer.Address[:12]+"...")
	fmt.Printf("   Voting Power: %s\n", proposer.VotingPower)
	fmt.Printf("   Proposer Priority: %s\n", proposer.ProposerPriority)

	// Votes 분석
	fmt.Println("\n🗳️ Votes 분석:")
	for i, vote := range consensusState.Result.RoundState.Votes {
		fmt.Printf("   Round %d:\n", vote.Round)
		fmt.Printf("     Prevotes: %s\n", vote.PrevotesBitArray)
		fmt.Printf("     Precommits: %s\n", vote.PrecommitsBitArray)
		if i < len(vote.Prevotes) {
			fmt.Printf("     Prevote Details: %v\n", vote.Prevotes)
		}
		if i < len(vote.Precommits) {
			fmt.Printf("     Precommit Details: %v\n", vote.Precommits)
		}
	}

	// Last Commit 분석
	fmt.Println("\n📝 Last Commit 분석:")
	fmt.Printf("   Votes Bit Array: %s\n", consensusState.Result.RoundState.LastCommit.VotesBitArray)
	fmt.Printf("   Total Votes: %d개\n", len(consensusState.Result.RoundState.LastCommit.Votes))

	for i, vote := range consensusState.Result.RoundState.LastCommit.Votes {
		if i < 3 { // 처음 3개만 출력
			fmt.Printf("   Vote[%d]: %s\n", i, vote[:50]+"...")
		}
	}

	// 메시지 변환 테스트
	fmt.Println("\n🔄 메시지 변환 테스트:")
	mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")

	// Last Commit Votes를 RawConsensusMessage로 변환
	successCount := 0
	for i, voteStr := range consensusState.Result.RoundState.LastCommit.Votes {
		if i >= 2 { // 처음 2개만 테스트
			break
		}

		fmt.Printf("\n📦 Last Commit Vote %d 변환:\n", i+1)

		// Vote 문자열 파싱
		rawMsg, err := parseVoteString(voteStr, consensusState.Result.RoundState.Height)
		if err != nil {
			fmt.Printf("   ❌ Vote 파싱 실패: %v\n", err)
			continue
		}

		fmt.Printf("   📋 RawCometBFT 메시지:\n")
		printRawMessage(rawMsg)

		// Canonical로 변환
		fmt.Printf("   🔄 RawCometBFT → Canonical 변환 중...\n")
		canonical, err := mapper.ToCanonical(rawMsg)
		if err != nil {
			fmt.Printf("   ❌ Canonical 변환 실패: %v\n", err)
			continue
		}

		fmt.Printf("   📋 Canonical 메시지:\n")
		printCanonicalMessage(canonical)

		successCount++
		fmt.Printf("   ✅ 변환 성공!\n")
	}

	fmt.Printf("\n📊 변환 결과: %d/%d 성공\n", successCount, min(2, len(consensusState.Result.RoundState.LastCommit.Votes)))
}

func readConsensusStateJSON() ([]byte, error) {
	data, err := os.ReadFile("examples/cometbft/curl.json")
	if err != nil {
		return nil, err
	}

	// JSON 데이터에서 주석 제거 (첫 번째 줄)
	lines := strings.Split(string(data), "\n")
	if len(lines) > 1 && strings.HasPrefix(lines[0], "//") {
		// 주석 줄 제거
		jsonData := strings.Join(lines[1:], "\n")
		// % 문자와 공백 제거 (curl 명령어의 프롬프트)
		jsonData = strings.TrimSuffix(jsonData, "%")
		jsonData = strings.TrimSuffix(jsonData, " ")
		jsonData = strings.TrimSpace(jsonData)
		return []byte(jsonData), nil
	}

	return data, nil
}

func parseVoteString(voteStr, height string) (abstraction.RawConsensusMessage, error) {
	// Vote{0:20CA1B3031F4 162/00/SIGNED_MSG_TYPE_PRECOMMIT(Precommit) 5DC0096D27B5 D55807B92BE1 000000000000 @ 2025-10-19T07:45:15.586964Z}
	// 형식에서 정보 추출

	parts := strings.Split(voteStr, " ")
	if len(parts) < 8 {
		return abstraction.RawConsensusMessage{}, fmt.Errorf("invalid vote format")
	}

	// Validator index 추출
	validatorIndexStr := strings.TrimPrefix(parts[0], "Vote{")
	validatorIndexStr = strings.TrimSuffix(validatorIndexStr, ":")
	validatorIndex, err := strconv.Atoi(validatorIndexStr)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	// Validator address 추출
	validatorAddress := parts[1]

	// Height/Round 추출
	heightRound := parts[2] // "162/00"
	heightRoundParts := strings.Split(heightRound, "/")
	if len(heightRoundParts) != 2 {
		return abstraction.RawConsensusMessage{}, fmt.Errorf("invalid height/round format")
	}

	// Round 추출
	roundStr := heightRoundParts[1]
	round, err := strconv.Atoi(roundStr)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	// Message type 추출
	msgTypeStr := parts[3] // "SIGNED_MSG_TYPE_PRECOMMIT(Precommit)"
	var msgType string
	var typeNum int32
	if strings.Contains(msgTypeStr, "PRECOMMIT") {
		msgType = "precommit"
		typeNum = 2
	} else if strings.Contains(msgTypeStr, "PREVOTE") {
		msgType = "prevote"
		typeNum = 1
	} else {
		msgType = "vote"
		typeNum = 0
	}

	// Block hash 추출
	blockHash := parts[4]

	// Signature 추출
	signature := parts[5]

	// Timestamp 추출
	timestampStr := parts[7] // "@ 2025-10-19T07:45:15.586964Z"
	timestampStr = strings.TrimPrefix(timestampStr, "@ ")
	timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	// RawConsensusMessage 생성
	voteData := map[string]interface{}{
		"type":   typeNum,
		"height": height,
		"round":  fmt.Sprintf("%d", round),
		"block_id": map[string]interface{}{
			"hash": blockHash,
			"parts": map[string]interface{}{
				"total": 1,
				"hash":  "",
			},
		},
		"validator_address": validatorAddress,
		"validator_index":   validatorIndex,
		"signature":         signature,
		"timestamp":         timestamp.Format(time.RFC3339Nano),
	}

	jsonPayload, err := json.Marshal(voteData)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: msgType,
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   timestamp,
		Metadata: map[string]interface{}{
			"source":          "consensus_state",
			"validator_index": validatorIndex,
		},
	}, nil
}
