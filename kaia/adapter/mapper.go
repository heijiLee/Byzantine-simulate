package adapter

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"codec/message/abstraction"
)

// KaiaMapper implements the Mapper interface for Kaia consensus messages
type KaiaMapper struct {
	chainID string
}

// NewKaiaMapper creates a new Kaia mapper
func NewKaiaMapper(chainID string) *KaiaMapper {
	return &KaiaMapper{
		chainID: chainID,
	}
}

// ToCanonical converts a Kaia raw message to canonical format
func (m *KaiaMapper) ToCanonical(raw abstraction.RawConsensusMessage) (*abstraction.CanonicalMessage, error) {
	if raw.ChainType != abstraction.ChainTypeKaia {
		return nil, abstraction.ErrChainMismatch
	}

	// Parse the payload based on encoding
	var kaiaMsg KaiaMessage
	switch raw.Encoding {
	case "json":
		if err := json.Unmarshal(raw.Payload, &kaiaMsg); err != nil {
			return nil, &abstraction.MessageValidationError{
				Field:   "payload",
				Message: fmt.Sprintf("failed to parse JSON: %v", err),
				Code:    "DECODE_FAILURE",
			}
		}
	case "rlp":
		// For RLP encoding, we'll use a simplified approach
		// In a real implementation, you'd use proper RLP decoding
		if err := json.Unmarshal(raw.Payload, &kaiaMsg); err != nil {
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
		Height:     kaiaMsg.BlockNumber,
		Round:      kaiaMsg.RoundNumber,
		Timestamp:  kaiaMsg.Timestamp,
		Type:       m.mapMessageType(kaiaMsg.Type),
		BlockHash:  kaiaMsg.BlockHash,
		PrevHash:   kaiaMsg.ParentHash,
		Proposer:   kaiaMsg.Proposer,
		Validator:  kaiaMsg.Validator,
		Signature:  kaiaMsg.Signature,
		RawPayload: raw.Payload,
		Extensions: map[string]interface{}{
			"gas_limit":       kaiaMsg.GasLimit,
			"gas_used":        kaiaMsg.GasUsed,
			"tx_count":        kaiaMsg.TxCount,
			"validator_count": kaiaMsg.ValidatorCount,
			"consensus_type":  kaiaMsg.ConsensusType,
			"governance_id":   kaiaMsg.GovernanceID,
		},
	}

	return canonical, nil
}

// FromCanonical converts a canonical message to Kaia format
func (m *KaiaMapper) FromCanonical(msg *abstraction.CanonicalMessage) (*abstraction.RawConsensusMessage, error) {
	if msg == nil {
		return nil, &abstraction.MessageValidationError{
			Field:   "message",
			Message: "message cannot be nil",
			Code:    "MISSING_FIELD",
		}
	}

	// Extract Kaia-specific extensions
	gasLimit := uint64(0)
	gasUsed := uint64(0)
	txCount := 0
	validatorCount := 0
	consensusType := "istanbul"
	governanceID := ""

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
		if gi, ok := msg.Extensions["governance_id"].(string); ok {
			governanceID = gi
		}
	}

	// Convert canonical message to Kaia format
	kaiaMsg := KaiaMessage{
		BlockNumber:    msg.Height,
		RoundNumber:    msg.Round,
		Timestamp:      msg.Timestamp,
		Type:           m.mapToKaiaType(msg.Type),
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
		GovernanceID:   governanceID,
	}

	// Serialize to JSON (default format)
	payload, err := json.Marshal(kaiaMsg)
	if err != nil {
		return nil, &abstraction.MessageValidationError{
			Field:   "payload",
			Message: fmt.Sprintf("failed to serialize: %v", err),
			Code:    "DECODE_FAILURE",
		}
	}

	raw := &abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeKaia,
		ChainID:     m.chainID,
		MessageType: string(kaiaMsg.Type),
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"gas_limit":       kaiaMsg.GasLimit,
			"gas_used":        kaiaMsg.GasUsed,
			"tx_count":        kaiaMsg.TxCount,
			"validator_count": kaiaMsg.ValidatorCount,
			"consensus_type":  kaiaMsg.ConsensusType,
			"governance_id":   kaiaMsg.GovernanceID,
		},
	}

	return raw, nil
}

// GetSupportedTypes returns the message types supported by Kaia
func (m *KaiaMapper) GetSupportedTypes() []abstraction.MsgType {
	return []abstraction.MsgType{
		abstraction.MsgTypeProposal,
		abstraction.MsgTypeVote,
		abstraction.MsgTypeBlock,
	}
}

// GetChainType returns the chain type this mapper handles
func (m *KaiaMapper) GetChainType() abstraction.ChainType {
	return abstraction.ChainTypeKaia
}

// mapMessageType maps Kaia message types to canonical types
func (m *KaiaMapper) mapMessageType(kaiaType string) abstraction.MsgType {
	switch kaiaType {
	case "PROPOSAL":
		return abstraction.MsgTypeProposal
	case "VOTE":
		return abstraction.MsgTypeVote
	case "BLOCK":
		return abstraction.MsgTypeBlock
	default:
		return abstraction.MsgType(kaiaType)
	}
}

// mapToKaiaType maps canonical message types to Kaia types
func (m *KaiaMapper) mapToKaiaType(canonicalType abstraction.MsgType) string {
	switch canonicalType {
	case abstraction.MsgTypeProposal:
		return "PROPOSAL"
	case abstraction.MsgTypeVote:
		return "VOTE"
	case abstraction.MsgTypeBlock:
		return "BLOCK"
	default:
		return string(canonicalType)
	}
}

// KaiaMessage represents the internal Kaia message structure
type KaiaMessage struct {
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
	GovernanceID   string    `json:"governance_id,omitempty"`
}
