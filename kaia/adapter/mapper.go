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

	// Convert to canonical message based on IBFT structure
	canonical := &abstraction.CanonicalMessage{
		ChainID:    m.chainID,
		Height:     nil,
		Round:      nil,
		Timestamp:  time.Now(),
		Type:       m.mapMessageType(kaiaMsg.MessageType),
		BlockHash:  "",
		PrevHash:   "",
		Proposer:   "",
		Validator:  kaiaMsg.Validator,
		Signature:  kaiaMsg.CommittedSeal,
		RawPayload: raw.Payload,
		Extensions: map[string]interface{}{
			"kaia_message_type": kaiaMsg.MessageType,
			"timestamp":         kaiaMsg.Timestamp,
		},
	}

	// Extract data based on message type
	switch kaiaMsg.MessageType {
	case "Preprepare":
		if kaiaMsg.View != nil {
			canonical.Height = big.NewInt(kaiaMsg.View.Sequence)
			canonical.Round = big.NewInt(int64(kaiaMsg.View.Round))
		}
		if kaiaMsg.Proposal != nil {
			canonical.BlockHash = kaiaMsg.Proposal.Hash
			canonical.PrevHash = kaiaMsg.Proposal.ParentHash
			canonical.Proposer = "proposer" // 실제로는 validator에서 추출
			canonical.Extensions["proposal"] = kaiaMsg.Proposal
		}
	case "Prepare", "Commit", "RoundChange":
		if kaiaMsg.Subject != nil {
			if kaiaMsg.Subject.View != nil {
				canonical.Height = big.NewInt(kaiaMsg.Subject.View.Sequence)
				canonical.Round = big.NewInt(int64(kaiaMsg.Subject.View.Round))
			}
			canonical.BlockHash = kaiaMsg.Subject.Digest
			canonical.PrevHash = kaiaMsg.Subject.PrevHash
			canonical.Extensions["subject"] = kaiaMsg.Subject
		}
	}

	// Add consensus message wrapper info
	if kaiaMsg.ConsensusMsg != nil {
		canonical.Extensions["consensus_msg"] = kaiaMsg.ConsensusMsg
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

	// Extract Kaia-specific extensions if needed
	// (Currently not used in the new IBFT structure)

	// Convert canonical message to Kaia format
	kaiaMsg := KaiaMessage{
		MessageType:   m.mapToKaiaType(msg.Type),
		Validator:     msg.Validator,
		CommittedSeal: msg.Signature,
		Timestamp:     msg.Timestamp.Format(time.RFC3339),
	}

	// Add View based on message type
	if msg.Height != nil && msg.Round != nil {
		kaiaMsg.View = &KaiaView{
			Round:    int32(msg.Round.Int64()),
			Sequence: msg.Height.Int64(),
		}
	}

	// Add Subject for Prepare/Commit/RoundChange
	if msg.Type == abstraction.MsgTypeVote || msg.Type == abstraction.MsgTypeBlock {
		kaiaMsg.Subject = &KaiaSubject{
			View:     kaiaMsg.View,
			Digest:   msg.BlockHash,
			PrevHash: msg.PrevHash,
		}
	}

	// Add Proposal for Preprepare
	if msg.Type == abstraction.MsgTypeProposal {
		kaiaMsg.Proposal = &KaiaProposal{
			Number:     msg.Height.Int64(),
			Hash:       msg.BlockHash,
			ParentHash: msg.PrevHash,
			Timestamp:  msg.Timestamp.Unix(),
			GasLimit:   30000000,
			GasUsed:    15000000,
			ExtraData:  "kaia-ibft-consensus",
			MixHash:    fmt.Sprintf("0x%x", time.Now().UnixNano()),
			Nonce:      "0x0000000000000000",
			BaseFee:    "25000000000",
		}
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
		MessageType: kaiaMsg.MessageType,
		Payload:     payload,
		Encoding:    "rlp",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"kaia_message_type": kaiaMsg.MessageType,
			"timestamp":         kaiaMsg.Timestamp,
		},
	}

	return raw, nil
}

// GetSupportedTypes returns the message types supported by Kaia IBFT
func (m *KaiaMapper) GetSupportedTypes() []abstraction.MsgType {
	return []abstraction.MsgType{
		abstraction.MsgTypeProposal, // Preprepare
		abstraction.MsgTypeVote,     // Prepare, Commit
		abstraction.MsgTypeBlock,    // RoundChange
	}
}

// GetChainType returns the chain type this mapper handles
func (m *KaiaMapper) GetChainType() abstraction.ChainType {
	return abstraction.ChainTypeKaia
}

// mapMessageType maps Kaia IBFT message types to canonical types
func (m *KaiaMapper) mapMessageType(kaiaType string) abstraction.MsgType {
	switch kaiaType {
	case "Preprepare":
		return abstraction.MsgTypeProposal
	case "Prepare", "Commit":
		return abstraction.MsgTypeVote
	case "RoundChange":
		return abstraction.MsgTypeBlock
	default:
		return abstraction.MsgType(kaiaType)
	}
}

// mapToKaiaType maps canonical message types to Kaia IBFT types
func (m *KaiaMapper) mapToKaiaType(canonicalType abstraction.MsgType) string {
	switch canonicalType {
	case abstraction.MsgTypeProposal:
		return "Preprepare"
	case abstraction.MsgTypeVote:
		return "Prepare" // 기본값으로 Prepare 사용
	case abstraction.MsgTypeBlock:
		return "RoundChange"
	default:
		return string(canonicalType)
	}
}

// KaiaMessage represents the internal Kaia IBFT message structure
type KaiaMessage struct {
	MessageType   string            `json:"message_type"`
	View          *KaiaView         `json:"view,omitempty"`
	Subject       *KaiaSubject      `json:"subject,omitempty"`
	Proposal      *KaiaProposal     `json:"proposal,omitempty"`
	Validator     string            `json:"validator,omitempty"`
	CommittedSeal string            `json:"committed_seal,omitempty"`
	Timestamp     string            `json:"timestamp"`
	ConsensusMsg  *KaiaConsensusMsg `json:"consensus_msg,omitempty"`
}

// KaiaView represents IBFT View (Round, Sequence)
type KaiaView struct {
	Round    int32 `json:"round"`
	Sequence int64 `json:"sequence"`
}

// KaiaSubject represents IBFT Subject (Prepare/Commit/RoundChange 공통)
type KaiaSubject struct {
	View     *KaiaView `json:"view"`
	Digest   string    `json:"digest"`    // 제안 블록 해시
	PrevHash string    `json:"prev_hash"` // 부모 블록 해시
}

// KaiaProposal represents IBFT Proposal (블록 헤더 기반)
type KaiaProposal struct {
	Number     int64  `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parent_hash"`
	Timestamp  int64  `json:"timestamp"`
	GasLimit   int64  `json:"gas_limit"`
	GasUsed    int64  `json:"gas_used"`
	ExtraData  string `json:"extra_data"`
	MixHash    string `json:"mix_hash"`
	Nonce      string `json:"nonce"`
	BaseFee    string `json:"base_fee"`
}

// KaiaConsensusMsg represents the outer wrapper (PrevHash, Payload)
type KaiaConsensusMsg struct {
	PrevHash string `json:"prev_hash"`
	Payload  string `json:"payload"`
}
