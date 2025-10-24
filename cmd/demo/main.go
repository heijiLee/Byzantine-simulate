package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
)

const (
	scenarioOverview   = "overview"
	scenarioSimulation = "simulation"
	scenarioVoteBatch  = "vote-batch"
	scenarioByzantine  = "byzantine"
)

func main() {
	scenario := flag.String("scenario", scenarioOverview, "Scenario to run (overview|simulation|vote-batch|byzantine)")
	duration := flag.Duration("duration", 12*time.Second, "Duration for the live simulation scenario")
	actionFlag := flag.String("action", string(cometbftAdapter.ByzantineActionDoubleVote), "Byzantine action to apply (double_vote|double_proposal|alter_validator|drop_signature|timestamp_skew|none)")
	canonicalPath := flag.String("canonical", "", "Path to a canonical message JSON file for the byzantine scenario")
	chainID := flag.String("chain-id", "cosmos-hub-4", "Chain identifier used when re-encoding messages")
	alternateBlock := flag.String("alternate-block", "", "Alternate block hash used for forged outputs")
	alternatePrev := flag.String("alternate-prev", "", "Alternate previous block hash used for forged proposals")
	alternateSig := flag.String("alternate-signature", "", "Alternate signature applied to forged messages")
	alternateValidator := flag.String("alternate-validator", "", "Alternate validator/proposer ID used by alter_validator")
	roundOffset := flag.Int("round-offset", 0, "Offset (positive or negative) applied to the canonical round")
	heightOffset := flag.Int("height-offset", 0, "Offset (positive or negative) applied to the canonical height")
	timestampSkew := flag.Duration("timestamp-skew", 0, "Duration added to canonical timestamps during mutation")
	flag.Parse()

	mapper := cometbftAdapter.NewCometBFTMapper(*chainID)

	switch strings.ToLower(*scenario) {
	case scenarioOverview:
		printOverview()
	case scenarioSimulation:
		runSimulationScenario(mapper, *duration)
	case scenarioVoteBatch:
		runVoteBatchScenario(mapper)
	case scenarioByzantine:
		runByzantineScenario(mapper, *actionFlag, *canonicalPath, *alternateBlock, *alternatePrev, *alternateSig, *alternateValidator, int64(*roundOffset), int64(*heightOffset), *timestampSkew)
	default:
		fmt.Fprintf(os.Stderr, "unknown scenario %q\n", *scenario)
		os.Exit(1)
	}
}

func printOverview() {
	fmt.Println("PBFT Message Abstraction Demo")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Scenarios:")
	fmt.Println("  - simulation: Stream randomly generated CometBFT messages through the canonical mapper.")
	fmt.Println("  - vote-batch: Replay vote samples from examples/cometbft/Vote.json and verify round-trips.")
	fmt.Println("  - byzantine:  Emit forged CometBFT payloads from a canonical message using the byzantine pipeline.")
	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Println("  go run cmd/demo/main.go -scenario=simulation -duration=15s")
	fmt.Println("  go run cmd/demo/main.go -scenario=vote-batch")
	fmt.Println("  go run cmd/demo/main.go -scenario=byzantine -action=double_proposal")
}
