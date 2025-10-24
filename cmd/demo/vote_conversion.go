package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func runVoteBatchScenario(mapper *cometbftAdapter.CometBFTMapper) {
	fmt.Println("🧪 CometBFT Vote Round-Trip")
	fmt.Println("===========================")

	voteData, err := readVoteJSON()
	if err != nil {
		fmt.Printf("failed to read examples/cometbft/Vote.json: %v\n", err)
		return
	}

	fmt.Println("Loaded vote fixtures from examples/cometbft/Vote.json")

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
		fmt.Printf("\nCase %d → %s\n", i+1, tc.name)
		fmt.Println(strings.Repeat("-", len(tc.name)+10))

		if runVoteCase(voteData, tc.key, mapper) {
			successCount++
			fmt.Println("Result: success")
		} else {
			fmt.Println("Result: failure")
		}
	}

	fmt.Printf("\nSummary: %d/%d cases succeeded (%.1f%%).\n", successCount, len(testCases), float64(successCount)/float64(len(testCases))*100)
}

func runVoteCase(voteData map[string]interface{}, key string, mapper *cometbftAdapter.CometBFTMapper) bool {
	rawVote, err := createRawVoteFromFixtures(voteData, key)
	if err != nil {
		fmt.Printf("   unable to prepare vote fixture: %v\n", err)
		return false
	}

	fmt.Println("   Fixture → Raw consensus message")
	printRawMessage(rawVote)

	fmt.Println("   Raw → Canonical")
	canonical, err := mapper.ToCanonical(rawVote)
	if err != nil {
		fmt.Printf("   canonical conversion failed: %v\n", err)
		return false
	}
	printCanonicalMessage(canonical)

	fmt.Println("   Canonical → Raw")
	rawConverted, err := mapper.FromCanonical(canonical)
	if err != nil {
		fmt.Printf("   reverse conversion failed: %v\n", err)
		return false
	}
	printRawMessage(*rawConverted)

	fmt.Println("   Comparing original and converted payloads")
	if compareVoteMessages(rawVote, *rawConverted) {
		printConversionSummary(canonical)
		return true
	}

	return false
}

func readVoteJSON() (map[string]interface{}, error) {
	file, err := os.Open("examples/cometbft/Vote.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var voteData map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&voteData); err != nil {
		return nil, err
	}
	return voteData, nil
}

func createRawVoteFromFixtures(voteData map[string]interface{}, key string) (abstraction.RawConsensusMessage, error) {
	vote, exists := voteData[key]
	if !exists {
		return abstraction.RawConsensusMessage{}, fmt.Errorf("vote fixture %q not found", key)
	}

	jsonPayload, err := json.Marshal(vote)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	timestamp, err := extractTimestamp(vote)
	if err != nil {
		return abstraction.RawConsensusMessage{}, err
	}

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "cosmos-hub-4",
		MessageType: "Vote",
		Payload:     jsonPayload,
		Encoding:    "json",
		Timestamp:   timestamp,
		Metadata: map[string]interface{}{
			"source": key,
		},
	}, nil
}

func extractTimestamp(vote interface{}) (time.Time, error) {
	voteMap, ok := vote.(map[string]interface{})
	if !ok {
		return time.Time{}, errors.New("vote payload is not an object")
	}

	timestampStr, ok := voteMap["timestamp"].(string)
	if !ok || timestampStr == "" {
		return time.Time{}, nil
	}

	ts, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func compareVoteMessages(original, converted abstraction.RawConsensusMessage) bool {
	if original.ChainType != converted.ChainType {
		fmt.Printf("      chain type mismatch: %s != %s\n", original.ChainType, converted.ChainType)
		return false
	}
	if original.MessageType != converted.MessageType {
		fmt.Printf("      message type mismatch: %s != %s\n", original.MessageType, converted.MessageType)
		return false
	}

	var origPayload, convPayload map[string]interface{}
	if err := json.Unmarshal(original.Payload, &origPayload); err != nil {
		fmt.Printf("      failed to decode original payload: %v\n", err)
		return false
	}
	if err := json.Unmarshal(converted.Payload, &convPayload); err != nil {
		fmt.Printf("      failed to decode converted payload: %v\n", err)
		return false
	}

	keyFields := []string{"type", "height", "round", "validator_address", "signature"}
	for _, field := range keyFields {
		if fmt.Sprintf("%v", origPayload[field]) != fmt.Sprintf("%v", convPayload[field]) {
			fmt.Printf("      field mismatch for %s: %v != %v\n", field, origPayload[field], convPayload[field])
			return false
		}
	}

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
		origStr := fmt.Sprintf("%v", orig)
		convStr := fmt.Sprintf("%v", conv)
		if origStr == "<nil>" {
			origStr = ""
		}
		if convStr == "<nil>" {
			convStr = ""
		}
		if origStr == convStr {
			return true
		}
		fmt.Printf("      block ID mismatch: %v != %v\n", orig, conv)
		return false
	}

	origMap, origOk := orig.(map[string]interface{})
	convMap, convOk := conv.(map[string]interface{})
	if !origOk || !convOk {
		fmt.Printf("      block ID type mismatch: %T != %T\n", orig, conv)
		return false
	}

	origHash := fmt.Sprintf("%v", origMap["hash"])
	convHash := fmt.Sprintf("%v", convMap["hash"])
	if origHash == "<nil>" {
		origHash = ""
	}
	if convHash == "<nil>" {
		convHash = ""
	}

	if origHash != convHash {
		fmt.Printf("      block hash mismatch: %s != %s\n", origHash, convHash)
		return false
	}

	return true
}

func printConversionSummary(canonical *abstraction.CanonicalMessage) {
	fmt.Println("   Summary")
	fmt.Printf("      Type: %s\n", canonical.Type)
	fmt.Printf("      Height: %v\n", canonical.Height)
	fmt.Printf("      Round: %v\n", canonical.Round)
	if canonical.BlockHash != "" {
		fmt.Printf("      Block hash: %s\n", canonical.BlockHash[:min(32, len(canonical.BlockHash))])
	}
	fmt.Printf("      Validator: %s\n", canonical.Validator)
	fmt.Printf("      Extension count: %d\n", len(canonical.Extensions))
}
