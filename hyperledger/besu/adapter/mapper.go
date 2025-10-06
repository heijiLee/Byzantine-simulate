package adapter

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"

	"codec/message/abstraction"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// BesuIBFTMessage represents Besu's IBFT2.0/QBFT message structure
type BesuIBFTMessage struct {
	Code      uint8       `json:"code"`       // Message type code
	Height    *big.Int    `json:"height"`     // Block height (h)
	Round     uint64      `json:"round"`      // Round number (r)
	BlockHash common.Hash `json:"block_hash"` // Block hash (H)
	Signature []byte      `json:"signature"`  // ECDSA signature (65 bytes)
}

// BesuCommitPayload represents IBFT Commit message with seal
type BesuCommitPayload struct {
	Body       BesuIBFTMessage `json:"body"`
	CommitSeal []byte          `json:"commit_seal"` // 65-byte secp256k1 signature
}

// BesuBlockExtraData represents Besu's block extraData RLP structure
type BesuBlockExtraData struct {
	Vanity     []byte    `json:"vanity"`     // 32-byte vanity data
	Validators [][]byte  `json:"validators"` // List of validator addresses
	Vote       *VoteData `json:"vote"`       // Vote data (nil in genesis)
	Round      uint64    `json:"round"`      // Round number (0 in genesis)
	Seals      [][]byte  `json:"seals"`      // List of commit seals
}

// VoteData represents IBFT vote information
type VoteData struct {
	RecipientAddress common.Address `json:"recipient_address"`
	VoteType         uint8          `json:"vote_type"` // 0=add, 1=remove
}

// BesuMapper implements the Mapper interface for Hyperledger Besu
type BesuMapper struct {
	chainID string
}

// NewBesuMapper creates a new Besu mapper
func NewBesuMapper(chainID string) *BesuMapper {
	return &BesuMapper{
		chainID: chainID,
	}
}

// ToCanonical converts a Besu message to canonical format
func (m *BesuMapper) ToCanonical(raw abstraction.RawConsensusMessage) (*abstraction.CanonicalMessage, error) {
	if raw.ChainType != abstraction.ChainTypeHyperledger {
		return nil, fmt.Errorf("invalid chain type: expected %s, got %s", abstraction.ChainTypeHyperledger, raw.ChainType)
	}

	// Parse the raw payload based on message type
	var canonicalType abstraction.MsgType
	var height *big.Int
	var round *big.Int
	var blockHash string
	var proposer string
	var validator string
	var signature string

	switch raw.MessageType {
	case "Proposal":
		var proposal BesuIBFTMessage
		if err := json.Unmarshal(raw.Payload, &proposal); err != nil {
			return nil, fmt.Errorf("failed to parse proposal: %w", err)
		}
		canonicalType = abstraction.MsgTypeProposal
		height = proposal.Height
		round = big.NewInt(int64(proposal.Round))
		blockHash = proposal.BlockHash.Hex()
		signature = fmt.Sprintf("0x%x", proposal.Signature)

	case "Prepare":
		var prepare BesuIBFTMessage
		if err := json.Unmarshal(raw.Payload, &prepare); err != nil {
			return nil, fmt.Errorf("failed to parse prepare: %w", err)
		}
		canonicalType = abstraction.MsgTypePrepare
		height = prepare.Height
		round = big.NewInt(int64(prepare.Round))
		blockHash = prepare.BlockHash.Hex()
		signature = fmt.Sprintf("0x%x", prepare.Signature)

	case "Commit":
		var commit BesuCommitPayload
		if err := json.Unmarshal(raw.Payload, &commit); err != nil {
			return nil, fmt.Errorf("failed to parse commit: %w", err)
		}
		canonicalType = abstraction.MsgTypeCommit
		height = commit.Body.Height
		round = big.NewInt(int64(commit.Body.Round))
		blockHash = commit.Body.BlockHash.Hex()
		signature = fmt.Sprintf("0x%x", commit.CommitSeal)

	case "RoundChange":
		var roundChange BesuIBFTMessage
		if err := json.Unmarshal(raw.Payload, &roundChange); err != nil {
			return nil, fmt.Errorf("failed to parse roundchange: %w", err)
		}
		canonicalType = abstraction.MsgTypeRoundChange
		height = roundChange.Height
		round = big.NewInt(int64(roundChange.Round))
		blockHash = roundChange.BlockHash.Hex()
		signature = fmt.Sprintf("0x%x", roundChange.Signature)

	default:
		return nil, fmt.Errorf("unsupported message type: %s", raw.MessageType)
	}

	// Extract validator from metadata
	if val, ok := raw.Metadata["validator"].(string); ok {
		validator = val
		proposer = val
	}

	// Create canonical message
	canonical := &abstraction.CanonicalMessage{
		ChainID:   m.chainID,
		Height:    height,
		Round:     round,
		Timestamp: raw.Timestamp,
		Type:      canonicalType,
		BlockHash: blockHash,
		Proposer:  proposer,
		Validator: validator,
		Signature: signature,
		Extensions: map[string]interface{}{
			"ibft_type":       raw.MessageType,
			"gas_limit":       raw.Metadata["gas_limit"],
			"gas_used":        raw.Metadata["gas_used"],
			"tx_count":        raw.Metadata["tx_count"],
			"validator_count": raw.Metadata["validator_count"],
			"consensus_type":  raw.Metadata["consensus_type"], // IBFT2.0 or QBFT
		},
	}

	return canonical, nil
}

// FromCanonical converts a canonical message to Besu format
func (m *BesuMapper) FromCanonical(canonical *abstraction.CanonicalMessage) (*abstraction.RawConsensusMessage, error) {
	if canonical.ChainID != m.chainID {
		return nil, fmt.Errorf("chain ID mismatch: expected %s, got %s", m.chainID, canonical.ChainID)
	}

	// Extract Besu-specific extensions
	gasLimit := uint64(30000000)
	gasUsed := uint64(15000000)
	txCount := 100
	validatorCount := 4
	consensusType := "IBFT2.0"

	if canonical.Extensions != nil {
		if gl, ok := canonical.Extensions["gas_limit"].(uint64); ok {
			gasLimit = gl
		}
		if gu, ok := canonical.Extensions["gas_used"].(uint64); ok {
			gasUsed = gu
		}
		if tc, ok := canonical.Extensions["tx_count"].(int); ok {
			txCount = tc
		}
		if vc, ok := canonical.Extensions["validator_count"].(int); ok {
			validatorCount = vc
		}
		if ct, ok := canonical.Extensions["consensus_type"].(string); ok {
			consensusType = ct
		}
	}

	// Convert block hash
	var blockHash common.Hash
	if canonical.BlockHash != "" {
		blockHash = common.HexToHash(canonical.BlockHash)
	}

	// Create Besu message based on type
	var payload []byte
	var msgType string
	var err error

	switch canonical.Type {
	case abstraction.MsgTypeProposal:
		besuMsg := BesuIBFTMessage{
			Code:      0x00, // MsgProposal
			Height:    canonical.Height,
			Round:     canonical.Round.Uint64(),
			BlockHash: blockHash,
			Signature: []byte("proposal_signature"), // In real implementation, this would be actual signature
		}
		payload, err = json.Marshal(besuMsg)
		msgType = "Proposal"

	case abstraction.MsgTypePrepare:
		besuMsg := BesuIBFTMessage{
			Code:      0x01, // MsgPrepare
			Height:    canonical.Height,
			Round:     canonical.Round.Uint64(),
			BlockHash: blockHash,
			Signature: []byte("prepare_signature"),
		}
		payload, err = json.Marshal(besuMsg)
		msgType = "Prepare"

	case abstraction.MsgTypeCommit:
		body := BesuIBFTMessage{
			Code:      0x02, // MsgCommit
			Height:    canonical.Height,
			Round:     canonical.Round.Uint64(),
			BlockHash: blockHash,
			Signature: []byte("commit_body_signature"),
		}
		commitPayload := BesuCommitPayload{
			Body:       body,
			CommitSeal: []byte("commit_seal_signature"), // In real implementation, this would be actual seal
		}
		payload, err = json.Marshal(commitPayload)
		msgType = "Commit"

	case abstraction.MsgTypeRoundChange:
		besuMsg := BesuIBFTMessage{
			Code:      0x03, // MsgRoundChange
			Height:    canonical.Height,
			Round:     canonical.Round.Uint64(),
			BlockHash: blockHash,
			Signature: []byte("roundchange_signature"),
		}
		payload, err = json.Marshal(besuMsg)
		msgType = "RoundChange"

	default:
		return nil, fmt.Errorf("unsupported canonical message type: %s", canonical.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal besu message: %w", err)
	}

	return &abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeHyperledger,
		ChainID:     m.chainID,
		MessageType: msgType,
		Payload:     payload,
		Encoding:    "rlp",
		Timestamp:   canonical.Timestamp,
		Metadata: map[string]interface{}{
			"gas_limit":       gasLimit,
			"gas_used":        gasUsed,
			"tx_count":        txCount,
			"validator_count": validatorCount,
			"consensus_type":  consensusType,
			"validator":       canonical.Validator,
			"source":          "besu_mapper",
		},
	}, nil
}

// GetSupportedTypes returns the supported message types
func (m *BesuMapper) GetSupportedTypes() []abstraction.MsgType {
	return []abstraction.MsgType{
		abstraction.MsgTypeProposal,
		abstraction.MsgTypePrepare,
		abstraction.MsgTypeCommit,
		abstraction.MsgTypeRoundChange,
	}
}

// GetChainType returns the chain type
func (m *BesuMapper) GetChainType() abstraction.ChainType {
	return abstraction.ChainTypeHyperledger
}

// Helper function to create IBFT extraData RLP structure
func CreateIBFTExtraData(validators []common.Address, vote *VoteData, round uint64, seals [][]byte) ([]byte, error) {
	vanity := make([]byte, 32) // 32-byte vanity data
	copy(vanity, []byte("besu-ibft-consensus"))

	extraData := BesuBlockExtraData{
		Vanity:     vanity,
		Validators: make([][]byte, len(validators)),
		Vote:       vote,
		Round:      round,
		Seals:      seals,
	}

	// Convert addresses to bytes
	for i, addr := range validators {
		extraData.Validators[i] = addr.Bytes()
	}

	return rlp.EncodeToBytes(extraData)
}

// Helper function to sign IBFT message body
func SignIBFTMessage(privKey *ecdsa.PrivateKey, body BesuIBFTMessage) ([]byte, error) {
	bodyBytes, err := rlp.EncodeToBytes(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %w", err)
	}

	hash := crypto.Keccak256Hash(bodyBytes)
	signature, err := crypto.Sign(hash.Bytes(), privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return signature, nil
}
