package adapter

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"codec/message/abstraction"
)

// CometBFTMapper implements the Mapper interface for CometBFT consensus messages
type CometBFTMapper struct {
	chainID string
}

// NewCometBFTMapper creates a new CometBFT mapper
func NewCometBFTMapper(chainID string) *CometBFTMapper {
	return &CometBFTMapper{
		chainID: chainID,
	}
}

// ToCanonical converts a CometBFT raw message to canonical format
func (m *CometBFTMapper) ToCanonical(raw abstraction.RawConsensusMessage) (*abstraction.CanonicalMessage, error) {
	if raw.ChainType != abstraction.ChainTypeCometBFT {
		return nil, abstraction.ErrChainMismatch
	}

	// Parse the payload based on encoding
	var cometMsg CometBFTConsensusMessage
	switch raw.Encoding {
	case "json":
		if err := json.Unmarshal(raw.Payload, &cometMsg); err != nil {
			return nil, &abstraction.MessageValidationError{
				Field:   "payload",
				Message: fmt.Sprintf("failed to parse JSON: %v", err),
				Code:    "DECODE_FAILURE",
			}
		}
	case "proto":
		// For protobuf, we'll parse as JSON for now since codec is not available
		if err := json.Unmarshal(raw.Payload, &cometMsg); err != nil {
			return nil, &abstraction.MessageValidationError{
				Field:   "payload",
				Message: fmt.Sprintf("failed to parse protobuf as JSON: %v", err),
				Code:    "DECODE_FAILURE",
			}
		}
	default:
		return nil, &abstraction.MessageValidationError{
			Field:   "encoding",
			Message: fmt.Sprintf("unsupported encoding: %s", raw.Encoding),
			Code:    "DECODE_FAILURE",
		}
	}

	// Convert to canonical message based on message type
	canonical := &abstraction.CanonicalMessage{
		ChainID:    m.chainID,
		Height:     cometMsg.Height,
		Round:      cometMsg.Round,
		Timestamp:  cometMsg.Timestamp,
		Type:       m.mapMessageType(cometMsg.MessageType),
		RawPayload: raw.Payload,
		Extensions: map[string]interface{}{
			"cometbft_version":  cometMsg.Version,
			"step":              cometMsg.Step,
			"last_commit_round": cometMsg.LastCommitRound,
		},
	}

	// Set specific fields based on message type
	switch cometMsg.MessageType {
	case "NewRoundStep":
		canonical.Extensions["step"] = cometMsg.Step
		canonical.Extensions["last_commit_round"] = cometMsg.LastCommitRound
		canonical.Extensions["seconds_since_start_time"] = cometMsg.SecondsSinceStartTime

	case "Proposal":
		canonical.BlockHash = cometMsg.BlockID.Hash
		canonical.PrevHash = cometMsg.BlockID.PrevHash
		canonical.Proposer = cometMsg.ProposerAddress
		canonical.Signature = cometMsg.Signature
		canonical.Extensions["pol_round"] = cometMsg.POLRound
		canonical.Extensions["part_set_header"] = cometMsg.BlockID.PartSetHeader

	case "Vote":
		canonical.BlockHash = cometMsg.BlockID.Hash
		canonical.Validator = cometMsg.ValidatorAddress
		canonical.Signature = cometMsg.Signature
		canonical.Extensions["vote_type"] = cometMsg.VoteType
		canonical.Extensions["validator_index"] = cometMsg.ValidatorIndex
		canonical.Extensions["extension"] = cometMsg.Extension
		canonical.Extensions["extension_signature"] = cometMsg.ExtensionSignature

	case "BlockPart":
		canonical.BlockHash = cometMsg.BlockID.Hash
		canonical.Extensions["part_index"] = cometMsg.PartIndex
		canonical.Extensions["part_bytes"] = cometMsg.PartBytes
		canonical.Extensions["part_proof"] = cometMsg.PartProof

	case "NewValidBlock":
		canonical.BlockHash = cometMsg.BlockID.Hash
		canonical.Extensions["is_commit"] = cometMsg.IsCommit
		canonical.Extensions["block_parts"] = cometMsg.BlockParts
		canonical.Extensions["part_set_header"] = cometMsg.BlockID.PartSetHeader

	case "VoteSetMaj23", "VoteSetBits":
		canonical.BlockHash = cometMsg.BlockID.Hash
		canonical.Extensions["vote_type"] = cometMsg.VoteType
		if cometMsg.MessageType == "VoteSetBits" {
			canonical.Extensions["votes_bit_array"] = cometMsg.VotesBitArray
		}

	case "HasVote":
		canonical.Extensions["vote_type"] = cometMsg.VoteType
		canonical.Extensions["validator_index"] = cometMsg.ValidatorIndex

	case "ProposalPOL":
		canonical.Extensions["proposal_pol_round"] = cometMsg.ProposalPOLRound
		canonical.Extensions["proposal_pol"] = cometMsg.ProposalPOL
	}

	return canonical, nil
}

// FromCanonical converts a canonical message to CometBFT format
func (m *CometBFTMapper) FromCanonical(msg *abstraction.CanonicalMessage) (*abstraction.RawConsensusMessage, error) {
	if msg == nil {
		return nil, &abstraction.MessageValidationError{
			Field:   "message",
			Message: "message cannot be nil",
			Code:    "MISSING_FIELD",
		}
	}

	// Convert canonical message to CometBFT format
	cometMsg := CometBFTConsensusMessage{
		Height:    msg.Height,
		Round:     msg.Round,
		Timestamp: msg.Timestamp,
		Version:   "0.38.17", // Default version
	}

	// Set message type and specific fields
	switch msg.Type {
	case abstraction.MsgTypeProposal:
		cometMsg.MessageType = "Proposal"
		cometMsg.BlockID = BlockID{
			Hash:          msg.BlockHash,
			PrevHash:      msg.PrevHash,
			PartSetHeader: PartSetHeader{Total: 1, Hash: []byte(msg.BlockHash)},
		}
		cometMsg.ProposerAddress = msg.Proposer
		cometMsg.Signature = msg.Signature
		if polRound, ok := msg.Extensions["pol_round"].(int32); ok {
			cometMsg.POLRound = polRound
		}

	case abstraction.MsgTypePrevote:
		cometMsg.MessageType = "Vote"
		cometMsg.VoteType = "PrevoteType"
		cometMsg.BlockID = BlockID{Hash: msg.BlockHash}
		cometMsg.ValidatorAddress = msg.Validator
		cometMsg.Signature = msg.Signature

	case abstraction.MsgTypePrecommit:
		cometMsg.MessageType = "Vote"
		cometMsg.VoteType = "PrecommitType"
		cometMsg.BlockID = BlockID{Hash: msg.BlockHash}
		cometMsg.ValidatorAddress = msg.Validator
		cometMsg.Signature = msg.Signature
		if ext, ok := msg.Extensions["extension"].([]byte); ok {
			cometMsg.Extension = ext
		}
		if extSig, ok := msg.Extensions["extension_signature"].([]byte); ok {
			cometMsg.ExtensionSignature = extSig
		}

	case abstraction.MsgTypeBlock:
		cometMsg.MessageType = "BlockPart"
		cometMsg.BlockID = BlockID{Hash: msg.BlockHash}
		if partIndex, ok := msg.Extensions["part_index"].(uint32); ok {
			cometMsg.PartIndex = partIndex
		}
		if partBytes, ok := msg.Extensions["part_bytes"].([]byte); ok {
			cometMsg.PartBytes = partBytes
		}

	default:
		// Default to NewRoundStep for unknown types
		cometMsg.MessageType = "NewRoundStep"
		if step, ok := msg.Extensions["step"].(uint32); ok {
			cometMsg.Step = step
		}
		if lastCommitRound, ok := msg.Extensions["last_commit_round"].(int32); ok {
			cometMsg.LastCommitRound = lastCommitRound
		}
	}

	// Serialize to JSON (default format)
	payload, err := json.Marshal(cometMsg)
	if err != nil {
		return nil, &abstraction.MessageValidationError{
			Field:   "payload",
			Message: fmt.Sprintf("failed to serialize: %v", err),
			Code:    "DECODE_FAILURE",
		}
	}

	raw := &abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     m.chainID,
		MessageType: cometMsg.MessageType,
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"version": cometMsg.Version,
			"step":    cometMsg.Step,
		},
	}

	return raw, nil
}

// GetSupportedTypes returns the message types supported by CometBFT
func (m *CometBFTMapper) GetSupportedTypes() []abstraction.MsgType {
	return []abstraction.MsgType{
		abstraction.MsgTypeProposal,
		abstraction.MsgTypePrevote,
		abstraction.MsgTypePrecommit,
		abstraction.MsgTypeBlock,
	}
}

// GetChainType returns the chain type this mapper handles
func (m *CometBFTMapper) GetChainType() abstraction.ChainType {
	return abstraction.ChainTypeCometBFT
}

// mapMessageType maps CometBFT message types to canonical types
func (m *CometBFTMapper) mapMessageType(cometType string) abstraction.MsgType {
	switch cometType {
	case "Proposal":
		return abstraction.MsgTypeProposal
	case "Vote":
		// Vote type is determined by VoteType field
		return abstraction.MsgTypeVote // Generic vote, specific type in extensions
	case "BlockPart":
		return abstraction.MsgTypeBlock
	case "NewRoundStep", "NewValidBlock", "HasVote", "VoteSetMaj23", "VoteSetBits", "ProposalPOL":
		return abstraction.MsgTypeProposal // Map internal messages to proposal
	default:
		return abstraction.MsgType(cometType)
	}
}

// CometBFTConsensusMessage represents the internal CometBFT consensus message structure
type CometBFTConsensusMessage struct {
	// Common fields
	Height    *big.Int  `json:"height"`
	Round     *big.Int  `json:"round"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`

	// Message type and step
	MessageType string `json:"message_type"`
	Step        uint32 `json:"step,omitempty"`

	// NewRoundStep specific
	LastCommitRound       int32 `json:"last_commit_round,omitempty"`
	SecondsSinceStartTime int64 `json:"seconds_since_start_time,omitempty"`

	// Proposal specific
	BlockID         BlockID `json:"block_id,omitempty"`
	ProposerAddress string  `json:"proposer_address,omitempty"`
	Signature       string  `json:"signature,omitempty"`
	POLRound        int32   `json:"pol_round,omitempty"`

	// Vote specific
	VoteType           string `json:"vote_type,omitempty"`
	ValidatorAddress   string `json:"validator_address,omitempty"`
	ValidatorIndex     int32  `json:"validator_index,omitempty"`
	Extension          []byte `json:"extension,omitempty"`
	ExtensionSignature []byte `json:"extension_signature,omitempty"`

	// BlockPart specific
	PartIndex uint32 `json:"part_index,omitempty"`
	PartBytes []byte `json:"part_bytes,omitempty"`
	PartProof []byte `json:"part_proof,omitempty"`

	// NewValidBlock specific
	IsCommit   bool     `json:"is_commit,omitempty"`
	BlockParts []string `json:"block_parts,omitempty"`

	// VoteSetBits specific
	VotesBitArray []string `json:"votes_bit_array,omitempty"`

	// ProposalPOL specific
	ProposalPOLRound int32    `json:"proposal_pol_round,omitempty"`
	ProposalPOL      []string `json:"proposal_pol,omitempty"`
}

// BlockID represents a CometBFT BlockID
type BlockID struct {
	Hash          string        `json:"hash,omitempty"`
	PrevHash      string        `json:"prev_hash,omitempty"`
	PartSetHeader PartSetHeader `json:"part_set_header,omitempty"`
}

// PartSetHeader represents a CometBFT PartSetHeader
type PartSetHeader struct {
	Total uint32 `json:"total"`
	Hash  []byte `json:"hash"`
}
