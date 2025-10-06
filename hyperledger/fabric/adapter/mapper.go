package adapter

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"codec/message/abstraction"
)

// FabricMapper implements the Mapper interface for Hyperledger Fabric consensus messages
type FabricMapper struct {
	chainID string
}

// NewFabricMapper creates a new Fabric mapper
func NewFabricMapper(chainID string) *FabricMapper {
	return &FabricMapper{
		chainID: chainID,
	}
}

// ToCanonical converts a Fabric raw message to canonical format
func (m *FabricMapper) ToCanonical(raw abstraction.RawConsensusMessage) (*abstraction.CanonicalMessage, error) {
	if raw.ChainType != abstraction.ChainTypeHyperledger {
		return nil, abstraction.ErrChainMismatch
	}

	// Parse the payload based on encoding
	var fabricMsg FabricMessage
	switch raw.Encoding {
	case "json":
		if err := json.Unmarshal(raw.Payload, &fabricMsg); err != nil {
			return nil, &abstraction.MessageValidationError{
				Field:   "payload",
				Message: fmt.Sprintf("failed to parse JSON: %v", err),
				Code:    "DECODE_FAILURE",
			}
		}
	case "proto":
		// For protobuf, we'll parse the basic structure
		// In a real implementation, you'd use the actual Fabric protobuf definitions
		if err := json.Unmarshal(raw.Payload, &fabricMsg); err != nil {
			return nil, &abstraction.MessageValidationError{
				Field:   "payload",
				Message: fmt.Sprintf("failed to parse protobuf: %v", err),
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
		Height:     fabricMsg.BlockNumber,
		View:       fabricMsg.ViewNumber,
		Timestamp:  fabricMsg.Timestamp,
		Type:       m.mapMessageType(fabricMsg.Type),
		BlockHash:  fabricMsg.BlockHash,
		PrevHash:   fabricMsg.PrevHash,
		Proposer:   fabricMsg.Proposer,
		Validator:  fabricMsg.Endorser,
		Signature:  fabricMsg.Signature,
		RawPayload: raw.Payload,
		Extensions: map[string]interface{}{
			"channel_id":     fabricMsg.ChannelID,
			"tx_count":       fabricMsg.TxCount,
			"endorser_count": fabricMsg.EndorserCount,
			"chaincode_id":   fabricMsg.ChaincodeID,
		},
	}

	return canonical, nil
}

// FromCanonical converts a canonical message to Fabric format
func (m *FabricMapper) FromCanonical(msg *abstraction.CanonicalMessage) (*abstraction.RawConsensusMessage, error) {
	if msg == nil {
		return nil, &abstraction.MessageValidationError{
			Field:   "message",
			Message: "message cannot be nil",
			Code:    "MISSING_FIELD",
		}
	}

	// Extract Fabric-specific extensions
	channelID := ""
	txCount := 0
	endorserCount := 0
	chaincodeID := ""

	if msg.Extensions != nil {
		if ch, ok := msg.Extensions["channel_id"].(string); ok {
			channelID = ch
		}
		if tc, ok := msg.Extensions["tx_count"].(int); ok {
			txCount = tc
		}
		if ec, ok := msg.Extensions["endorser_count"].(int); ok {
			endorserCount = ec
		}
		if cc, ok := msg.Extensions["chaincode_id"].(string); ok {
			chaincodeID = cc
		}
	}

	// Convert canonical message to Fabric format
	fabricMsg := FabricMessage{
		BlockNumber:   msg.Height,
		ViewNumber:    msg.View,
		Timestamp:     msg.Timestamp,
		Type:          m.mapToFabricType(msg.Type),
		BlockHash:     msg.BlockHash,
		PrevHash:      msg.PrevHash,
		Proposer:      msg.Proposer,
		Endorser:      msg.Validator,
		Signature:     msg.Signature,
		ChannelID:     channelID,
		TxCount:       txCount,
		EndorserCount: endorserCount,
		ChaincodeID:   chaincodeID,
	}

	// Serialize to JSON (default format)
	payload, err := json.Marshal(fabricMsg)
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
		MessageType: string(fabricMsg.Type),
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"channel_id":     fabricMsg.ChannelID,
			"tx_count":       fabricMsg.TxCount,
			"endorser_count": fabricMsg.EndorserCount,
			"chaincode_id":   fabricMsg.ChaincodeID,
		},
	}

	return raw, nil
}

// GetSupportedTypes returns the message types supported by Fabric
func (m *FabricMapper) GetSupportedTypes() []abstraction.MsgType {
	return []abstraction.MsgType{
		abstraction.MsgTypeProposal,
		abstraction.MsgTypePrepare,
		abstraction.MsgTypeCommit,
		abstraction.MsgTypeViewChange,
		abstraction.MsgTypeNewView,
	}
}

// GetChainType returns the chain type this mapper handles
func (m *FabricMapper) GetChainType() abstraction.ChainType {
	return abstraction.ChainTypeHyperledger
}

// mapMessageType maps Fabric message types to canonical types
func (m *FabricMapper) mapMessageType(fabricType string) abstraction.MsgType {
	switch fabricType {
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
		return abstraction.MsgType(fabricType)
	}
}

// mapToFabricType maps canonical message types to Fabric types
func (m *FabricMapper) mapToFabricType(canonicalType abstraction.MsgType) string {
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

// FabricMessage represents the internal Fabric message structure
type FabricMessage struct {
	BlockNumber   *big.Int  `json:"block_number"`
	ViewNumber    *big.Int  `json:"view_number"`
	Timestamp     time.Time `json:"timestamp"`
	Type          string    `json:"type"`
	BlockHash     string    `json:"block_hash,omitempty"`
	PrevHash      string    `json:"prev_hash,omitempty"`
	Proposer      string    `json:"proposer,omitempty"`
	Endorser      string    `json:"endorser,omitempty"`
	Signature     string    `json:"signature,omitempty"`
	ChannelID     string    `json:"channel_id,omitempty"`
	TxCount       int       `json:"tx_count,omitempty"`
	EndorserCount int       `json:"endorser_count,omitempty"`
	ChaincodeID   string    `json:"chaincode_id,omitempty"`
}
