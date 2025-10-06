package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	besuAdapter "codec/hyperledger/besu/adapter"
	"codec/message/abstraction"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	fmt.Println("🏗️  Hyperledger Besu IBFT2.0/QBFT Message Generator")
	fmt.Println("=====================================================")

	// Create examples directory
	examplesDir := "examples/besu"
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create examples directory: %v\n", err)
		return
	}

	// Initialize Besu mapper
	mapper := besuAdapter.NewBesuMapper("besu-testnet")

	// Generate different types of Besu IBFT messages
	messages := []abstraction.RawConsensusMessage{
		generateBesuProposal(),
		generateBesuPrepare(),
		generateBesuCommit(),
		generateBesuRoundChange(),
	}

	// Save individual message files
	for i, msg := range messages {
		filename := filepath.Join(examplesDir, fmt.Sprintf("%s.json", msg.MessageType))
		if err := saveMessageToFile(msg, filename); err != nil {
			fmt.Printf("❌ Failed to save %s: %v\n", msg.MessageType, err)
			continue
		}
		fmt.Printf("✅ Saved %s\n", filename)
	}

	// Save all messages in one file
	allFilename := filepath.Join(examplesDir, "all_messages.json")
	if err := saveMessagesToFile(messages, allFilename); err != nil {
		fmt.Printf("❌ Failed to save all messages: %v\n", err)
		return
	}
	fmt.Printf("✅ Saved %s\n", allFilename)

	// Save sample messages
	sampleFilename := filepath.Join(examplesDir, "samples.json")
	sampleMessages := messages[:3] // First 3 messages as samples
	if err := saveMessagesToFile(sampleMessages, sampleFilename); err != nil {
		fmt.Printf("❌ Failed to save samples: %v\n", err)
		return
	}
	fmt.Printf("✅ Saved %s\n", sampleFilename)

	// Test conversion to canonical format
	fmt.Println("\n🔄 Testing Besu → Canonical Conversion")
	fmt.Println("======================================")
	for i, msg := range messages {
		fmt.Printf("\n📦 Message %d: %s\n", i+1, msg.MessageType)

		canonical, err := mapper.ToCanonical(msg)
		if err != nil {
			fmt.Printf("❌ Conversion failed: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Height: %d\n", canonical.Height.Uint64())
		fmt.Printf("   ✅ Round: %d\n", canonical.Round.Uint64())
		fmt.Printf("   ✅ Type: %s\n", canonical.Type)
		fmt.Printf("   ✅ Block Hash: %s\n", canonical.BlockHash)
		fmt.Printf("   ✅ Proposer: %s\n", canonical.Proposer)
		fmt.Printf("   ✅ Validator: %s\n", canonical.Validator)

		if len(canonical.Extensions) > 0 {
			fmt.Printf("   ✅ Extensions: %d items\n", len(canonical.Extensions))
			if gasLimit, ok := canonical.Extensions["gas_limit"].(uint64); ok {
				fmt.Printf("      - Gas Limit: %d\n", gasLimit)
			}
			if consensusType, ok := canonical.Extensions["consensus_type"].(string); ok {
				fmt.Printf("      - Consensus Type: %s\n", consensusType)
			}
		}
	}

	fmt.Println("\n🎉 Besu IBFT message generation completed!")
	fmt.Printf("📁 Check the %s directory for generated files\n", examplesDir)
}

func generateBesuProposal() abstraction.RawConsensusMessage {
	height := big.NewInt(1000000)
	round := uint64(0)
	blockHash := common.HexToHash("0x1234567890abcdef")

	// Create IBFT proposal message
	proposal := besuAdapter.BesuIBFTMessage{
		Code:      0x00, // MsgProposal
		Height:    height,
		Round:     round,
		BlockHash: blockHash,
		Signature: generateSignature(height, round, blockHash),
	}

	payload, _ := json.Marshal(proposal)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeHyperledger,
		ChainID:     "besu-testnet",
		MessageType: "Proposal",
		Payload:     payload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"gas_limit":       30000000,
			"gas_used":        15000000,
			"tx_count":        150,
			"validator_count": 4,
			"consensus_type":  "IBFT2.0",
			"validator":       "validator0",
			"source":          "besu_generator",
		},
	}
}

func generateBesuPrepare() abstraction.RawConsensusMessage {
	height := big.NewInt(1000000)
	round := uint64(0)
	blockHash := common.HexToHash("0x1234567890abcdef")

	// Create IBFT prepare message
	prepare := besuAdapter.BesuIBFTMessage{
		Code:      0x01, // MsgPrepare
		Height:    height,
		Round:     round,
		BlockHash: blockHash,
		Signature: generateSignature(height, round, blockHash),
	}

	payload, _ := json.Marshal(prepare)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeHyperledger,
		ChainID:     "besu-testnet",
		MessageType: "Prepare",
		Payload:     payload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"gas_limit":       30000000,
			"gas_used":        15000000,
			"tx_count":        150,
			"validator_count": 4,
			"consensus_type":  "IBFT2.0",
			"validator":       "validator1",
			"source":          "besu_generator",
		},
	}
}

func generateBesuCommit() abstraction.RawConsensusMessage {
	height := big.NewInt(1000000)
	round := uint64(0)
	blockHash := common.HexToHash("0x1234567890abcdef")

	// Create IBFT commit message with seal
	body := besuAdapter.BesuIBFTMessage{
		Code:      0x02, // MsgCommit
		Height:    height,
		Round:     round,
		BlockHash: blockHash,
		Signature: generateSignature(height, round, blockHash),
	}

	commitPayload := besuAdapter.BesuCommitPayload{
		Body:       body,
		CommitSeal: generateCommitSeal(height, round, blockHash),
	}

	payload, _ := json.Marshal(commitPayload)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeHyperledger,
		ChainID:     "besu-testnet",
		MessageType: "Commit",
		Payload:     payload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"gas_limit":       30000000,
			"gas_used":        15000000,
			"tx_count":        150,
			"validator_count": 4,
			"consensus_type":  "IBFT2.0",
			"validator":       "validator2",
			"source":          "besu_generator",
		},
	}
}

func generateBesuRoundChange() abstraction.RawConsensusMessage {
	height := big.NewInt(1000000)
	round := uint64(1)         // Next round
	blockHash := common.Hash{} // Empty for round change

	// Create IBFT round change message
	roundChange := besuAdapter.BesuIBFTMessage{
		Code:      0x03, // MsgRoundChange
		Height:    height,
		Round:     round,
		BlockHash: blockHash,
		Signature: generateSignature(height, round, blockHash),
	}

	payload, _ := json.Marshal(roundChange)

	return abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeHyperledger,
		ChainID:     "besu-testnet",
		MessageType: "RoundChange",
		Payload:     payload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"gas_limit":       30000000,
			"gas_used":        15000000,
			"tx_count":        150,
			"validator_count": 4,
			"consensus_type":  "IBFT2.0",
			"validator":       "validator3",
			"source":          "besu_generator",
		},
	}
}

func generateSignature(height *big.Int, round uint64, blockHash common.Hash) []byte {
	// Generate a random private key for demonstration
	privKey, _ := crypto.GenerateKey()

	// Create message body for signing
	body := besuAdapter.BesuIBFTMessage{
		Code:      0x00,
		Height:    height,
		Round:     round,
		BlockHash: blockHash,
		Signature: nil, // Will be filled after signing
	}

	// Sign the message body
	signature, _ := besuAdapter.SignIBFTMessage(privKey, body)
	return signature
}

func generateCommitSeal(height *big.Int, round uint64, blockHash common.Hash) []byte {
	// Generate a random private key for demonstration
	privKey, _ := crypto.GenerateKey()

	// Create message body for commit seal
	body := besuAdapter.BesuIBFTMessage{
		Code:      0x02, // MsgCommit
		Height:    height,
		Round:     round,
		BlockHash: blockHash,
		Signature: nil,
	}

	// Sign the message body for commit seal
	signature, _ := besuAdapter.SignIBFTMessage(privKey, body)
	return signature
}

func saveMessageToFile(msg abstraction.RawConsensusMessage, filename string) error {
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func saveMessagesToFile(messages []abstraction.RawConsensusMessage, filename string) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
