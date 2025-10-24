package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

type outputMessage struct {
	ChainType   abstraction.ChainType  `json:"chain_type"`
	ChainID     string                 `json:"chain_id"`
	MessageType string                 `json:"message_type"`
	Encoding    string                 `json:"encoding"`
	Timestamp   string                 `json:"timestamp"`
	Payload     json.RawMessage        `json:"payload"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func main() {
	inputPath := flag.String("input", "", "Path to a canonical message JSON file")
	actionFlag := flag.String("action", string(cometbftAdapter.ByzantineActionDoubleVote), "Byzantine action to apply (double-vote|double-proposal|none)")
	chainID := flag.String("chain-id", "cosmos-hub-4", "Chain identifier used when re-encoding the message")
	alternateBlock := flag.String("alternate-block", "", "Alternate block hash to use for the forged message")
	alternatePrev := flag.String("alternate-prev-hash", "", "Alternate previous block hash (used for proposals)")
	alternateSig := flag.String("alternate-signature", "", "Alternate signature to attach to the forged message")
	outputPath := flag.String("output", "", "Optional path to write the resulting CometBFT messages as JSON")
	flag.Parse()

	if strings.TrimSpace(*inputPath) == "" {
		fmt.Fprintln(os.Stderr, "input path is required")
		os.Exit(1)
	}

	canonical, err := loadCanonical(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load canonical message: %v\n", err)
		os.Exit(1)
	}

	mapper := cometbftAdapter.NewCometBFTMapper(*chainID)

	action, err := cometbftAdapter.ParseByzantineAction(*actionFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid byzantine action: %v\n", err)
		os.Exit(1)
	}

	opts := cometbftAdapter.ByzantineOptions{
		AlternateBlockHash: *alternateBlock,
		AlternatePrevHash:  *alternatePrev,
		AlternateSignature: *alternateSig,
	}

	rawMessages, err := mapper.FromCanonicalByzantine(canonical, action, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "conversion failed: %v\n", err)
		os.Exit(1)
	}

	outputs := make([]outputMessage, len(rawMessages))
	for i, raw := range rawMessages {
		outputs[i] = outputMessage{
			ChainType:   raw.ChainType,
			ChainID:     raw.ChainID,
			MessageType: raw.MessageType,
			Encoding:    raw.Encoding,
			Timestamp:   raw.Timestamp.Format(time.RFC3339Nano),
			Payload:     json.RawMessage(raw.Payload),
			Metadata:    raw.Metadata,
		}
	}

	result, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(*outputPath) != "" {
		if err := os.WriteFile(*outputPath, result, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated %d messages with action %s and wrote them to %s\n", len(outputs), action, *outputPath)
		return
	}

	fmt.Println(string(result))
}

func loadCanonical(path string) (*abstraction.CanonicalMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var canonical abstraction.CanonicalMessage
	if err := json.Unmarshal(data, &canonical); err != nil {
		return nil, err
	}
	if canonical.Timestamp.IsZero() {
		canonical.Timestamp = time.Now().UTC()
	}
	return &canonical, nil
}
