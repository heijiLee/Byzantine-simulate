package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cometbft/cometbft/consensus"
	"google.golang.org/protobuf/encoding/protojson"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func main() {
	fmt.Println("📖 CometBFT WAL 파일 메시지 캡처 도구")
	fmt.Println("=====================================")

	// WAL 파일 경로 설정
	walPath := getWALPath()
	if walPath == "" {
		fmt.Println("❌ WAL 파일을 찾을 수 없습니다.")
		fmt.Println("💡 CometBFT 노드를 실행하거나 CMTHOME 환경변수를 설정해주세요.")
		return
	}

	fmt.Printf("📁 WAL 파일 경로: %s\n", walPath)

	// WAL 파일 읽기 및 메시지 캡처
	captureWALMessages(walPath)
}

func getWALPath() string {
	// 1. CMTHOME 환경변수에서 찾기
	if cmtHome := os.Getenv("CMTHOME"); cmtHome != "" {
		walPath := filepath.Join(cmtHome, "data", "cs.wal", "wal")
		if _, err := os.Stat(walPath); err == nil {
			return walPath
		}
	}

	// 2. 기본 경로들에서 찾기
	defaultPaths := []string{
		"./data/cs.wal/wal",
		"~/.cometbft/data/cs.wal/wal",
		"~/.gaia/data/cs.wal/wal",
		"~/.osmosis/data/cs.wal/wal",
		"./cometbft-localnet/node0/data/cs.wal/wal",
		"./cometbft-localnet/node1/data/cs.wal/wal",
		"./cometbft-localnet/node2/data/cs.wal/wal",
		"./cometbft-localnet/node3/data/cs.wal/wal",
	}

	for _, path := range defaultPaths {
		expandedPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath
		}
	}

	return ""
}

func captureWALMessages(walPath string) {
	fmt.Println("\n📖 WAL 파일 읽기 시작...")

	// WAL 파일 열기
	f, err := os.Open(walPath)
	if err != nil {
		fmt.Printf("❌ WAL 파일 열기 실패: %v\n", err)
		return
	}
	defer f.Close()

	// WAL 디코더 생성
	dec := consensus.NewWALDecoder(f)

	// 메시지 매퍼 생성 (체인 ID는 기본값 사용)
	mapper := cometbftAdapter.NewCometBFTMapper("test-chain")

	messageCount := 0
	successCount := 0

	fmt.Println("\n🔄 메시지 디코딩 중...")

	for {
		// WAL 메시지 디코딩
		tm, err := dec.Decode()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("⚠️  디코딩 오류: %v\n", err)
			continue
		}

		messageCount++
		fmt.Printf("\n📦 메시지 %d: %s\n", messageCount, getMessageType(tm.Msg))

		// Proto 메시지로 변환
		pm, err := consensus.WALToProto(tm.Msg)
		if err != nil {
			fmt.Printf("   ❌ Proto 변환 실패: %v\n", err)
			fmt.Printf("   📋 원본 메시지: %#v\n", tm.Msg)
			continue
		}

		// JSON으로 출력
		jsonData, err := protojson.MarshalOptions{
			Multiline: true,
			Indent:    "   ",
		}.Marshal(pm)
		if err != nil {
			fmt.Printf("   ❌ JSON 변환 실패: %v\n", err)
			continue
		}

		fmt.Printf("   📄 Proto JSON:\n%s\n", string(jsonData))

		// Byzantine Message Bridge로 변환 테스트
		rawMsg, err := convertProtoToRawMessage(pm, tm.Msg)
		if err != nil {
			fmt.Printf("   ❌ Raw 메시지 변환 실패: %v\n", err)
			continue
		}

		// Canonical로 변환
		canonical, err := mapper.ToCanonical(rawMsg)
		if err != nil {
			fmt.Printf("   ❌ Canonical 변환 실패: %v\n", err)
			continue
		}

		successCount++
		fmt.Printf("   ✅ Byzantine Message Bridge 변환 성공!\n")
		fmt.Printf("      📊 Height: %v\n", canonical.Height)
		fmt.Printf("      📊 Round: %v\n", canonical.Round)
		fmt.Printf("      📊 Type: %s\n", canonical.Type)
		if canonical.Proposer != "" {
			fmt.Printf("      📊 Proposer: %s\n", canonical.Proposer)
		}
		if canonical.Validator != "" {
			fmt.Printf("      📊 Validator: %s\n", canonical.Validator)
		}
		if canonical.BlockHash != "" {
			fmt.Printf("      📊 BlockHash: %s\n", canonical.BlockHash)
		}
		if len(canonical.Extensions) > 0 {
			fmt.Printf("      📊 Extensions: %d개\n", len(canonical.Extensions))
		}
	}

	fmt.Printf("\n📊 캡처 완료!\n")
	fmt.Printf("   총 메시지: %d개\n", messageCount)
	fmt.Printf("   성공 변환: %d개\n", successCount)
	fmt.Printf("   성공률: %.2f%%\n", float64(successCount)/float64(messageCount)*100)
}

func getMessageType(msg interface{}) string {
	switch msg.(type) {
	case *consensus.NewRoundStepMessage:
		return "NewRoundStep"
	case *consensus.NewValidBlockMessage:
		return "NewValidBlock"
	case *consensus.ProposalMessage:
		return "Proposal"
	case *consensus.ProposalPOLMessage:
		return "ProposalPOL"
	case *consensus.BlockPartMessage:
		return "BlockPart"
	case *consensus.VoteMessage:
		return "Vote"
	case *consensus.HasVoteMessage:
		return "HasVote"
	case *consensus.VoteSetMaj23Message:
		return "VoteSetMaj23"
	case *consensus.VoteSetBitsMessage:
		return "VoteSetBits"
	case *consensus.CommitMessage:
		return "Commit"
	case *consensus.ExtendedCommitMessage:
		return "ExtendedCommit"
	default:
		return fmt.Sprintf("%T", msg)
	}
}

func convertProtoToRawMessage(pm interface{}, originalMsg interface{}) (abstraction.RawConsensusMessage, error) {
	// Proto 메시지를 JSON으로 변환
	jsonData, err := protojson.MarshalOptions{
		Multiline: false,
		Indent:    "",
	}.Marshal(pm)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	// 메시지 타입 결정
	messageType := getMessageType(originalMsg)

	// RawConsensusMessage 생성
	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "test-chain", // 실제 체인 ID로 변경 가능
		MessageType: messageType,
		Payload:     jsonData,
		Encoding:    "proto",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"source": "wal",
			"type":   fmt.Sprintf("%T", originalMsg),
		},
	}

	return rawMsg, nil
}
