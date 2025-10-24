package adapter

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"codec/message/abstraction"
)

// ByzantineAction describes the manipulation to apply when converting back to a CometBFT message.
type ByzantineAction string

const (
	// ByzantineActionNone returns the canonical message without any manipulation.
	ByzantineActionNone ByzantineAction = "none"
	// ByzantineActionDoubleVote emits two vote messages with conflicting block hashes.
	ByzantineActionDoubleVote ByzantineAction = "double_vote"
	// ByzantineActionDoubleProposal emits two proposal messages that reference different blocks.
	ByzantineActionDoubleProposal ByzantineAction = "double_proposal"
	// ByzantineActionAlterValidator rewrites the validator or proposer identity.
	ByzantineActionAlterValidator ByzantineAction = "alter_validator"
	// ByzantineActionDropSignature removes the signature from the message.
	ByzantineActionDropSignature ByzantineAction = "drop_signature"
	// ByzantineActionTimestampSkew applies a timestamp shift to the message.
	ByzantineActionTimestampSkew ByzantineAction = "timestamp_skew"
)

// ByzantineOptions contains optional overrides for the mutated messages.
type ByzantineOptions struct {
	AlternateBlockHash string
	AlternatePrevHash  string
	AlternateSignature string
	AlternateValidator string
	RoundOffset        int64
	HeightOffset       int64
	TimestampShift     time.Duration
}

// ParseByzantineAction converts a CLI string to the typed action.
func ParseByzantineAction(value string) (ByzantineAction, error) {
	switch strings.ToLower(value) {
	case "", string(ByzantineActionNone):
		return ByzantineActionNone, nil
	case string(ByzantineActionDoubleVote):
		return ByzantineActionDoubleVote, nil
	case string(ByzantineActionDoubleProposal):
		return ByzantineActionDoubleProposal, nil
	case string(ByzantineActionAlterValidator):
		return ByzantineActionAlterValidator, nil
	case string(ByzantineActionDropSignature):
		return ByzantineActionDropSignature, nil
	case string(ByzantineActionTimestampSkew):
		return ByzantineActionTimestampSkew, nil
	default:
		return ByzantineActionNone, fmt.Errorf("unknown byzantine action: %s", value)
	}
}

// ApplyByzantineCanonical mutates a canonical message according to the requested action.
// It returns the set of canonical messages that should subsequently be encoded.
func ApplyByzantineCanonical(msg *abstraction.CanonicalMessage, action ByzantineAction, opts ByzantineOptions) ([]*abstraction.CanonicalMessage, error) {
	if msg == nil {
		return nil, fmt.Errorf("canonical message cannot be nil")
	}

	switch action {
	case ByzantineActionNone:
		return []*abstraction.CanonicalMessage{cloneCanonicalMessage(msg)}, nil
	case ByzantineActionDoubleVote:
		return applyDoubleVoteMutation(msg, opts)
	case ByzantineActionDoubleProposal:
		return applyDoubleProposalMutation(msg, opts)
	case ByzantineActionAlterValidator:
		return applyAlterValidatorMutation(msg, opts)
	case ByzantineActionDropSignature:
		return applyDropSignatureMutation(msg, opts)
	case ByzantineActionTimestampSkew:
		return applyTimestampSkewMutation(msg, opts)
	default:
		return nil, fmt.Errorf("unsupported byzantine action: %s", action)
	}
}

// FromCanonicalByzantine converts a canonical message back to CometBFT format while applying a byzantine action.
func (m *CometBFTMapper) FromCanonicalByzantine(msg *abstraction.CanonicalMessage, action ByzantineAction, opts ByzantineOptions) ([]*abstraction.RawConsensusMessage, error) {
	canonicals, err := ApplyByzantineCanonical(msg, action, opts)
	if err != nil {
		return nil, err
	}

	raws := make([]*abstraction.RawConsensusMessage, len(canonicals))
	for i, canonical := range canonicals {
		raw, err := m.FromCanonical(canonical)
		if err != nil {
			return nil, err
		}
		raws[i] = raw
	}

	return raws, nil
}

func chooseAlternateHash(original, provided string) string {
	if provided != "" {
		return provided
	}
	if original == "" {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}

	runes := []rune(original)
	for i := len(runes) - 1; i >= 0; i-- {
		switch runes[i] {
		case '0':
			runes[i] = '1'
			return string(runes)
		case '1':
			runes[i] = '0'
			return string(runes)
		case 'a', 'A':
			runes[i] = 'b'
			return string(runes)
		case 'f', 'F':
			runes[i] = 'e'
			return string(runes)
		}
	}
	return original + "0"
}

func cloneCanonicalMessage(msg *abstraction.CanonicalMessage) *abstraction.CanonicalMessage {
	if msg == nil {
		return nil
	}
	cloned := *msg
	if msg.Height != nil {
		cloned.Height = new(big.Int).Set(msg.Height)
	}
	if msg.Round != nil {
		cloned.Round = new(big.Int).Set(msg.Round)
	}
	if msg.View != nil {
		cloned.View = new(big.Int).Set(msg.View)
	}
	if msg.CommitSeals != nil {
		cloned.CommitSeals = append([]string(nil), msg.CommitSeals...)
	}
	if msg.ViewChanges != nil {
		cloned.ViewChanges = make([]abstraction.ViewChangeEntry, len(msg.ViewChanges))
		for i, vc := range msg.ViewChanges {
			entry := vc
			if vc.View != nil {
				entry.View = new(big.Int).Set(vc.View)
			}
			if vc.Height != nil {
				entry.Height = new(big.Int).Set(vc.Height)
			}
			cloned.ViewChanges[i] = entry
		}
	}
	if msg.Extensions != nil {
		copied := make(map[string]interface{}, len(msg.Extensions))
		for k, v := range msg.Extensions {
			copied[k] = v
		}
		cloned.Extensions = copied
	}
	if msg.RawPayload != nil {
		cloned.RawPayload = append([]byte(nil), msg.RawPayload...)
	}
	return &cloned
}

func applyDoubleVoteMutation(msg *abstraction.CanonicalMessage, opts ByzantineOptions) ([]*abstraction.CanonicalMessage, error) {
	if msg.Type != abstraction.MsgTypePrevote && msg.Type != abstraction.MsgTypePrecommit && msg.Type != abstraction.MsgTypeVote {
		return nil, fmt.Errorf("double_vote action requires a vote canonical message")
	}

	original := cloneCanonicalMessage(msg)
	mutated := cloneCanonicalMessage(msg)
	mutated.BlockHash = chooseAlternateHash(msg.BlockHash, opts.AlternateBlockHash)
	if opts.AlternateSignature != "" {
		mutated.Signature = opts.AlternateSignature
	}

	applyCommonMutations(mutated, opts)
	ensureTimestampProgress(mutated, msg.Timestamp)

	return []*abstraction.CanonicalMessage{original, mutated}, nil
}

func applyDoubleProposalMutation(msg *abstraction.CanonicalMessage, opts ByzantineOptions) ([]*abstraction.CanonicalMessage, error) {
	if msg.Type != abstraction.MsgTypeProposal {
		return nil, fmt.Errorf("double_proposal action requires a proposal canonical message")
	}

	original := cloneCanonicalMessage(msg)
	mutated := cloneCanonicalMessage(msg)
	mutated.BlockHash = chooseAlternateHash(msg.BlockHash, opts.AlternateBlockHash)
	if opts.AlternatePrevHash != "" {
		mutated.PrevHash = opts.AlternatePrevHash
	} else if mutated.PrevHash == msg.PrevHash {
		mutated.PrevHash = chooseAlternateHash(msg.PrevHash, "")
	}
	if opts.AlternateSignature != "" {
		mutated.Signature = opts.AlternateSignature
	}

	applyCommonMutations(mutated, opts)
	ensureTimestampProgress(mutated, msg.Timestamp)

	return []*abstraction.CanonicalMessage{original, mutated}, nil
}

func applyAlterValidatorMutation(msg *abstraction.CanonicalMessage, opts ByzantineOptions) ([]*abstraction.CanonicalMessage, error) {
	if opts.AlternateValidator == "" {
		return nil, fmt.Errorf("alter_validator action requires AlternateValidator to be set")
	}

	mutated := cloneCanonicalMessage(msg)
	switch msg.Type {
	case abstraction.MsgTypePrevote, abstraction.MsgTypePrecommit, abstraction.MsgTypeVote:
		mutated.Validator = opts.AlternateValidator
	case abstraction.MsgTypeProposal:
		mutated.Proposer = opts.AlternateValidator
	default:
		return nil, fmt.Errorf("alter_validator action requires a proposal or vote canonical message")
	}

	applyCommonMutations(mutated, opts)
	ensureTimestampProgress(mutated, msg.Timestamp)

	return []*abstraction.CanonicalMessage{mutated}, nil
}

func applyDropSignatureMutation(msg *abstraction.CanonicalMessage, opts ByzantineOptions) ([]*abstraction.CanonicalMessage, error) {
	mutated := cloneCanonicalMessage(msg)
	mutated.Signature = ""

	applyCommonMutations(mutated, opts)
	ensureTimestampProgress(mutated, msg.Timestamp)

	return []*abstraction.CanonicalMessage{mutated}, nil
}

func applyTimestampSkewMutation(msg *abstraction.CanonicalMessage, opts ByzantineOptions) ([]*abstraction.CanonicalMessage, error) {
	if opts.TimestampShift == 0 {
		return nil, fmt.Errorf("timestamp_skew action requires TimestampShift to be non-zero")
	}

	mutated := cloneCanonicalMessage(msg)
	applyCommonMutations(mutated, opts)
	ensureTimestampProgress(mutated, msg.Timestamp)

	return []*abstraction.CanonicalMessage{mutated}, nil
}

func applyCommonMutations(target *abstraction.CanonicalMessage, opts ByzantineOptions) {
	if opts.HeightOffset != 0 {
		target.Height = shiftBigInt(target.Height, opts.HeightOffset)
	}
	if opts.RoundOffset != 0 {
		target.Round = shiftBigInt(target.Round, opts.RoundOffset)
	}
	if opts.TimestampShift != 0 {
		target.Timestamp = target.Timestamp.Add(opts.TimestampShift)
	}
}

func ensureTimestampProgress(mutated *abstraction.CanonicalMessage, original time.Time) {
	if original.IsZero() {
		return
	}
	if mutated.Timestamp.Equal(original) {
		mutated.Timestamp = mutated.Timestamp.Add(1 * time.Millisecond)
	}
}

func shiftBigInt(value *big.Int, offset int64) *big.Int {
	if offset == 0 {
		if value == nil {
			return nil
		}
		return new(big.Int).Set(value)
	}

	if value == nil {
		return big.NewInt(offset)
	}
	return new(big.Int).Add(value, big.NewInt(offset))
}
