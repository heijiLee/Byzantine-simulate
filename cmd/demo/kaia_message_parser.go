package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kaiaAdapter "codec/kaia/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("📖 Kaia IBFT 메시지 파일 파서")
	fmt.Println("============================")

	// 명령행 인수 확인
	if len(os.Args) < 2 {
		printKaiaUsage()
		return
	}

	messageFile := os.Args[1]
	fmt.Printf("📁 메시지 파일: %s\n", messageFile)

	// 파일 존재 확인
	if _, err := os.Stat(messageFile); os.IsNotExist(err) {
		fmt.Printf("❌ 파일이 존재하지 않습니다: %s\n", messageFile)
		return
	}

	// 메시지 파일 파싱
	parseKaiaMessageFile(messageFile)
}

func printKaiaUsage() {
	fmt.Println("사용법:")
	fmt.Println("  go run cmd/demo/kaia_message_parser.go <메시지파일경로>")
	fmt.Println("")
	fmt.Println("예제:")
	fmt.Println("  go run cmd/demo/kaia_message_parser.go examples/kaia/samples.json")
	fmt.Println("  go run cmd/demo/kaia_message_parser.go examples/kaia/all_messages.json")
	fmt.Println("  go run cmd/demo/kaia_message_parser.go examples/kaia/Preprepare.json")
	fmt.Println("")
	fmt.Println("사용 가능한 파일들:")
	printKaiaAvailableFiles()
}

func printKaiaAvailableFiles() {
	examplesDir := "examples/kaia"
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		fmt.Println("  (examples/kaia 디렉토리가 없습니다. 먼저 예제를 생성해주세요)")
		return
	}

	files, err := filepath.Glob(examplesDir + "/*.json")
	if err != nil {
		fmt.Println("  (파일 목록을 가져올 수 없습니다)")
		return
	}

	for _, file := range files {
		fmt.Printf("  %s\n", file)
	}
}

func parseKaiaMessageFile(messageFile string) {
	fmt.Println("\n📖 Kaia IBFT 메시지 파일 파싱 중...")

	// 메시지 로드
	messages, err := loadKaiaMessagesFromJSON(messageFile)
	if err != nil {
		fmt.Printf("❌ 메시지 로드 실패: %v\n", err)
		return
	}

	fmt.Printf("✅ %d개 메시지 로드 완료\n", len(messages))

	// 메시지 타입별 통계
	messageStats := make(map[string]int)
	for _, msg := range messages {
		messageStats[msg.MessageType]++
	}

	fmt.Println("\n📊 Kaia IBFT 메시지 타입별 통계:")
	for msgType, count := range messageStats {
		fmt.Printf("   %s: %d개\n", msgType, count)
	}

	// Kaia 매퍼 생성
	mapper := kaiaAdapter.NewKaiaMapper("kaia-mainnet")

	// 각 메시지 변환 테스트
	fmt.Println("\n🔄 Kaia IBFT 메시지 변환 테스트...")
	successCount := 0
	errorCount := 0

	for i, msg := range messages {
		fmt.Printf("\n📦 메시지 %d/%d: %s\n", i+1, len(messages), msg.MessageType)

		// 메시지 상세 정보 출력
		printKaiaMessageDetails(msg)

		// Canonical로 변환
		canonical, err := mapper.ToCanonical(msg)
		if err != nil {
			fmt.Printf("   ❌ 변환 실패: %v\n", err)
			errorCount++
			continue
		}

		successCount++
		fmt.Printf("   ✅ 변환 성공!\n")
		printKaiaCanonicalDetails(canonical)
	}

	// 결과 요약
	fmt.Printf("\n📊 Kaia IBFT 변환 결과 요약:\n")
	fmt.Printf("   총 메시지: %d개\n", len(messages))
	fmt.Printf("   성공: %d개\n", successCount)
	fmt.Printf("   실패: %d개\n", errorCount)
	fmt.Printf("   성공률: %.2f%%\n", float64(successCount)/float64(len(messages))*100)

	// 변환 성공 메시지 출력
	if successCount > 0 {
		fmt.Printf("\n🎉 %d개 Kaia IBFT 메시지가 성공적으로 Canonical 형식으로 변환되었습니다!\n", successCount)
	}
}

func loadKaiaMessagesFromJSON(filename string) ([]abstraction.RawConsensusMessage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []abstraction.RawConsensusMessage
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&messages)
	return messages, err
}

func printKaiaMessageDetails(msg abstraction.RawConsensusMessage) {
	fmt.Printf("   📋 Kaia IBFT 원본 메시지 정보:\n")
	fmt.Printf("      ChainType: %s\n", msg.ChainType)
	fmt.Printf("      ChainID: %s\n", msg.ChainID)
	fmt.Printf("      MessageType: %s\n", msg.MessageType)
	fmt.Printf("      Encoding: %s\n", msg.Encoding)
	fmt.Printf("      Timestamp: %s\n", msg.Timestamp.Format("2006-01-02 15:04:05"))

	if len(msg.Metadata) > 0 {
		fmt.Printf("      Metadata: %d개 항목\n", len(msg.Metadata))
		for k, v := range msg.Metadata {
			fmt.Printf("         %s: %v\n", k, v)
		}
	}

	// Payload 일부 출력 (너무 길면 생략)
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &payload); err == nil {
		fmt.Printf("      Payload 키: %s\n", strings.Join(getMapKeys(payload), ", "))
	}
}

func printKaiaCanonicalDetails(canonical *abstraction.CanonicalMessage) {
	fmt.Printf("   📋 Canonical 메시지 정보:\n")
	fmt.Printf("      Height: %v\n", canonical.Height)
	fmt.Printf("      Round: %v\n", canonical.Round)
	fmt.Printf("      Type: %s\n", canonical.Type)

	if canonical.Proposer != "" {
		fmt.Printf("      Proposer: %s\n", canonical.Proposer)
	}
	if canonical.Validator != "" {
		fmt.Printf("      Validator: %s\n", canonical.Validator)
	}
	if canonical.BlockHash != "" {
		fmt.Printf("      BlockHash: %s\n", canonical.BlockHash)
	}
	if canonical.PrevHash != "" {
		fmt.Printf("      PrevHash: %s\n", canonical.PrevHash)
	}
	if canonical.Signature != "" {
		fmt.Printf("      Signature: %s\n", canonical.Signature)
	}
	if !canonical.Timestamp.IsZero() {
		fmt.Printf("      Timestamp: %s\n", canonical.Timestamp.Format("2006-01-02 15:04:05"))
	}
	if len(canonical.Extensions) > 0 {
		fmt.Printf("      Extensions: %d개\n", len(canonical.Extensions))
		for k, v := range canonical.Extensions {
			fmt.Printf("         %s: %v\n", k, v)
		}
	}
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
