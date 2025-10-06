package cometbft

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"codec/message/abstraction"
)

// ValidatorSet represents a CometBFT validator set
type ValidatorSet struct {
	Validators []Validator `json:"validators"`
	Proposer   Validator   `json:"proposer"`
	TotalPower int64       `json:"total_power"`
}

// Validator represents a CometBFT validator
type Validator struct {
	Address          string `json:"address"`
	PubKey           string `json:"pub_key"`
	VotingPower      int64  `json:"voting_power"`
	ProposerPriority int64  `json:"proposer_priority"`
}

// ConsensusState represents the current consensus state
type ConsensusState struct {
	Height           int64        `json:"height"`
	Round            int32        `json:"round"`
	Step             uint32       `json:"step"`
	StartTime        time.Time    `json:"start_time"`
	Validators       ValidatorSet `json:"validators"`
	LastCommitRound  int32        `json:"last_commit_round"`
	LastCommitHeight int64        `json:"last_commit_height"`
}

// ConsensusEngine simulates CometBFT consensus engine behavior
type ConsensusEngine struct {
	state      ConsensusState
	validators map[string]Validator
	proposer   string
}

// NewConsensusEngine creates a new CometBFT consensus engine
func NewConsensusEngine(validators []Validator) *ConsensusEngine {
	validatorMap := make(map[string]Validator)
	var totalPower int64
	var proposer Validator

	for _, val := range validators {
		validatorMap[val.Address] = val
		totalPower += val.VotingPower
		if val.ProposerPriority > proposer.ProposerPriority {
			proposer = val
		}
	}

	return &ConsensusEngine{
		state: ConsensusState{
			Height:           0,
			Round:            0,
			Step:             0,
			StartTime:        time.Now(),
			Validators:       ValidatorSet{Validators: validators, Proposer: proposer, TotalPower: totalPower},
			LastCommitRound:  -1,
			LastCommitHeight: -1,
		},
		validators: validatorMap,
		proposer:   proposer.Address,
	}
}

// ProcessMessage processes a consensus message and updates state
func (ce *ConsensusEngine) ProcessMessage(msg *abstraction.CanonicalMessage) error {
	switch msg.Type {
	case abstraction.MsgTypeProposal:
		return ce.processProposal(msg)
	case abstraction.MsgTypePrevote:
		return ce.processPrevote(msg)
	case abstraction.MsgTypePrecommit:
		return ce.processPrecommit(msg)
	case abstraction.MsgTypeBlock:
		return ce.processBlockPart(msg)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

// processProposal processes a proposal message
func (ce *ConsensusEngine) processProposal(msg *abstraction.CanonicalMessage) error {
	// Validate proposer
	if msg.Proposer != ce.proposer {
		return fmt.Errorf("invalid proposer: expected %s, got %s", ce.proposer, msg.Proposer)
	}

	// Validate height and round
	if msg.Height.Cmp(big.NewInt(ce.state.Height)) != 0 {
		return fmt.Errorf("invalid height: expected %d, got %v", ce.state.Height, msg.Height)
	}

	if msg.Round.Cmp(big.NewInt(int64(ce.state.Round))) != 0 {
		return fmt.Errorf("invalid round: expected %d, got %v", ce.state.Round, msg.Round)
	}

	// Update state
	ce.state.Step = 1 // Propose step
	ce.state.StartTime = msg.Timestamp

	fmt.Printf("✅ Proposal processed: height=%v, round=%v, proposer=%s\n",
		msg.Height, msg.Round, msg.Proposer)

	return nil
}

// processPrevote processes a prevote message
func (ce *ConsensusEngine) processPrevote(msg *abstraction.CanonicalMessage) error {
	// Validate validator
	validator, exists := ce.validators[msg.Validator]
	if !exists {
		return fmt.Errorf("unknown validator: %s", msg.Validator)
	}

	// Validate height and round
	if msg.Height.Cmp(big.NewInt(ce.state.Height)) != 0 {
		return fmt.Errorf("invalid height: expected %d, got %v", ce.state.Height, msg.Height)
	}

	if msg.Round.Cmp(big.NewInt(int64(ce.state.Round))) != 0 {
		return fmt.Errorf("invalid round: expected %d, got %v", ce.state.Round, msg.Round)
	}

	// Update state
	if ce.state.Step < 2 {
		ce.state.Step = 2 // Prevote step
	}

	fmt.Printf("✅ Prevote processed: height=%v, round=%v, validator=%s, power=%d\n",
		msg.Height, msg.Round, msg.Validator, validator.VotingPower)

	return nil
}

// processPrecommit processes a precommit message
func (ce *ConsensusEngine) processPrecommit(msg *abstraction.CanonicalMessage) error {
	// Validate validator
	validator, exists := ce.validators[msg.Validator]
	if !exists {
		return fmt.Errorf("unknown validator: %s", msg.Validator)
	}

	// Validate height and round
	if msg.Height.Cmp(big.NewInt(ce.state.Height)) != 0 {
		return fmt.Errorf("invalid height: expected %d, got %v", ce.state.Height, msg.Height)
	}

	if msg.Round.Cmp(big.NewInt(int64(ce.state.Round))) != 0 {
		return fmt.Errorf("invalid round: expected %d, got %v", ce.state.Round, msg.Round)
	}

	// Update state
	if ce.state.Step < 3 {
		ce.state.Step = 3 // Precommit step
	}

	fmt.Printf("✅ Precommit processed: height=%v, round=%v, validator=%s, power=%d\n",
		msg.Height, msg.Round, msg.Validator, validator.VotingPower)

	return nil
}

// processBlockPart processes a block part message
func (ce *ConsensusEngine) processBlockPart(msg *abstraction.CanonicalMessage) error {
	// Validate height and round
	if msg.Height.Cmp(big.NewInt(ce.state.Height)) != 0 {
		return fmt.Errorf("invalid height: expected %d, got %v", ce.state.Height, msg.Height)
	}

	if msg.Round.Cmp(big.NewInt(int64(ce.state.Round))) != 0 {
		return fmt.Errorf("invalid round: expected %d, got %v", ce.state.Round, msg.Round)
	}

	fmt.Printf("✅ BlockPart processed: height=%v, round=%v, block_hash=%s\n",
		msg.Height, msg.Round, msg.BlockHash)

	return nil
}

// GetState returns the current consensus state
func (ce *ConsensusEngine) GetState() ConsensusState {
	return ce.state
}

// GetCurrentHeight returns the current height
func (ce *ConsensusEngine) GetCurrentHeight() int64 {
	return ce.state.Height
}

// GetCurrentRound returns the current round
func (ce *ConsensusEngine) GetCurrentRound() int32 {
	return ce.state.Round
}

// IsConsensusReached checks if consensus has been reached
func (ce *ConsensusEngine) IsConsensusReached() bool {
	return ce.state.Step >= 3 // Precommit step indicates consensus
}

// AdvanceRound advances to the next round
func (ce *ConsensusEngine) AdvanceRound() {
	ce.state.Round++
	ce.state.Step = 0
	ce.state.StartTime = time.Now()

	// Update proposer (round-robin)
	ce.updateProposer()
}

// AdvanceHeight advances to the specified height
func (ce *ConsensusEngine) AdvanceHeight(height int64) {
	ce.state.Height = height
	ce.state.Round = 0
	ce.state.Step = 0
	ce.state.StartTime = time.Now()
	ce.state.LastCommitHeight = height - 1
	ce.state.LastCommitRound = ce.state.Round

	// Update proposer
	ce.updateProposer()
}

// updateProposer updates the proposer based on round-robin
func (ce *ConsensusEngine) updateProposer() {
	if len(ce.state.Validators.Validators) == 0 {
		return
	}

	proposerIndex := int(ce.state.Round) % len(ce.state.Validators.Validators)
	ce.proposer = ce.state.Validators.Validators[proposerIndex].Address
	ce.state.Validators.Proposer = ce.state.Validators.Validators[proposerIndex]
}

// GenerateBlockHash generates a deterministic block hash
func (ce *ConsensusEngine) GenerateBlockHash(height int64, round int32, proposer string) string {
	data := fmt.Sprintf("%d:%d:%s:%d", height, round, proposer, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ValidateMessage validates a consensus message
func (ce *ConsensusEngine) ValidateMessage(msg *abstraction.CanonicalMessage) error {
	// Basic validation
	if msg.Height == nil || msg.Round == nil {
		return fmt.Errorf("height and round are required")
	}

	if msg.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	// Height validation
	if msg.Height.Cmp(big.NewInt(ce.state.Height)) < 0 {
		return fmt.Errorf("message height %v is less than current height %d", msg.Height, ce.state.Height)
	}

	// Round validation
	if msg.Round.Cmp(big.NewInt(int64(ce.state.Round))) < 0 {
		return fmt.Errorf("message round %v is less than current round %d", msg.Round, ce.state.Round)
	}

	// Validator validation for vote messages
	if msg.Type == abstraction.MsgTypePrevote || msg.Type == abstraction.MsgTypePrecommit {
		if msg.Validator == "" {
			return fmt.Errorf("validator is required for vote messages")
		}

		if _, exists := ce.validators[msg.Validator]; !exists {
			return fmt.Errorf("unknown validator: %s", msg.Validator)
		}
	}

	// Proposer validation for proposal messages
	if msg.Type == abstraction.MsgTypeProposal {
		if msg.Proposer == "" {
			return fmt.Errorf("proposer is required for proposal messages")
		}

		if msg.Proposer != ce.proposer {
			return fmt.Errorf("invalid proposer: expected %s, got %s", ce.proposer, msg.Proposer)
		}
	}

	return nil
}

// GetValidatorPower returns the voting power of a validator
func (ce *ConsensusEngine) GetValidatorPower(address string) int64 {
	if validator, exists := ce.validators[address]; exists {
		return validator.VotingPower
	}
	return 0
}

// GetTotalPower returns the total voting power
func (ce *ConsensusEngine) GetTotalPower() int64 {
	return ce.state.Validators.TotalPower
}

// IsValidProposer checks if the given address is the valid proposer for current round
func (ce *ConsensusEngine) IsValidProposer(address string) bool {
	return address == ce.proposer
}
