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
)

// ByzantineOptions contains optional overrides for the mutated messages.
type ByzantineOptions struct {
	AlternateBlockHash string
	AlternatePrevHash  string
	AlternateSignature string
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

	base := cloneCanonicalMessage(msg)

	switch action {
	case ByzantineActionNone:
		return []*abstraction.CanonicalMessage{base}, nil
	case ByzantineActionDoubleVote:
		if msg.Type != abstraction.MsgTypePrevote && msg.Type != abstraction.MsgTypePrecommit && msg.Type != abstraction.MsgTypeVote {
			return nil, fmt.Errorf("double_vote action requires a vote canonical message")
		}

		mutated := cloneCanonicalMessage(msg)
		mutated.BlockHash = chooseAlternateHash(msg.BlockHash, opts.AlternateBlockHash)
		if opts.AlternateSignature != "" {
			mutated.Signature = opts.AlternateSignature
		}
		if mutated.Timestamp.Equal(msg.Timestamp) {
			mutated.Timestamp = mutated.Timestamp.Add(1 * time.Millisecond)
		}

		return []*abstraction.CanonicalMessage{base, mutated}, nil
	case ByzantineActionDoubleProposal:
		if msg.Type != abstraction.MsgTypeProposal {
			return nil, fmt.Errorf("double_proposal action requires a proposal canonical message")
		}

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
		if mutated.Timestamp.Equal(msg.Timestamp) {
			mutated.Timestamp = mutated.Timestamp.Add(1 * time.Millisecond)
		}

		return []*abstraction.CanonicalMessage{base, mutated}, nil
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
