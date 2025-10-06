package abstraction

import (
	"math/big"
	"time"
)

// ChainType represents the supported blockchain platforms
type ChainType string

const (
	ChainTypeCometBFT    ChainType = "cometbft"
	ChainTypeHyperledger ChainType = "hyperledger"
	ChainTypeKaia        ChainType = "kaia"
)

// MsgType represents consensus message types across different chains
type MsgType string

const (
	MsgTypeProposal   MsgType = "proposal"
	MsgTypePrepare    MsgType = "prepare"
	MsgTypeVote       MsgType = "vote"
	MsgTypeCommit     MsgType = "commit"
	MsgTypeViewChange MsgType = "view_change"
	MsgTypeNewView    MsgType = "new_view"
	MsgTypeBlock      MsgType = "block"
	MsgTypePrevote    MsgType = "prevote"
	MsgTypePrecommit  MsgType = "precommit"
)

// CanonicalMessage represents the normalized consensus message format
type CanonicalMessage struct {
	// Common header fields
	ChainID   string    `json:"chain_id"`        // Chain identifier
	Height    *big.Int  `json:"height"`          // Block height
	Round     *big.Int  `json:"round,omitempty"` // Consensus round
	View      *big.Int  `json:"view,omitempty"`  // View number (for PBFT-style protocols)
	Timestamp time.Time `json:"timestamp"`       // Message creation time
	Type      MsgType   `json:"type"`            // Message type

	// Consensus-specific fields
	BlockHash string `json:"block_hash,omitempty"` // Proposed block hash
	PrevHash  string `json:"prev_hash,omitempty"`  // Previous block hash
	Proposer  string `json:"proposer,omitempty"`   // Proposer node ID
	Validator string `json:"validator,omitempty"`  // Validator node ID
	Signature string `json:"signature,omitempty"`  // Message signature

	// Advanced consensus fields
	CommitSeals []string          `json:"commit_seals,omitempty"` // Commit signatures
	ViewChanges []ViewChangeEntry `json:"view_changes,omitempty"` // View change entries

	// Extension fields for chain-specific data
	Extensions map[string]interface{} `json:"extensions,omitempty"`

	// Metadata
	RawPayload []byte `json:"raw_payload,omitempty"` // Original message bytes
}

// RawConsensusMessage represents a chain-specific consensus message
type RawConsensusMessage struct {
	ChainType   ChainType `json:"chain_type"`   // Source chain type
	ChainID     string    `json:"chain_id"`     // Chain identifier
	MessageType string    `json:"message_type"` // Original message type name
	Payload     []byte    `json:"payload"`      // Serialized message data
	Encoding    string    `json:"encoding"`     // Encoding format (proto, json, rlp, etc.)
	Timestamp   time.Time `json:"timestamp"`    // Message reception time

	// Chain-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ViewChangeEntry represents a view change entry in PBFT-style protocols
type ViewChangeEntry struct {
	View      *big.Int `json:"view"`      // View number
	Height    *big.Int `json:"height"`    // Block height at that view
	Validator string   `json:"validator"` // Validator ID
	Signature string   `json:"signature"` // Validator signature
}

// Mapper interface for converting between chain-specific and canonical formats
type Mapper interface {
	// ToCanonical converts a raw consensus message to canonical format
	ToCanonical(raw RawConsensusMessage) (*CanonicalMessage, error)

	// FromCanonical converts a canonical message to chain-specific format
	FromCanonical(msg *CanonicalMessage) (*RawConsensusMessage, error)

	// GetSupportedTypes returns the message types supported by this mapper
	GetSupportedTypes() []MsgType

	// GetChainType returns the chain type this mapper handles
	GetChainType() ChainType
}

// ChainMetadata contains chain-specific configuration and metadata
type ChainMetadata struct {
	ChainType   ChainType              `json:"chain_type"`
	ChainID     string                 `json:"chain_id"`
	Endpoint    string                 `json:"endpoint"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Credentials map[string]string      `json:"credentials,omitempty"`
}

// MessageValidationError represents validation errors
type MessageValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *MessageValidationError) Error() string {
	return e.Message
}

// Validation errors
var (
	ErrMissingField     = &MessageValidationError{Code: "MISSING_FIELD", Message: "required field is missing"}
	ErrUnsupportedType  = &MessageValidationError{Code: "UNSUPPORTED_TYPE", Message: "unsupported message type"}
	ErrDecodeFailure    = &MessageValidationError{Code: "DECODE_FAILURE", Message: "failed to decode message"}
	ErrInvalidSignature = &MessageValidationError{Code: "INVALID_SIGNATURE", Message: "invalid signature"}
	ErrChainMismatch    = &MessageValidationError{Code: "CHAIN_MISMATCH", Message: "chain type mismatch"}
)
