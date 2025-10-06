package main

import (
	"fmt"
	"testing"
	"time"

	"codec/abstraction"
	cometbftAdapter "codec/cometbft/adapter"
	besuAdapter "codec/hyperledger/besu/adapter"
	kaiaAdapter "codec/kaia/adapter"
)

func BenchmarkCometBFTConversion(b *testing.B) {
	mapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canonical, _ := mapper.ToCanonical(rawMsg)
		mapper.FromCanonical(canonical)
	}
}

func BenchmarkCrossChainConversion(b *testing.B) {
	cometbftMapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")
	besuMapper := besuAdapter.NewBesuMapper("testnet-besu")

	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canonical, _ := cometbftMapper.ToCanonical(rawMsg)
		besuRaw, _ := besuMapper.FromCanonical(canonical)
		besuMapper.ToCanonical(*besuRaw)
	}
}

func BenchmarkAllChainsConversion(b *testing.B) {
	mappers := map[string]abstraction.Mapper{
		"cometbft": cometbftAdapter.NewCometBFTMapper("testnet-cometbft"),
		"besu":     besuAdapter.NewBesuMapper("testnet-besu"),
		"kaia":     kaiaAdapter.NewKaiaMapper("testnet-kaia"),
	}

	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canonical, _ := mappers["cometbft"].ToCanonical(rawMsg)
		for _, mapper := range mappers {
			raw, _ := mapper.FromCanonical(canonical)
			mapper.ToCanonical(*raw)
		}
	}
}

func main() {
	fmt.Println("⚡ 성능 벤치마크 테스트")
	fmt.Println("======================")
	fmt.Println("다음 명령어로 벤치마크를 실행하세요:")
	fmt.Println()
	fmt.Println("go test -bench=. -benchmem")
	fmt.Println()
	fmt.Println("또는 개별 벤치마크:")
	fmt.Println("go test -bench=BenchmarkCometBFTConversion -benchmem")
	fmt.Println("go test -bench=BenchmarkCrossChainConversion -benchmem")
	fmt.Println("go test -bench=BenchmarkAllChainsConversion -benchmem")
}
