package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	fabricAdapter "codec/hyperledger/fabric/adapter"
	"codec/message/abstraction"
)

func TestCometBFTAdvancedMapper(t *testing.T) {
	fmt.Println("🔬 CometBFT 고급 실험 - 실제 프로토콜 구조 기반")
	fmt.Println("================================================")

	// 다양한 CometBFT 메시지 타입 테스트
	messageTypes := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "NewRoundStep",
			payload: map[string]interface{}{
				"height":                   1000,
				"round":                    1,
				"step":                     1,
				"seconds_since_start_time": 10,
				"last_commit_round":        0,
				"message_type":             "NewRoundStep",
				"timestamp":                time.Now().Format(time.RFC3339),
			},
		},
		{
			name: "Proposal",
			payload: map[string]interface{}{
				"height":    1000,
				"round":     1,
				"timestamp": time.Now().Format(time.RFC3339),
				"block_id": map[string]interface{}{
					"hash":      "0xabc123def456",
					"prev_hash": "0x789abc123def",
					"part_set_header": map[string]interface{}{
						"total": 1,
						"hash":  []byte("part_hash"),
					},
				},
				"proposer_address": "validator1",
				"signature":        "sig_proposal_123",
				"pol_round":        -1,
				"message_type":     "Proposal",
			},
		},
		{
			name: "Vote (Prevote)",
			payload: map[string]interface{}{
				"height":    1000,
				"round":     1,
				"timestamp": time.Now().Format(time.RFC3339),
				"vote_type": "PrevoteType",
				"block_id": map[string]interface{}{
					"hash": "0xabc123def456",
				},
				"validator_address": "validator2",
				"validator_index":   1,
				"signature":         "sig_prevote_456",
				"message_type":      "Vote",
			},
		},
		{
			name: "BlockPart",
			payload: map[string]interface{}{
				"height":    1000,
				"round":     1,
				"timestamp": time.Now().Format(time.RFC3339),
				"block_id": map[string]interface{}{
					"hash": "0xabc123def456",
				},
				"part_index":   0,
				"part_bytes":   []byte("block_part_data"),
				"part_proof":   []byte("merkle_proof"),
				"message_type": "BlockPart",
			},
		},
	}

	mapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

	for i, msgType := range messageTypes {
		t.Run(fmt.Sprintf("Test%d_%s", i+1, msgType.name), func(t *testing.T) {
			fmt.Printf("\n📤 테스트 %d: %s\n", i+1, msgType.name)

			// JSON 페이로드 생성
			payload, _ := json.Marshal(msgType.payload)

			// Raw 메시지 생성
			rawMsg := abstraction.RawConsensusMessage{
				ChainType:   abstraction.ChainTypeCometBFT,
				ChainID:     "testnet-cometbft",
				MessageType: msgType.name,
				Payload:     payload,
				Encoding:    "json",
				Timestamp:   time.Now(),
			}

			// Canonical로 변환
			canonical, err := mapper.ToCanonical(rawMsg)
			if err != nil {
				t.Errorf("변환 실패: %v", err)
				return
			}

			fmt.Printf("   ✅ %s 변환 성공 (높이: %v)\n", rawMsg.MessageType, canonical.Height)

			// 역변환 테스트
			rawBack, err := mapper.FromCanonical(canonical)
			if err != nil {
				t.Errorf("역변환 실패: %v", err)
				return
			}

			fmt.Printf("   ✅ 역변환 성공: %s\n", rawBack.MessageType)
		})
	}

	// 크로스체인 변환 테스트
	t.Run("CrossChainConversion", func(t *testing.T) {
		fmt.Printf("\n📤 CometBFT -> Fabric 변환 테스트\n")

		cometbftMapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")
		fabricMapper := fabricAdapter.NewFabricMapper("testnet-fabric")

		cometbftProposal := createCometBFTMessage("Proposal", 1000, 1, map[string]interface{}{
			"block_id": map[string]interface{}{
				"hash":      "0xcometbft_block_hash",
				"prev_hash": "0xprev_hash",
			},
			"proposer_address": "cometbft_validator1",
			"signature":        "cometbft_sig_123",
		})

		// CometBFT -> Canonical
		canonical, err := cometbftMapper.ToCanonical(cometbftProposal)
		if err != nil {
			t.Errorf("CometBFT -> Canonical 실패: %v", err)
			return
		}

		fmt.Printf("   ✅ CometBFT -> Canonical 성공\n")

		// Canonical -> Fabric
		fabricRaw, err := fabricMapper.FromCanonical(canonical)
		if err != nil {
			t.Errorf("Canonical -> Fabric 실패: %v", err)
			return
		}

		fmt.Printf("   ✅ Canonical -> Fabric 성공: %s\n", fabricRaw.MessageType)

		// 데이터 보존 확인
		fabricCanonical, err := fabricMapper.ToCanonical(*fabricRaw)
		if err != nil {
			t.Errorf("Fabric -> Canonical 실패: %v", err)
			return
		}

		if canonical.Height.Cmp(fabricCanonical.Height) == 0 {
			fmt.Printf("   ✅ 높이 보존 확인: %v\n", canonical.Height)
		} else {
			t.Errorf("높이 불일치: %v != %v", canonical.Height, fabricCanonical.Height)
		}
	})

	// 성능 테스트
	t.Run("PerformanceTest", func(t *testing.T) {
		mapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

		// 대량 메시지 처리 테스트
		fmt.Printf("\n📊 대량 메시지 처리 테스트 (1000개 메시지)\n")

		start := time.Now()
		successCount := 0

		for i := 0; i < 1000; i++ {
			msg := createCometBFTMessage("Vote", int64(1000+i), int32(i%10), map[string]interface{}{
				"vote_type":         "PrevoteType",
				"validator_address": fmt.Sprintf("validator%d", i%21),
				"validator_index":   int32(i % 21),
				"signature":         fmt.Sprintf("sig_%d", i),
			})

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
		fmt.Printf("   성공률: %d/1000 (%.2f%%)\n", successCount, float64(successCount)/10)
		fmt.Printf("   평균 처리 시간: %v/메시지\n", duration/1000)

		// 메모리 사용량 시뮬레이션
		fmt.Printf("\n💾 메모리 사용량 시뮬레이션\n")

		var messages []abstraction.CanonicalMessage
		for i := 0; i < 100; i++ {
			msg := createCometBFTMessage("Proposal", int64(1000+i), 1, map[string]interface{}{
				"proposer_address": fmt.Sprintf("validator%d", i%21),
				"signature":        fmt.Sprintf("large_sig_%d", i),
			})

			canonical, err := mapper.ToCanonical(msg)
			if err == nil {
				messages = append(messages, *canonical)
			}
		}

		fmt.Printf("   메모리에 저장된 메시지 수: %d\n", len(messages))
		if len(messages) > 0 {
			fmt.Printf("   평균 메시지 크기: ~%d bytes\n", estimateMessageSize(messages[0]))
		}
	})
}

func createCometBFTMessage(msgType string, height int64, round int32, extraFields map[string]interface{}) abstraction.RawConsensusMessage {
	payload := map[string]interface{}{
		"height":       height,
		"round":        round,
		"timestamp":    time.Now().Format(time.RFC3339),
		"message_type": msgType,
	}

	// 추가 필드 병합
	for k, v := range extraFields {
		payload[k] = v
	}

	jsonPayload, _ := json.Marshal(payload)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: msgType,
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}

func estimateMessageSize(msg abstraction.CanonicalMessage) int {
	size := 100 // 기본 크기
	size += len(msg.ChainID) + len(msg.BlockHash) + len(msg.Proposer) + len(msg.Validator) + len(msg.Signature)

	// Extensions 크기 추정
	for k := range msg.Extensions {
		size += len(k) + 50 // 대략적인 값 크기
	}

	return size
}
