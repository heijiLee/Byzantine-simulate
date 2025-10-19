package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

// TestVoteConversionFromJSON tests conversion using actual Vote.json file
func TestVoteConversionFromJSON_DISABLED(t *testing.T) {
	fmt.Println("🧪 Vote.json → Canonical → Vote.json 변환 테스트")
	fmt.Println("===============================================")

	// 1. Vote.json 파일 읽기
	voteData, err := readVoteJSON()
	if err != nil {
		t.Fatalf("Vote.json 읽기 실패: %v", err)
	}
	fmt.Println("✅ Vote.json 파일 읽기 완료")

	// 2. 각 Vote 예제에 대해 변환 테스트
	mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")

	testCases := []struct {
		name string
		key  string
	}{
		{"Prevote for Block", "prevote_for_block"},
		{"Prevote Nil", "prevote_nil"},
		{"Precommit Basic", "precommit_basic"},
		{"Precommit with Extension", "precommit_with_extension"},
		{"Precommit Nil", "precommit_nil"},
		{"Prevote Round 1", "prevote_round_1"},
	}

	successCount := 0
	for i, tc := range testCases {
		fmt.Printf("\n📦 테스트 %d: %s\n", i+1, tc.name)
		fmt.Println("----------------------------------------")

		if testVoteConversion(t, voteData, tc.key, mapper) {
			successCount++
			fmt.Println("✅ 변환 성공!")
		} else {
			fmt.Println("❌ 변환 실패!")
		}
	}

	fmt.Printf("\n📊 전체 결과: %d/%d 성공 (%.1f%%)\n",
		successCount, len(testCases), float64(successCount)/float64(len(testCases))*100)

	if successCount != len(testCases) {
		t.Errorf("일부 테스트 실패: %d/%d", successCount, len(testCases))
	}
}

// TestSingleVoteConversion tests a single vote conversion in detail
func TestSingleVoteConversion_DISABLED(t *testing.T) {
	fmt.Println("🧪 단일 Vote 변환 테스트")
	fmt.Println("=======================")

	// Vote.json의 prevote_for_block 예제 직접 사용
	voteJSON := `{
		"type": 1,
		"height": "1000",
		"round": "0",
		"block_id": {
			"hash": "7B1C3F5E8D9A2E4F6C8B0A1D3E5F7A9B2C4D6E8F0A1B3C5D7E9F1A3B5C7D9E0F",
			"parts": {
				"total": 1,
				"hash": "A1B2C3D4E5F67890123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0"
			}
		},
		"timestamp": "2025-10-18T10:30:00.123456789Z",
		"validator_address": "95CEC8D3BCD896B97A9195BCC9FC3F5A7C65E092",
		"validator_index": 0,
		"signature": "3045022100E1F23456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABC0220DE67890ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF012"
	}`

	// 1. RawCometBFT 메시지 생성
	rawVote := createRawVoteFromJSON(voteJSON)
	fmt.Println("✅ RawCometBFT 메시지 생성 완료")

	// 2. Mapper 생성
	mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")
	fmt.Println("✅ CometBFT Mapper 생성 완료")

	// 3. RawCometBFT → Canonical 변환
	fmt.Println("\n🔄 RawCometBFT → Canonical 변환...")
	canonical, err := mapper.ToCanonical(rawVote)
	if err != nil {
		t.Fatalf("Canonical 변환 실패: %v", err)
	}
	fmt.Println("✅ Canonical 변환 성공!")

	// 4. Canonical → RawCometBFT 변환
	fmt.Println("\n🔄 Canonical → RawCometBFT 변환...")
	rawConverted, err := mapper.FromCanonical(canonical)
	if err != nil {
		t.Fatalf("RawCometBFT 변환 실패: %v", err)
	}
	fmt.Println("✅ RawCometBFT 역변환 성공!")

	// 5. 결과 비교
	fmt.Println("\n🔍 원본과 변환된 메시지 비교:")
	if !compareVoteMessages(t, rawVote, *rawConverted) {
		t.Fatal("변환 테스트 실패! 원본과 일치하지 않습니다.")
	}

	fmt.Println("🎉 변환 테스트 성공! 원본과 완전히 일치합니다.")
}

// TestVoteConversionBasic tests basic conversion functionality
func TestVoteConversionBasic_DISABLED(t *testing.T) {
	fmt.Println("🔍 Vote.json 변환 기본 검증")
	fmt.Println("=======================")

	// 1. Vote.json 파일 읽기
	voteData, err := readVoteJSON()
	if err != nil {
		t.Fatalf("Vote.json 읽기 실패: %v", err)
	}
	fmt.Println("✅ Vote.json 파일 읽기 완료")

	// 2. 첫 번째 Vote 예제 테스트 (prevote_for_block)
	vote, exists := voteData["prevote_for_block"]
	if !exists {
		t.Fatal("prevote_for_block 데이터 없음")
	}

	// 3. RawCometBFT 메시지 생성
	rawVote, err := createRawVote(vote)
	if err != nil {
		t.Fatalf("Raw Vote 생성 실패: %v", err)
	}
	fmt.Println("✅ RawCometBFT 메시지 생성 완료")

	// 4. 변환 테스트
	mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")

	fmt.Println("\n🔄 RawCometBFT → Canonical 변환...")
	canonical, err := mapper.ToCanonical(rawVote)
	if err != nil {
		t.Fatalf("변환 실패: %v", err)
	}
	fmt.Println("✅ Canonical 변환 성공!")

	// 5. 기본 검증
	expectedHeight := "1000"
	actualHeight := canonical.Height.String()
	if actualHeight != expectedHeight {
		t.Errorf("Height 불일치: expected %s, got %s", expectedHeight, actualHeight)
	}

	if canonical.Type != abstraction.MsgTypePrevote {
		t.Errorf("Type 불일치: expected %s, got %s", abstraction.MsgTypePrevote, canonical.Type)
	}

	expectedValidator := "95CEC8D3BCD896B97A9195BCC9FC3F5A7C65E092"
	if canonical.Validator != expectedValidator {
		t.Errorf("Validator 불일치: expected %s, got %s", expectedValidator, canonical.Validator)
	}

	fmt.Println("✅ 기본 검증 통과!")
}

// Helper functions

func readVoteJSON() (map[string]interface{}, error) {
	file, err := os.Open("../../examples/cometbft/Vote.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var voteData map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&voteData)
	return voteData, err
}

func testVoteConversion(t *testing.T, voteData map[string]interface{}, key string, mapper *cometbftAdapter.CometBFTMapper) bool {
	// 1. Vote 데이터 추출
	vote, exists := voteData[key]
	if !exists {
		t.Errorf("Vote 데이터 없음: %s", key)
		return false
	}

	// 2. RawCometBFT 메시지 생성
	rawVote, err := createRawVote(vote)
	if err != nil {
		t.Errorf("Raw Vote 생성 실패: %v", err)
		return false
	}

	// 3. RawCometBFT → Canonical 변환
	fmt.Println("   🔄 RawCometBFT → Canonical 변환 중...")
	canonical, err := mapper.ToCanonical(rawVote)
	if err != nil {
		t.Errorf("Canonical 변환 실패: %v", err)
		return false
	}

	// 4. Canonical → RawCometBFT 변환
	fmt.Println("   🔄 Canonical → RawCometBFT 변환 중...")
	rawConverted, err := mapper.FromCanonical(canonical)
	if err != nil {
		t.Errorf("RawCometBFT 변환 실패: %v", err)
		return false
	}

	// 5. 결과 비교
	fmt.Println("   🔍 원본과 변환된 메시지 비교 중...")
	if compareVoteMessages(t, rawVote, *rawConverted) {
		printConversionSummary(canonical)
		return true
	}

	return false
}

func createRawVoteFromJSON(voteJSON string) abstraction.RawConsensusMessage {
	jsonPayload := []byte(voteJSON)
	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Metadata: map[string]interface{}{
			"source": "single_vote_test",
		},
	}
}

func createRawVote(voteData interface{}) (abstraction.RawConsensusMessage, error) {
	// Vote 데이터를 JSON으로 변환
	jsonPayload, err := json.Marshal(voteData)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Metadata: map[string]interface{}{
			"source": "vote_test",
		},
	}, nil
}

func compareVoteMessages(t *testing.T, original, converted abstraction.RawConsensusMessage) bool {
	// 1. 기본 필드 비교
	if original.ChainType != converted.ChainType {
		t.Errorf("ChainType 불일치: %s != %s", original.ChainType, converted.ChainType)
		return false
	}
	if original.MessageType != converted.MessageType {
		t.Errorf("MessageType 불일치: %s != %s", original.MessageType, converted.MessageType)
		return false
	}

	// 2. Payload 비교
	var origPayload, convPayload map[string]interface{}
	if err := json.Unmarshal(original.Payload, &origPayload); err != nil {
		t.Errorf("원본 Payload 파싱 실패: %v", err)
		return false
	}
	if err := json.Unmarshal(converted.Payload, &convPayload); err != nil {
		t.Errorf("변환된 Payload 파싱 실패: %v", err)
		return false
	}

	// 3. 핵심 필드 비교
	keyFields := []string{"type", "height", "round", "validator_address", "signature"}
	for _, field := range keyFields {
		origVal := origPayload[field]
		convVal := convPayload[field]

		if fmt.Sprintf("%v", origVal) != fmt.Sprintf("%v", convVal) {
			t.Errorf("%s 불일치: %v != %v", field, origVal, convVal)
			return false
		}
	}

	// 4. BlockID 비교
	if !compareBlockID(t, origPayload["block_id"], convPayload["block_id"]) {
		return false
	}

	return true
}

func compareBlockID(t *testing.T, orig, conv interface{}) bool {
	if orig == nil && conv == nil {
		return true
	}
	if orig == nil || conv == nil {
		t.Error("BlockID nil 불일치")
		return false
	}

	origMap, origOk := orig.(map[string]interface{})
	convMap, convOk := conv.(map[string]interface{})

	if !origOk || !convOk {
		t.Error("BlockID 타입 불일치")
		return false
	}

	// Hash 비교 (빈 문자열과 nil을 동일하게 처리)
	origHash := origMap["hash"]
	convHash := convMap["hash"]

	// 빈 문자열과 nil을 동일하게 처리
	origHashStr := ""
	if origHash != nil {
		origHashStr = fmt.Sprintf("%v", origHash)
	}
	convHashStr := ""
	if convHash != nil {
		convHashStr = fmt.Sprintf("%v", convHash)
	}

	if origHashStr != convHashStr {
		t.Errorf("BlockID hash 불일치: %v != %v", origHash, convHash)
		return false
	}

	return true
}

func printConversionSummary(canonical *abstraction.CanonicalMessage) {
	fmt.Printf("   📊 변환 요약:\n")
	fmt.Printf("      Type: %s\n", canonical.Type)
	fmt.Printf("      Height: %v\n", canonical.Height)
	fmt.Printf("      Round: %v\n", canonical.Round)
	if len(canonical.BlockHash) > 20 {
		fmt.Printf("      BlockHash: %s...\n", canonical.BlockHash[:20])
	} else {
		fmt.Printf("      BlockHash: %s\n", canonical.BlockHash)
	}
	fmt.Printf("      Validator: %s\n", canonical.Validator)
	fmt.Printf("      Extensions: %d개\n", len(canonical.Extensions))
}
