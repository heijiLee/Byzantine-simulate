package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// DetailedConsensusState represents the detailed structure of CometBFT consensus state
type DetailedConsensusState struct {
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

// RunDetailedConsensusFormatter runs the detailed consensus state formatter
func RunDetailedConsensusFormatter() {
	fmt.Println("🔍 CometBFT Detailed Consensus State Formatter")
	fmt.Println("=============================================")

	// JSON 파일 읽기
	data, err := readDetailedConsensusStateJSON()
	if err != nil {
		fmt.Printf("❌ JSON 파일 읽기 실패: %v\n", err)
		return
	}

	// JSON 파싱
	var consensusState DetailedConsensusState
	if err := json.Unmarshal(data, &consensusState); err != nil {
		fmt.Printf("❌ JSON 파싱 실패: %v\n", err)
		return
	}

	fmt.Println("✅ Consensus State JSON 파일 읽기 및 파싱 완료")
	fmt.Println()

	// 상세 분석 및 포맷팅
	formatConsensusState(&consensusState)
}

func formatConsensusState(state *DetailedConsensusState) {
	rs := state.Result.RoundState

	// 기본 정보
	fmt.Println("📊 Consensus State Overview")
	fmt.Println("============================")
	fmt.Printf("   Height: %s\n", rs.Height)
	fmt.Printf("   Round: %d\n", rs.Round)
	fmt.Printf("   Step: %d\n", rs.Step)
	fmt.Printf("   Start Time: %s\n", rs.StartTime)
	fmt.Printf("   Commit Time: %s\n", rs.CommitTime)
	fmt.Printf("   Triggered Timeout Precommit: %t\n", rs.TriggeredTimeoutPrecommit)
	fmt.Println()

	// Validator 정보
	fmt.Println("👥 Current Validators")
	fmt.Println("====================")
	fmt.Printf("   Total Validators: %d\n", len(rs.Validators.Validators))
	fmt.Println()

	for i, validator := range rs.Validators.Validators {
		fmt.Printf("   [%d] Address: %s\n", i+1, validator.Address)
		fmt.Printf("        PubKey Type: %s\n", validator.PubKey.Type)
		fmt.Printf("        PubKey Value: %s\n", validator.PubKey.Value)
		fmt.Printf("        Voting Power: %s\n", validator.VotingPower)
		fmt.Printf("        Proposer Priority: %s\n", validator.ProposerPriority)
		fmt.Println()
	}

	// Current Proposer
	fmt.Println("🎯 Current Proposer")
	fmt.Println("==================")
	fmt.Printf("   Address: %s\n", rs.Validators.Proposer.Address)
	fmt.Printf("   PubKey Type: %s\n", rs.Validators.Proposer.PubKey.Type)
	fmt.Printf("   PubKey Value: %s\n", rs.Validators.Proposer.PubKey.Value)
	fmt.Printf("   Voting Power: %s\n", rs.Validators.Proposer.VotingPower)
	fmt.Printf("   Proposer Priority: %s\n", rs.Validators.Proposer.ProposerPriority)
	fmt.Println()

	// Vote 상태 분석
	fmt.Println("🗳️ Vote Status Analysis")
	fmt.Println("=======================")
	for _, vote := range rs.Votes {
		fmt.Printf("   Round %d:\n", vote.Round)
		fmt.Printf("     Prevotes: %s\n", vote.PrevotesBitArray)
		fmt.Printf("     Precommits: %s\n", vote.PrecommitsBitArray)
		fmt.Printf("     Prevote Details: %v\n", vote.Prevotes)
		fmt.Printf("     Precommit Details: %v\n", vote.Precommits)
		fmt.Println()
	}

	// Last Commit 분석
	fmt.Println("📝 Last Commit Analysis")
	fmt.Println("======================")
	fmt.Printf("   Votes Bit Array: %s\n", rs.LastCommit.VotesBitArray)
	fmt.Printf("   Total Votes: %d개\n", len(rs.LastCommit.Votes))
	fmt.Println()

	for i, vote := range rs.LastCommit.Votes {
		fmt.Printf("   Vote[%d]: %s\n", i, vote)
	}
	fmt.Println()

	// Last Validators
	fmt.Println("🔄 Last Validators (Previous Height)")
	fmt.Println("===================================")
	fmt.Printf("   Total Validators: %d\n", len(rs.LastValidators.Validators))
	fmt.Println()

	for i, validator := range rs.LastValidators.Validators {
		fmt.Printf("   [%d] Address: %s\n", i+1, validator.Address)
		fmt.Printf("        Voting Power: %s\n", validator.VotingPower)
		fmt.Printf("        Proposer Priority: %s\n", validator.ProposerPriority)
		fmt.Println()
	}

	// Last Proposer
	fmt.Println("🎯 Last Proposer (Previous Height)")
	fmt.Println("=================================")
	fmt.Printf("   Address: %s\n", rs.LastValidators.Proposer.Address)
	fmt.Printf("   Voting Power: %s\n", rs.LastValidators.Proposer.VotingPower)
	fmt.Printf("   Proposer Priority: %s\n", rs.LastValidators.Proposer.ProposerPriority)
	fmt.Println()

	// Peer 정보
	fmt.Println("🌐 Peer Network Status")
	fmt.Println("======================")
	fmt.Printf("   Total Peers: %d\n", len(state.Result.Peers))
	fmt.Println()

	for i, peer := range state.Result.Peers {
		fmt.Printf("   Peer[%d]: %s\n", i+1, peer.NodeAddress)
		fmt.Printf("     Height: %s\n", peer.PeerState.RoundState.Height)
		fmt.Printf("     Round: %d\n", peer.PeerState.RoundState.Round)
		fmt.Printf("     Step: %d\n", peer.PeerState.RoundState.Step)
		fmt.Printf("     Start Time: %s\n", peer.PeerState.RoundState.StartTime)
		fmt.Printf("     Proposal: %t\n", peer.PeerState.RoundState.Proposal)
		fmt.Printf("     Proposal Block Parts Total: %d\n", peer.PeerState.RoundState.ProposalBlockPartSetHeader.Total)
		fmt.Printf("     Proposal Block Parts Hash: %s\n", peer.PeerState.RoundState.ProposalBlockPartSetHeader.Hash)
		fmt.Printf("     Proposal Pol: %s\n", peer.PeerState.RoundState.ProposalPol)
		fmt.Printf("     Prevotes: %s\n", peer.PeerState.RoundState.Prevotes)
		fmt.Printf("     Precommits: %s\n", peer.PeerState.RoundState.Precommits)
		fmt.Printf("     Last Commit Round: %d\n", peer.PeerState.RoundState.LastCommitRound)
		fmt.Printf("     Last Commit: %s\n", peer.PeerState.RoundState.LastCommit)
		fmt.Printf("     Catchup Commit Round: %d\n", peer.PeerState.RoundState.CatchupCommitRound)
		fmt.Printf("     Catchup Commit: %s\n", peer.PeerState.RoundState.CatchupCommit)
		fmt.Printf("     Stats - Votes: %s\n", peer.PeerState.Stats.Votes)
		fmt.Printf("     Stats - Block Parts: %s\n", peer.PeerState.Stats.BlockParts)
		fmt.Println()
	}

	// 상태 요약
	fmt.Println("📈 Consensus State Summary")
	fmt.Println("=========================")
	fmt.Printf("   Current Height: %s\n", rs.Height)
	fmt.Printf("   Current Round: %d\n", rs.Round)
	fmt.Printf("   Current Step: %d\n", rs.Step)
	fmt.Printf("   Validators Count: %d\n", len(rs.Validators.Validators))
	fmt.Printf("   Peers Count: %d\n", len(state.Result.Peers))
	fmt.Printf("   Commit Round: %d\n", rs.CommitRound)
	fmt.Printf("   Locked Round: %d\n", rs.LockedRound)
	fmt.Printf("   Valid Round: %d\n", rs.ValidRound)
	fmt.Printf("   Proposal: %v\n", rs.Proposal != nil)
	fmt.Printf("   Proposal Block: %v\n", rs.ProposalBlock != nil)
	fmt.Printf("   Locked Block: %v\n", rs.LockedBlock != nil)
	fmt.Printf("   Valid Block: %v\n", rs.ValidBlock != nil)
	fmt.Println()

	// 시간 분석
	fmt.Println("⏰ Time Analysis")
	fmt.Println("===============")
	if startTime, err := time.Parse(time.RFC3339, rs.StartTime); err == nil {
		if commitTime, err := time.Parse(time.RFC3339, rs.CommitTime); err == nil {
			duration := startTime.Sub(commitTime)
			fmt.Printf("   Time since last commit: %v\n", duration)
			fmt.Printf("   Start Time: %s\n", startTime.Format("2006-01-02 15:04:05 MST"))
			fmt.Printf("   Commit Time: %s\n", commitTime.Format("2006-01-02 15:04:05 MST"))
		}
	}
	fmt.Println()
}

func readDetailedConsensusStateJSON() ([]byte, error) {
	data, err := os.ReadFile("examples/cometbft/curl.json")
	if err != nil {
		return nil, err
	}

	// JSON 데이터에서 주석 제거 (첫 번째 줄)
	lines := strings.Split(string(data), "\n")
	if len(lines) > 1 && strings.HasPrefix(lines[0], "//") {
		// 주석 줄 제거
		jsonData := strings.Join(lines[1:], "\n")
		// % 문자 제거 (curl 명령어의 프롬프트)
		jsonData = strings.TrimSuffix(jsonData, "%")
		jsonData = strings.TrimSpace(jsonData)
		return []byte(jsonData), nil
	}

	return data, nil
}
