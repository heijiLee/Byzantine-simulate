package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

// CometBFTMessageSimulator generates synthetic CometBFT messages and shows how
// they travel through the canonical mapper.
type CometBFTMessageSimulator struct {
	mapper *cometbftAdapter.CometBFTMapper
	height int64
	round  int64
}

func NewCometBFTMessageSimulator(mapper *cometbftAdapter.CometBFTMapper) *CometBFTMessageSimulator {
	return &CometBFTMessageSimulator{
		mapper: mapper,
		height: 1000,
		round:  1,
	}
}

func runSimulationScenario(mapper *cometbftAdapter.CometBFTMapper, duration time.Duration) {
	fmt.Println("🎛️  Live CometBFT Simulation")
	fmt.Println("===========================")
	fmt.Printf("Running for %s\n\n", duration)

	simulator := NewCometBFTMessageSimulator(mapper)
	simulator.Run(duration)
}

// Run spins up the simulator and prints details for each generated message.
func (ms *CometBFTMessageSimulator) Run(duration time.Duration) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(duration)
	messageCount := 0

	for {
		select {
		case <-ticker.C:
			messageCount++
			ms.generateAndProcessMessage(messageCount)
		case <-timeout:
			fmt.Printf("Simulation finished after %d messages.\n", messageCount)
			return
		}
	}
}

func (ms *CometBFTMessageSimulator) generateAndProcessMessage(count int) {
	msgTypes := []string{"proposal", "prevote", "precommit", "new_round_step"}
	msgType := msgTypes[rand.Intn(len(msgTypes))]

	fmt.Printf("Message #%d → %s\n", count, strings.ToUpper(msgType))

	rawMsg := ms.generateRawMessage(msgType)
	printRawMessage(rawMsg)

	canonical, err := ms.mapper.ToCanonical(rawMsg)
	if err != nil {
		fmt.Printf("   conversion failed: %v\n\n", err)
		return
	}

	fmt.Println("   Raw → Canonical")
	printCanonicalMessage(canonical)

	targetRaw, err := ms.mapper.FromCanonical(canonical)
	if err != nil {
		fmt.Printf("   canonical → raw failed: %v\n\n", err)
		return
	}

	fmt.Println("   Canonical → Raw")
	printRawMessage(*targetRaw)
	fmt.Println()

	ms.height++
	if ms.height%10 == 0 {
		ms.round++
	}
}

func (ms *CometBFTMessageSimulator) generateRawMessage(msgType string) abstraction.RawConsensusMessage {
	typeNum := int32(0)
	switch msgType {
	case "proposal":
		typeNum = 32
	case "prevote":
		typeNum = 1
	case "precommit":
		typeNum = 2
	case "new_round_step":
		typeNum = 0
	}

	baseMsg := map[string]interface{}{
		"height":    fmt.Sprintf("%d", ms.height),
		"round":     fmt.Sprintf("%d", ms.round),
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      typeNum,
	}

	baseMsg["block_id"] = map[string]interface{}{
		"hash": fmt.Sprintf("0x%x", rand.Int63()),
		"parts": map[string]interface{}{
			"total": 1,
			"hash":  fmt.Sprintf("0x%x", rand.Int63()),
		},
	}
	baseMsg["proposer_address"] = fmt.Sprintf("node%d", rand.Intn(10)+1)
	baseMsg["validator_address"] = fmt.Sprintf("validator%d", rand.Intn(10)+1)
	baseMsg["signature"] = fmt.Sprintf("sig_%d", rand.Int63())

	payload, _ := json.Marshal(baseMsg)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: msgType,
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   time.Now(),
	}
}
