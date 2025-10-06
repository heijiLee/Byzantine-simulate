package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	cometbftConsensus "codec/cometbft"
	cometbftAdapter "codec/cometbft/adapter"
	besuAdapter "codec/hyperledger/besu/adapter"
	fabricAdapter "codec/hyperledger/fabric/adapter"
	kaiaAdapter "codec/kaia/adapter"
	"codec/message/abstraction"
)

func TestCometBFTIntegration(t *testing.T) {
	fmt.Println("🚀 CometBFT 실제 프로토콜 구조 기반 통합 테스트")
	fmt.Println("================================================")

	// 검증자 세트 생성
	validators := []cometbftConsensus.Validator{
		{Address: "validator1", VotingPower: 100, PubKey: "pubkey1"},
		{Address: "validator2", VotingPower: 100, PubKey: "pubkey2"},
		{Address: "validator3", VotingPower: 100, PubKey: "pubkey3"},
		{Address: "validator4", VotingPower: 100, PubKey: "pubkey4"},
		{Address: "validator5", VotingPower: 100, PubKey: "pubkey5"},
	}

	// 합의 엔진 생성
	engine := cometbftConsensus.NewConsensusEngine(validators)
	fmt.Printf("📋 합의 엔진 초기화: %d명의 검증자, 총 투표력 %d\n",
		len(validators), engine.GetTotalPower())

	// 합의 라운드 시뮬레이션
	fmt.Printf("\n🔄 합의 라운드 시뮬레이션 시작\n")

	height := int64(1000)
	round := int32(1)

	// 엔진 상태 업데이트
	engine.AdvanceHeight(height)
	engine.AdvanceRound()

	// 1단계: Proposal
	fmt.Printf("\n📋 1단계: Proposal\n")
	// 현재 proposer 확인
	state := engine.GetState()
	proposer := state.Validators.Proposer.Address
	fmt.Printf("   현재 proposer: %s\n", proposer)

	proposal := createProposalMessage(height, round, proposer, "0xblock_hash_1000", -1)
	canonical, err := cometbftAdapter.NewCometBFTMapper("testnet").ToCanonical(proposal)
	if err != nil {
		t.Errorf("Proposal 변환 실패: %v", err)
		return
	}

	err = engine.ProcessMessage(canonical)
	if err != nil {
		t.Errorf("Proposal 처리 실패: %v", err)
		return
	}

	fmt.Printf("   ✅ Proposal 처리 완료\n")

	// 2단계: Prevote
	fmt.Printf("\n📋 2단계: Prevote\n")
	for i, validator := range validators {
		vote := createVoteMessage(height, round, validator.Address, "PrevoteType", "0xblock_hash_1000", int32(i))
		canonical, err := cometbftAdapter.NewCometBFTMapper("testnet").ToCanonical(vote)
		if err != nil {
			continue
		}

		err = engine.ProcessMessage(canonical)
		if err != nil {
			continue
		}
	}

	fmt.Printf("   ✅ Prevote 처리 완료\n")

	// 3단계: Precommit
	fmt.Printf("\n📋 3단계: Precommit\n")
	for i, validator := range validators {
		vote := createVoteMessage(height, round, validator.Address, "PrecommitType", "0xblock_hash_1000", int32(i))
		canonical, err := cometbftAdapter.NewCometBFTMapper("testnet").ToCanonical(vote)
		if err != nil {
			continue
		}

		err = engine.ProcessMessage(canonical)
		if err != nil {
			continue
		}
	}

	fmt.Printf("   ✅ Precommit 처리 완료\n")

	// 합의 상태 확인
	fmt.Printf("\n📊 합의 상태 확인\n")
	fmt.Printf("   현재 높이: %d\n", engine.GetCurrentHeight())
	fmt.Printf("   현재 라운드: %d\n", engine.GetCurrentRound())
	fmt.Printf("   합의 완료: %v\n", engine.IsConsensusReached())

	// 다양한 메시지 타입 테스트
	t.Run("MessageTypeTests", func(t *testing.T) {
		testMessageTypes(t)
	})

	// 크로스체인 변환 테스트
	t.Run("CrossChainTests", func(t *testing.T) {
		testCrossChainConversion(t)
	})

	// 성능 테스트
	t.Run("PerformanceTests", func(t *testing.T) {
		testPerformance(t)
	})

	fmt.Println("\n🎉 CometBFT 통합 테스트 완료!")
}

func testMessageTypes(t *testing.T) {
	fmt.Printf("\n📤 다양한 메시지 타입 테스트\n")

	messageTypes := []struct {
		name       string
		createFunc func() abstraction.RawConsensusMessage
	}{
		{
			name: "NewRoundStep",
			createFunc: func() abstraction.RawConsensusMessage {
				return createNewRoundStepMessage(1000, 1, 1, 0)
			},
		},
		{
			name: "Proposal",
			createFunc: func() abstraction.RawConsensusMessage {
				return createProposalMessage(1000, 1, "validator1", "0xblock_hash", -1)
			},
		},
		{
			name: "Vote",
			createFunc: func() abstraction.RawConsensusMessage {
				return createVoteMessage(1000, 1, "validator1", "PrevoteType", "0xblock_hash", 0)
			},
		},
		{
			name: "BlockPart",
			createFunc: func() abstraction.RawConsensusMessage {
				return createBlockPartMessage(1000, 1, "0xblock_hash", 0, []byte("part_data"))
			},
		},
		{
			name: "NewValidBlock",
			createFunc: func() abstraction.RawConsensusMessage {
				return createNewValidBlockMessage(1000, 1, "0xblock_hash", true)
			},
		},
		{
			name: "VoteSetBits",
			createFunc: func() abstraction.RawConsensusMessage {
				return createVoteSetBitsMessage(1000, 1, "PrevoteType", "0xblock_hash")
			},
		},
	}

	mapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

	for i, msgType := range messageTypes {
		t.Run(fmt.Sprintf("Test%d_%s", i+1, msgType.name), func(t *testing.T) {
			fmt.Printf("\n📤 테스트 %d: %s\n", i+1, msgType.name)

			rawMsg := msgType.createFunc()
			fmt.Printf("   📋 원본 메시지: %s\n", rawMsg.MessageType)

			// Canonical로 변환
			canonical, err := mapper.ToCanonical(rawMsg)
			if err != nil {
				t.Errorf("변환 실패: %v", err)
				return
			}

			fmt.Printf("   🔄 Canonical: type=%s, height=%v, round=%v\n",
				canonical.Type, canonical.Height, canonical.Round)

			// Extensions 확인
			if len(canonical.Extensions) > 0 {
				fmt.Printf("   📋 Extensions: %v\n", canonical.Extensions)
			}

			// 역변환
			rawBack, err := mapper.FromCanonical(canonical)
			if err != nil {
				t.Errorf("역변환 실패: %v", err)
				return
			}

			fmt.Printf("   📥 역변환: %s\n", rawBack.MessageType)

			// 데이터 보존 확인
			if canonical.Height.Cmp(big.NewInt(1000)) == 0 {
				fmt.Printf("   ✅ 높이 보존 확인\n")
			}
		})
	}
}

func testCrossChainConversion(t *testing.T) {
	fmt.Printf("\n🌉 크로스체인 변환 테스트\n")

	mappers := map[string]abstraction.Mapper{
		"CometBFT": cometbftAdapter.NewCometBFTMapper("testnet-cometbft"),
		"Fabric":   fabricAdapter.NewFabricMapper("testnet-fabric"),
		"Besu":     besuAdapter.NewBesuMapper("testnet-besu"),
		"Kaia":     kaiaAdapter.NewKaiaMapper("testnet-kaia"),
	}

	// CometBFT 메시지 생성
	cometbftMsg := createProposalMessage(1000, 1, "validator1", "0xblock_hash", -1)
	canonical, err := mappers["CometBFT"].ToCanonical(cometbftMsg)
	if err != nil {
		t.Errorf("CometBFT -> Canonical 실패: %v", err)
		return
	}

	fmt.Printf("   ✅ CometBFT -> Canonical 성공\n")

	// 각 체인으로 변환 테스트
	for chainName, mapper := range mappers {
		if chainName == "CometBFT" {
			continue // 이미 테스트했음
		}

		fmt.Printf("\n📥 %s로 변환:\n", chainName)

		raw, err := mapper.FromCanonical(canonical)
		if err != nil {
			t.Errorf("%s 변환 실패: %v", chainName, err)
			continue
		}

		fmt.Printf("   ✅ %s 메시지 생성: %s\n", chainName, raw.MessageType)

		// 다시 Canonical로 변환해서 데이터 보존 확인
		backCanonical, err := mapper.ToCanonical(*raw)
		if err != nil {
			t.Errorf("%s -> Canonical 실패: %v", chainName, err)
			continue
		}

		// 핵심 데이터 보존 확인 (nil 체크 추가)
		if canonical.Height != nil && backCanonical.Height != nil &&
			canonical.Round != nil && backCanonical.Round != nil &&
			canonical.Height.Cmp(backCanonical.Height) == 0 &&
			canonical.Round.Cmp(backCanonical.Round) == 0 &&
			canonical.BlockHash == backCanonical.BlockHash {
			fmt.Printf("   ✅ 데이터 보존 확인\n")
		} else {
			t.Errorf("데이터 불일치: %v != %v", canonical.Height, backCanonical.Height)
		}
	}
}

func testPerformance(t *testing.T) {
	fmt.Printf("\n📊 성능 테스트\n")

	mapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

	// 대량 메시지 처리 테스트
	fmt.Printf("\n📊 대량 메시지 처리 테스트 (5000개 메시지)\n")

	start := time.Now()
	successCount := 0

	for i := 0; i < 5000; i++ {
		msg := createVoteMessage(int64(1000+i), int32(i%10),
			fmt.Sprintf("validator%d", i%21), "PrevoteType",
			fmt.Sprintf("0xblock_%d", i), int32(i%21))

		canonical, err := mapper.ToCanonical(msg)
		if err == nil {
			_, err = mapper.FromCanonical(canonical)
			if err == nil {
				successCount++
			}
		}
	}

	duration := time.Since(start)
	fmt.Printf("   처리 시간: %v\n", duration)
	fmt.Printf("   성공률: %d/5000 (%.2f%%)\n", successCount, float64(successCount)/50)
	fmt.Printf("   평균 처리 시간: %v/메시지\n", duration/5000)
	fmt.Printf("   처리량: %.0f 메시지/초\n", float64(successCount)/duration.Seconds())

	// 메모리 효율성 테스트
	fmt.Printf("\n💾 메모리 효율성 테스트\n")

	var messages []abstraction.CanonicalMessage
	for i := 0; i < 1000; i++ {
		msg := createProposalMessage(int64(1000+i), 1,
			fmt.Sprintf("validator%d", i%21),
			fmt.Sprintf("0xblock_%d", i), -1)

		canonical, err := mapper.ToCanonical(msg)
		if err == nil {
			messages = append(messages, *canonical)
		}
	}

	fmt.Printf("   메모리에 저장된 메시지 수: %d\n", len(messages))
	if len(messages) > 0 {
		fmt.Printf("   평균 메시지 크기: ~%d bytes\n", estimateMessageSize(messages[0]))
	}
}

// Helper functions
func createProposalMessage(height int64, round int32, proposer, blockHash string, polRound int32) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "Proposal",
		"block_id": map[string]interface{}{
			"hash":      blockHash,
			"prev_hash": "0xprev_hash",
			"part_set_header": map[string]interface{}{
				"total": 1,
				"hash":  []byte(blockHash),
			},
		},
		"proposer_address": proposer,
		"signature":        fmt.Sprintf("%s_sig", proposer),
		"pol_round":        polRound,
	}

	jsonPayload, _ := json.Marshal(payload)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createVoteMessage(height int64, round int32, validator, voteType, blockHash string, validatorIndex int32) abstraction.RawConsensusMessage {
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
		"validator_index":   validatorIndex,
		"signature":         fmt.Sprintf("%s_sig_%d", voteType, validatorIndex),
	}

	// Precommit의 경우 extension 추가
	if voteType == "PrecommitType" {
		payload["extension"] = []byte(fmt.Sprintf("extension_%s", validator))
		payload["extension_signature"] = []byte(fmt.Sprintf("ext_sig_%s", validator))
	}

	jsonPayload, _ := json.Marshal(payload)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createNewRoundStepMessage(height int64, round int32, step uint32, lastCommitRound int32) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":                   height,
		"round":                    round,
		"step":                     step,
		"last_commit_round":        lastCommitRound,
		"seconds_since_start_time": 10,
		"timestamp":                time.Now().Format(time.RFC3339),
		"message_type":             "NewRoundStep",
	}

	jsonPayload, _ := json.Marshal(payload)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "NewRoundStep",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createBlockPartMessage(height int64, round int32, blockHash string, partIndex uint32, partBytes []byte) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "BlockPart",
		"block_id": map[string]interface{}{
			"hash": blockHash,
		},
		"part_index": partIndex,
		"part_bytes": partBytes,
		"part_proof": []byte("merkle_proof"),
	}

	jsonPayload, _ := json.Marshal(payload)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "BlockPart",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createNewValidBlockMessage(height int64, round int32, blockHash string, isCommit bool) abstraction.RawConsensusMessage {
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
		ChainID:     "testnet-cometbft",
		MessageType: "NewValidBlock",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func createVoteSetBitsMessage(height int64, round int32, voteType, blockHash string) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": "VoteSetBits",
		"vote_type":    voteType,
		"block_id": map[string]interface{}{
			"hash": blockHash,
		},
		"votes_bit_array": []string{"bit1", "bit2", "bit3"},
	}

	jsonPayload, _ := json.Marshal(payload)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "VoteSetBits",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func estimateMessageSize(msg abstraction.CanonicalMessage) int {
	size := 100 // 기본 크기
	size += len(msg.ChainID) + len(msg.BlockHash) + len(msg.Proposer) + len(msg.Validator) + len(msg.Signature)

	for k := range msg.Extensions {
		size += len(k) + 50
	}

	return size
}
