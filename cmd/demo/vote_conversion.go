package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func RunVoteConversionTest() {
	fmt.Println("🧪 Vote 변환 테스트")
	fmt.Println("==================")

	// Vote.json 파일 읽기
	voteData, err := readVoteJSON()
	if err != nil {
		fmt.Printf("❌ Vote.json 읽기 실패: %v\n", err)
		return
	}
	fmt.Println("✅ Vote.json 파일 읽기 완료")

	// 각 Vote 예제에 대해 변환 테스트
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

		if testVoteConversionCase(voteData, tc.key, mapper) {
			successCount++
			fmt.Println("✅ 변환 성공!")
		} else {
			fmt.Println("❌ 변환 실패!")
		}
	}

	fmt.Printf("\n📊 전체 결과: %d/%d 성공 (%.1f%%)\n",
		successCount, len(testCases), float64(successCount)/float64(len(testCases))*100)

	if successCount == len(testCases) {
		fmt.Println("🎉 모든 Vote 변환 테스트 통과!")
	} else {
		fmt.Printf("⚠️  %d개 테스트 실패\n", len(testCases)-successCount)
	}
}

func readVoteJSON() (map[string]interface{}, error) {
	file, err := os.Open("examples/cometbft/Vote.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var voteData map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&voteData)
	return voteData, err
}

func testVoteConversionCase(voteData map[string]interface{}, key string, mapper *cometbftAdapter.CometBFTMapper) bool {
	// 1. Vote 데이터 추출
	vote, exists := voteData[key]
	if !exists {
		fmt.Printf("   ❌ Vote 데이터 없음: %s\n", key)
		return false
	}

	// 2. RawCometBFT 메시지 생성
	rawVote, err := createRawVoteFromData(vote)
	if err != nil {
		fmt.Printf("   ❌ Raw Vote 생성 실패: %v\n", err)
		return false
	}

	// 원본 RawCometBFT 메시지 출력
	fmt.Printf("   📋 원본 RawCometBFT 메시지:\n")
	printRawMessage(rawVote)

	// 3. RawCometBFT → Canonical 변환
	fmt.Println("   🔄 RawCometBFT → Canonical 변환 중...")
	canonical, err := mapper.ToCanonical(rawVote)
	if err != nil {
		fmt.Printf("   ❌ Canonical 변환 실패: %v\n", err)
		return false
	}

	// Canonical 메시지 출력
	fmt.Printf("   📋 Canonical 메시지:\n")
	printCanonicalMessage(canonical)

	// 4. Canonical → RawCometBFT 변환
	fmt.Println("   🔄 Canonical → RawCometBFT 변환 중...")
	rawConverted, err := mapper.FromCanonical(canonical)
	if err != nil {
		fmt.Printf("   ❌ RawCometBFT 변환 실패: %v\n", err)
		return false
	}

	// 변환된 RawCometBFT 메시지 출력
	fmt.Printf("   📋 변환된 RawCometBFT 메시지:\n")
	printRawMessage(*rawConverted)

	// 5. 결과 비교
	fmt.Println("   🔍 원본과 변환된 메시지 비교 중...")
	if compareVoteMessages(rawVote, *rawConverted) {
		printConversionSummary(canonical)
		return true
	}

	return false
}

func createRawVoteFromData(voteData interface{}) (abstraction.RawConsensusMessage, error) {
	// Vote 데이터를 JSON으로 변환
	jsonPayload, err := json.Marshal(voteData)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	// 원본 타임스탬프 추출
	var timestamp time.Time
	if voteMap, ok := voteData.(map[string]interface{}); ok {
		if timestampStr, exists := voteMap["timestamp"]; exists {
			if timestampStr, ok := timestampStr.(string); ok {
				if parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr); err == nil {
					timestamp = parsedTime
				}
			}
		}
	}

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   timestamp, // 원본 타임스탬프 사용
		Metadata: map[string]interface{}{
			"source": "vote_test",
		},
	}, nil
}

func compareVoteMessages(original, converted abstraction.RawConsensusMessage) bool {
	// 1. 기본 필드 비교
	if original.ChainType != converted.ChainType {
		fmt.Printf("   ❌ ChainType 불일치: %s != %s\n", original.ChainType, converted.ChainType)
		return false
	}
	if original.MessageType != converted.MessageType {
		fmt.Printf("   ❌ MessageType 불일치: %s != %s\n", original.MessageType, converted.MessageType)
		return false
	}

	// 2. Payload 비교
	var origPayload, convPayload map[string]interface{}
	if err := json.Unmarshal(original.Payload, &origPayload); err != nil {
		fmt.Printf("   ❌ 원본 Payload 파싱 실패: %v\n", err)
		return false
	}
	if err := json.Unmarshal(converted.Payload, &convPayload); err != nil {
		fmt.Printf("   ❌ 변환된 Payload 파싱 실패: %v\n", err)
		return false
	}

	// 3. 핵심 필드 비교
	keyFields := []string{"type", "height", "round", "validator_address", "signature"}
	for _, field := range keyFields {
		origVal := origPayload[field]
		convVal := convPayload[field]

		if fmt.Sprintf("%v", origVal) != fmt.Sprintf("%v", convVal) {
			fmt.Printf("   ❌ %s 불일치: %v != %v\n", field, origVal, convVal)
			return false
		}
	}

	// 4. BlockID 비교
	if !compareBlockID(origPayload["block_id"], convPayload["block_id"]) {
		return false
	}

	return true
}

func compareBlockID(orig, conv interface{}) bool {
	if orig == nil && conv == nil {
		return true
	}
	if orig == nil || conv == nil {
		// nil과 빈 문자열은 동일하게 처리
		if orig == nil && conv == "" {
			return true
		}
		if orig == "" && conv == nil {
			return true
		}
		fmt.Printf("   ❌ BlockID nil 불일치: %v != %v\n", orig, conv)
		return false
	}

	origMap, origOk := orig.(map[string]interface{})
	convMap, convOk := conv.(map[string]interface{})

	if !origOk || !convOk {
		fmt.Printf("   ❌ BlockID 타입 불일치: %T != %T\n", orig, conv)
		return false
	}

	// Hash 비교
	origHash := origMap["hash"]
	convHash := convMap["hash"]

	// nil과 빈 문자열을 동일하게 처리
	origHashStr := fmt.Sprintf("%v", origHash)
	convHashStr := fmt.Sprintf("%v", convHash)

	if origHashStr == "<nil>" {
		origHashStr = ""
	}
	if convHashStr == "<nil>" {
		convHashStr = ""
	}

	if origHashStr != convHashStr {
		fmt.Printf("   ❌ BlockID hash 불일치: '%s' != '%s'\n", origHashStr, convHashStr)
		return false
	}

	return true
}

func printConversionSummary(canonical *abstraction.CanonicalMessage) {
	fmt.Printf("   📊 변환 요약:\n")
	fmt.Printf("      Type: %s\n", canonical.Type)
	fmt.Printf("      Height: %v\n", canonical.Height)
	fmt.Printf("      Round: %v\n", canonical.Round)
	if canonical.BlockHash != "" {
		fmt.Printf("      BlockHash: %s...\n", canonical.BlockHash[:min(20, len(canonical.BlockHash))])
	}
	fmt.Printf("      Validator: %s\n", canonical.Validator)
	fmt.Printf("      Extensions: %d개\n", len(canonical.Extensions))
}
