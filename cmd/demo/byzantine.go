package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
)

func runByzantineScenario(mapper *cometbftAdapter.CometBFTMapper, actionFlag, canonicalPath, alternateBlock, alternatePrev, alternateSig string) {
	fmt.Println("🧨 Byzantine Message Emission")
	fmt.Println("============================")

	action, err := cometbftAdapter.ParseByzantineAction(actionFlag)
	if err != nil {
		fmt.Printf("invalid byzantine action %q: %v\n", actionFlag, err)
		return
	}

	canonical, sourceDescription, err := loadCanonicalForScenario(mapper, canonicalPath)
	if err != nil {
		fmt.Printf("failed to prepare canonical message: %v\n", err)
		return
	}

	fmt.Printf("Using canonical message from %s\n", sourceDescription)
	printCanonicalMessage(canonical)

	opts := cometbftAdapter.ByzantineOptions{
		AlternateBlockHash: alternateBlock,
		AlternatePrevHash:  alternatePrev,
		AlternateSignature: alternateSig,
	}

	byzCanonicals, err := cometbftAdapter.ApplyByzantineCanonical(canonical, action, opts)
	if err != nil {
		fmt.Printf("byzantine conversion failed: %v\n", err)
		return
	}

	for i, byzCanonical := range byzCanonicals {
		fmt.Printf("\nByz-canonical #%d\n", i+1)
		printCanonicalMessage(byzCanonical)

		raw, err := mapper.FromCanonical(byzCanonical)
		if err != nil {
			fmt.Printf("failed to encode byz-canonical #%d: %v\n", i+1, err)
			return
		}

		fmt.Printf("\nForged message #%d (%s)\n", i+1, strings.ToUpper(raw.MessageType))
		printRawMessage(*raw)
	}

	fmt.Printf("\nGenerated %d CometBFT messages via action %s.\n", len(byzCanonicals), action)
}

func loadCanonicalForScenario(mapper *cometbftAdapter.CometBFTMapper, canonicalPath string) (*abstraction.CanonicalMessage, string, error) {
	if strings.TrimSpace(canonicalPath) != "" {
		canonical, err := readCanonicalFromFile(canonicalPath)
		if err != nil {
			return nil, "file://" + canonicalPath, err
		}
		if canonical.Timestamp.IsZero() {
			canonical.Timestamp = time.Now().UTC()
		}
		return canonical, "file://" + canonicalPath, nil
	}

	voteData, err := readVoteJSON()
	if err != nil {
		return nil, "examples/cometbft/Vote.json", err
	}

	rawVote, err := createRawVoteFromFixtures(voteData, "prevote_for_block")
	if err != nil {
		return nil, "examples/cometbft/Vote.json:prevote_for_block", err
	}

	canonical, err := mapper.ToCanonical(rawVote)
	if err != nil {
		return nil, "examples/cometbft/Vote.json:prevote_for_block", err
	}
	return canonical, "examples/cometbft/Vote.json:prevote_for_block", nil
}

func readCanonicalFromFile(path string) (*abstraction.CanonicalMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var canonical abstraction.CanonicalMessage
	if err := json.Unmarshal(data, &canonical); err != nil {
		return nil, err
	}
	return &canonical, nil
}
