package adapter

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"codec/message/abstraction"
)

// BesuMapper implements the Mapper interface for Hyperledger Besu consensus messages
type BesuMapper struct {
	chainID string
}

// NewBesuMapper creates a new Besu mapper
func NewBesuMapper(chainID string) *BesuMapper {
	return &BesuMapper{
		chainID: chainID,
	}
}

// ToCanonical converts a Besu raw message to canonical format
func (m *BesuMapper) ToCanonical(raw abstraction.RawConsensusMessage) (*abstraction.CanonicalMessage, error) {
	if raw.ChainType != abstraction.ChainTypeHyperledger {
		return nil, abstraction.ErrChainMismatch
	}

	// Parse the payload based on encoding
	var besuMsg BesuMessage
	switch raw.Encoding {
	case "json":
		if err := json.Unmarshal(raw.Payload, &besuMsg); err != nil {
			return nil, &abstraction.MessageValidationError{
				Field:   "payload",
				Message: fmt.Sprintf("failed to parse JSON: %v", err),
				Code:    "DECODE_FAILURE",
			}
		}
	case "rlp":
		// For RLP encoding, we'll use a simplified approach
		// In a real implementation, you'd use proper RLP decoding
		if err := json.Unmarshal(raw.Payload, &besuMsg); err != nil {
			return nil, &abstraction.MessageValidationError{
				Field:   "payload",
				Message: fmt.Sprintf("failed to parse RLP: %v", err),
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

	// Convert to canonical message
	canonical := &abstraction.CanonicalMessage{
		ChainID:    m.chainID,
		Height:     besuMsg.BlockNumber,
		Round:      besuMsg.RoundNumber,
		Timestamp:  besuMsg.Timestamp,
		Type:       m.mapMessageType(besuMsg.Type),
		BlockHash:  besuMsg.BlockHash,
		PrevHash:   besuMsg.ParentHash,
		Proposer:   besuMsg.Proposer,
		Validator:  besuMsg.Validator,
		Signature:  besuMsg.Signature,
		RawPayload: raw.Payload,
		Extensions: map[string]interface{}{
			"gas_limit":       besuMsg.GasLimit,
			"gas_used":        besuMsg.GasUsed,
			"tx_count":        besuMsg.TxCount,
			"validator_count": besuMsg.ValidatorCount,
			"consensus_type":  besuMsg.ConsensusType,
		},
	}

	return canonical, nil
}

// FromCanonical converts a canonical message to Besu format
func (m *BesuMapper) FromCanonical(msg *abstraction.CanonicalMessage) (*abstraction.RawConsensusMessage, error) {
	if msg == nil {
		return nil, &abstraction.MessageValidationError{
			Field:   "message",
			Message: "message cannot be nil",
			Code:    "MISSING_FIELD",
		}
	}

	// Extract Besu-specific extensions
	gasLimit := uint64(0)
	gasUsed := uint64(0)
	txCount := 0
	validatorCount := 0
	consensusType := "ibft2"

	if msg.Extensions != nil {
		if gl, ok := msg.Extensions["gas_limit"].(float64); ok {
			gasLimit = uint64(gl)
		}
		if gu, ok := msg.Extensions["gas_used"].(float64); ok {
			gasUsed = uint64(gu)
		}
		if tc, ok := msg.Extensions["tx_count"].(int); ok {
			txCount = tc
		}
		if vc, ok := msg.Extensions["validator_count"].(int); ok {
			validatorCount = vc
		}
		if ct, ok := msg.Extensions["consensus_type"].(string); ok {
			consensusType = ct
		}
	}

	// Convert canonical message to Besu format
	besuMsg := BesuMessage{
		BlockNumber:    msg.Height,
		RoundNumber:    msg.Round,
		Timestamp:      msg.Timestamp,
		Type:           m.mapToBesuType(msg.Type),
		BlockHash:      msg.BlockHash,
		ParentHash:     msg.PrevHash,
		Proposer:       msg.Proposer,
		Validator:      msg.Validator,
		Signature:      msg.Signature,
		GasLimit:       gasLimit,
		GasUsed:        gasUsed,
		TxCount:        txCount,
		ValidatorCount: validatorCount,
		ConsensusType:  consensusType,
	}

	// Serialize to JSON (default format)
	payload, err := json.Marshal(besuMsg)
	if err != nil {
		return nil, &abstraction.MessageValidationError{
			Field:   "payload",
			Message: fmt.Sprintf("failed to serialize: %v", err),
			Code:    "DECODE_FAILURE",
		}
	}

	raw := &abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeHyperledger,
		ChainID:     m.chainID,
		MessageType: string(besuMsg.Type),
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"gas_limit":       besuMsg.GasLimit,
			"gas_used":        besuMsg.GasUsed,
			"tx_count":        besuMsg.TxCount,
			"validator_count": besuMsg.ValidatorCount,
			"consensus_type":  besuMsg.ConsensusType,
		},
	}

	return raw, nil
}

// GetSupportedTypes returns the message types supported by Besu
func (m *BesuMapper) GetSupportedTypes() []abstraction.MsgType {
	return []abstraction.MsgType{
		abstraction.MsgTypeProposal,
		abstraction.MsgTypePrepare,
		abstraction.MsgTypeCommit,
		abstraction.MsgTypeViewChange,
		abstraction.MsgTypeNewView,
	}
}

// GetChainType returns the chain type this mapper handles
func (m *BesuMapper) GetChainType() abstraction.ChainType {
	return abstraction.ChainTypeHyperledger
}

// mapMessageType maps Besu message types to canonical types
func (m *BesuMapper) mapMessageType(besuType string) abstraction.MsgType {
	switch besuType {
	case "PROPOSAL":
		return abstraction.MsgTypeProposal
	case "PREPARE":
		return abstraction.MsgTypePrepare
	case "COMMIT":
		return abstraction.MsgTypeCommit
	case "VIEW_CHANGE":
		return abstraction.MsgTypeViewChange
	case "NEW_VIEW":
		return abstraction.MsgTypeNewView
	default:
		return abstraction.MsgType(besuType)
	}
}

// mapToBesuType maps canonical message types to Besu types
func (m *BesuMapper) mapToBesuType(canonicalType abstraction.MsgType) string {
	switch canonicalType {
	case abstraction.MsgTypeProposal:
		return "PROPOSAL"
	case abstraction.MsgTypePrepare:
		return "PREPARE"
	case abstraction.MsgTypeCommit:
		return "COMMIT"
	case abstraction.MsgTypeViewChange:
		return "VIEW_CHANGE"
	case abstraction.MsgTypeNewView:
		return "NEW_VIEW"
	default:
		return string(canonicalType)
	}
}

// BesuMessage represents the internal Besu message structure
type BesuMessage struct {
	BlockNumber    *big.Int  `json:"block_number"`
	RoundNumber    *big.Int  `json:"round_number"`
	Timestamp      time.Time `json:"timestamp"`
	Type           string    `json:"type"`
	BlockHash      string    `json:"block_hash,omitempty"`
	ParentHash     string    `json:"parent_hash,omitempty"`
	Proposer       string    `json:"proposer,omitempty"`
	Validator      string    `json:"validator,omitempty"`
	Signature      string    `json:"signature,omitempty"`
	GasLimit       uint64    `json:"gas_limit,omitempty"`
	GasUsed        uint64    `json:"gas_used,omitempty"`
	TxCount        int       `json:"tx_count,omitempty"`
	ValidatorCount int       `json:"validator_count,omitempty"`
	ConsensusType  string    `json:"consensus_type,omitempty"`
}
