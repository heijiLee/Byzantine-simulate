package test

import (
	"testing"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

// TestVoteTypeParsing tests Vote type parsing specifically
func TestVoteTypeParsing(t *testing.T) {
	t.Log("Vote 타입 파싱 테스트")

	// 간단한 Prevote JSON
	prevoteJSON := `{
		"type": 1,
		"height": "1000",
		"round": "0",
		"block_id": {
			"hash": "test-hash"
		},
		"validator_address": "test-validator",
		"signature": "test-signature"
	}`

	// RawCometBFT 메시지 생성
	rawVote := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "test-chain",
		MessageType: "Vote",
		Payload:     []byte(prevoteJSON),
		Encoding:    "json",
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	// Mapper 생성
	mapper := cometbftAdapter.NewCometBFTMapper("test-chain")

	// RawCometBFT → Canonical 변환
	canonical, err := mapper.ToCanonical(rawVote)
	if err != nil {
		t.Fatalf("Canonical 변환 실패: %v", err)
	}

	// 타입 검증
	if canonical.Type != abstraction.MsgTypePrevote {
		t.Errorf("Prevote Type 불일치: expected '%s', got '%s'", abstraction.MsgTypePrevote, canonical.Type)
	}

	t.Logf("✅ Prevote 타입 파싱 성공: %s", canonical.Type)
}


