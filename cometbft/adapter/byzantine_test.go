package adapter

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"codec/message/abstraction"
)

func TestApplyByzantineCanonical(t *testing.T) {
	voteTimestamp := time.Unix(1700000000, 0).UTC()
	proposalTimestamp := time.Unix(1700005000, 0).UTC()

	voteCanonical := &abstraction.CanonicalMessage{
		ChainID:   "test-chain",
		Height:    big.NewInt(5),
		Round:     big.NewInt(1),
		Timestamp: voteTimestamp,
		Type:      abstraction.MsgTypePrevote,
		BlockHash: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Validator: "validator-1",
		Signature: "sig-1",
	}

	proposalCanonical := &abstraction.CanonicalMessage{
		ChainID:   "test-chain",
		Height:    big.NewInt(10),
		Round:     big.NewInt(0),
		Timestamp: proposalTimestamp,
		Type:      abstraction.MsgTypeProposal,
		BlockHash: "1111111111111111111111111111111111111111111111111111111111111111",
		PrevHash:  "2222222222222222222222222222222222222222222222222222222222222222",
		Proposer:  "proposer-1",
		Signature: "sig-1",
	}

	tests := []struct {
		name    string
		msg     *abstraction.CanonicalMessage
		action  ByzantineAction
		opts    ByzantineOptions
		wantLen int
		wantErr bool
		assert  func(t *testing.T, canonicals []*abstraction.CanonicalMessage)
	}{
		{
			name:   "double vote with overrides",
			msg:    voteCanonical,
			action: ByzantineActionDoubleVote,
			opts: ByzantineOptions{
				AlternateBlockHash: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
				AlternateSignature: "sig-2",
				RoundOffset:        1,
				HeightOffset:       2,
				TimestampShift:     2 * time.Millisecond,
			},
			wantLen: 2,
			assert: func(t *testing.T, canonicals []*abstraction.CanonicalMessage) {
				if canonicals[0] == voteCanonical {
					t.Fatalf("expected cloned canonical message, not original pointer")
				}
				if canonicals[0].BlockHash != voteCanonical.BlockHash {
					t.Fatalf("expected original block hash to remain unchanged, got %s", canonicals[0].BlockHash)
				}
				if canonicals[0].Signature != voteCanonical.Signature {
					t.Fatalf("expected original signature to remain unchanged")
				}

				mutated := canonicals[1]
				if mutated.BlockHash != "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" {
					t.Fatalf("expected alternate block hash, got %s", mutated.BlockHash)
				}
				if mutated.Signature != "sig-2" {
					t.Fatalf("expected alternate signature, got %s", mutated.Signature)
				}
				expectedRound := new(big.Int).Add(voteCanonical.Round, big.NewInt(1))
				if mutated.Round == nil || mutated.Round.Cmp(expectedRound) != 0 {
					t.Fatalf("expected round %s, got %v", expectedRound, mutated.Round)
				}
				expectedHeight := new(big.Int).Add(voteCanonical.Height, big.NewInt(2))
				if mutated.Height == nil || mutated.Height.Cmp(expectedHeight) != 0 {
					t.Fatalf("expected height %s, got %v", expectedHeight, mutated.Height)
				}
				expectedTimestamp := voteCanonical.Timestamp.Add(2 * time.Millisecond)
				if !mutated.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, mutated.Timestamp)
				}
			},
		},
		{
			name:   "double proposal with overrides",
			msg:    proposalCanonical,
			action: ByzantineActionDoubleProposal,
			opts: ByzantineOptions{
				AlternateBlockHash: "3333333333333333333333333333333333333333333333333333333333333333",
				AlternatePrevHash:  "4444444444444444444444444444444444444444444444444444444444444444",
				AlternateSignature: "sig-2",
				TimestampShift:     time.Millisecond,
			},
			wantLen: 2,
			assert: func(t *testing.T, canonicals []*abstraction.CanonicalMessage) {
				mutated := canonicals[1]
				if mutated.BlockHash != "3333333333333333333333333333333333333333333333333333333333333333" {
					t.Fatalf("expected alternate block hash, got %s", mutated.BlockHash)
				}
				if mutated.PrevHash != "4444444444444444444444444444444444444444444444444444444444444444" {
					t.Fatalf("expected alternate prev hash, got %s", mutated.PrevHash)
				}
				if mutated.Signature != "sig-2" {
					t.Fatalf("expected alternate signature, got %s", mutated.Signature)
				}
				expectedTimestamp := proposalCanonical.Timestamp.Add(1 * time.Millisecond)
				if !mutated.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, mutated.Timestamp)
				}
			},
		},
		{
			name:    "alter validator on vote",
			msg:     voteCanonical,
			action:  ByzantineActionAlterValidator,
			opts:    ByzantineOptions{AlternateValidator: "validator-2"},
			wantLen: 1,
			assert: func(t *testing.T, canonicals []*abstraction.CanonicalMessage) {
				mutated := canonicals[0]
				if mutated.Validator != "validator-2" {
					t.Fatalf("expected validator override, got %s", mutated.Validator)
				}
				if !mutated.Timestamp.After(voteCanonical.Timestamp) {
					t.Fatalf("expected timestamp to advance for altered validator action")
				}
			},
		},
		{
			name:    "alter validator on proposal",
			msg:     proposalCanonical,
			action:  ByzantineActionAlterValidator,
			opts:    ByzantineOptions{AlternateValidator: "proposer-2"},
			wantLen: 1,
			assert: func(t *testing.T, canonicals []*abstraction.CanonicalMessage) {
				mutated := canonicals[0]
				if mutated.Proposer != "proposer-2" {
					t.Fatalf("expected proposer override, got %s", mutated.Proposer)
				}
				if !mutated.Timestamp.After(proposalCanonical.Timestamp) {
					t.Fatalf("expected timestamp to advance for altered proposer action")
				}
			},
		},
		{
			name:   "drop signature with offsets",
			msg:    voteCanonical,
			action: ByzantineActionDropSignature,
			opts: ByzantineOptions{
				RoundOffset:    2,
				HeightOffset:   -1,
				TimestampShift: 5 * time.Millisecond,
			},
			wantLen: 1,
			assert: func(t *testing.T, canonicals []*abstraction.CanonicalMessage) {
				mutated := canonicals[0]
				if mutated.Signature != "" {
					t.Fatalf("expected signature to be dropped, got %s", mutated.Signature)
				}
				expectedRound := new(big.Int).Add(voteCanonical.Round, big.NewInt(2))
				if mutated.Round == nil || mutated.Round.Cmp(expectedRound) != 0 {
					t.Fatalf("expected round %s, got %v", expectedRound, mutated.Round)
				}
				expectedHeight := new(big.Int).Add(voteCanonical.Height, big.NewInt(-1))
				if mutated.Height == nil || mutated.Height.Cmp(expectedHeight) != 0 {
					t.Fatalf("expected height %s, got %v", expectedHeight, mutated.Height)
				}
				expectedTimestamp := voteCanonical.Timestamp.Add(5 * time.Millisecond)
				if !mutated.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, mutated.Timestamp)
				}
			},
		},
		{
			name:    "timestamp skew",
			msg:     proposalCanonical,
			action:  ByzantineActionTimestampSkew,
			opts:    ByzantineOptions{TimestampShift: -3 * time.Second},
			wantLen: 1,
			assert: func(t *testing.T, canonicals []*abstraction.CanonicalMessage) {
				mutated := canonicals[0]
				expectedTimestamp := proposalCanonical.Timestamp.Add(-3 * time.Second)
				if !mutated.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, mutated.Timestamp)
				}
			},
		},
		{
			name:    "double vote wrong type",
			msg:     proposalCanonical,
			action:  ByzantineActionDoubleVote,
			wantErr: true,
		},
		{
			name:    "alter validator missing override",
			msg:     voteCanonical,
			action:  ByzantineActionAlterValidator,
			wantErr: true,
		},
		{
			name:    "timestamp skew missing duration",
			msg:     voteCanonical,
			action:  ByzantineActionTimestampSkew,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			canonicals, err := ApplyByzantineCanonical(tc.msg, tc.action, tc.opts)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("ApplyByzantineCanonical returned error: %v", err)
			}
			if len(canonicals) != tc.wantLen {
				t.Fatalf("expected %d canonical messages, got %d", tc.wantLen, len(canonicals))
			}
			if tc.assert != nil {
				tc.assert(t, canonicals)
			}
		})
	}
}

func TestFromCanonicalByzantine(t *testing.T) {
	mapper := NewCometBFTMapper("test-chain")
	voteTimestamp := time.Unix(1700000000, 0).UTC()
	proposalTimestamp := time.Unix(1700005000, 0).UTC()

	voteCanonical := &abstraction.CanonicalMessage{
		ChainID:   "test-chain",
		Height:    big.NewInt(5),
		Round:     big.NewInt(1),
		Timestamp: voteTimestamp,
		Type:      abstraction.MsgTypePrevote,
		BlockHash: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Validator: "validator-1",
		Signature: "sig-1",
	}

	proposalCanonical := &abstraction.CanonicalMessage{
		ChainID:   "test-chain",
		Height:    big.NewInt(10),
		Round:     big.NewInt(0),
		Timestamp: proposalTimestamp,
		Type:      abstraction.MsgTypeProposal,
		BlockHash: "1111111111111111111111111111111111111111111111111111111111111111",
		PrevHash:  "2222222222222222222222222222222222222222222222222222222222222222",
		Proposer:  "proposer-1",
		Signature: "sig-1",
	}

	tests := []struct {
		name    string
		msg     *abstraction.CanonicalMessage
		action  ByzantineAction
		opts    ByzantineOptions
		wantLen int
		assert  func(t *testing.T, raws []*abstraction.RawConsensusMessage)
	}{
		{
			name:   "double vote encoding",
			msg:    voteCanonical,
			action: ByzantineActionDoubleVote,
			opts: ByzantineOptions{
				AlternateBlockHash: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
				AlternateSignature: "sig-2",
				RoundOffset:        1,
				HeightOffset:       2,
				TimestampShift:     2 * time.Millisecond,
			},
			wantLen: 2,
			assert: func(t *testing.T, raws []*abstraction.RawConsensusMessage) {
				if len(raws) != 2 {
					t.Fatalf("expected 2 raw messages, got %d", len(raws))
				}
				var first, second CometBFTConsensusMessage
				if err := json.Unmarshal(raws[0].Payload, &first); err != nil {
					t.Fatalf("failed to unmarshal first payload: %v", err)
				}
				if err := json.Unmarshal(raws[1].Payload, &second); err != nil {
					t.Fatalf("failed to unmarshal second payload: %v", err)
				}
				if first.BlockID.Hash != voteCanonical.BlockHash {
					t.Fatalf("expected original block hash, got %s", first.BlockID.Hash)
				}
				if second.BlockID.Hash != "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" {
					t.Fatalf("expected alternate block hash, got %s", second.BlockID.Hash)
				}
				if second.Signature != "sig-2" {
					t.Fatalf("expected alternate signature, got %s", second.Signature)
				}
				if first.Height != "5" {
					t.Fatalf("expected original height 5, got %s", first.Height)
				}
				if second.Height != "7" {
					t.Fatalf("expected mutated height 7, got %s", second.Height)
				}
				if first.Round != "1" {
					t.Fatalf("expected original round 1, got %s", first.Round)
				}
				if second.Round != "2" {
					t.Fatalf("expected mutated round 2, got %s", second.Round)
				}
				expectedTimestamp := voteCanonical.Timestamp.Add(2 * time.Millisecond)
				if !second.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, second.Timestamp)
				}
			},
		},
		{
			name:   "double proposal encoding",
			msg:    proposalCanonical,
			action: ByzantineActionDoubleProposal,
			opts: ByzantineOptions{
				AlternateBlockHash: "3333333333333333333333333333333333333333333333333333333333333333",
				AlternatePrevHash:  "4444444444444444444444444444444444444444444444444444444444444444",
				AlternateSignature: "sig-2",
				TimestampShift:     time.Millisecond,
			},
			wantLen: 2,
			assert: func(t *testing.T, raws []*abstraction.RawConsensusMessage) {
				var first, second CometBFTConsensusMessage
				if err := json.Unmarshal(raws[0].Payload, &first); err != nil {
					t.Fatalf("failed to unmarshal first payload: %v", err)
				}
				if err := json.Unmarshal(raws[1].Payload, &second); err != nil {
					t.Fatalf("failed to unmarshal second payload: %v", err)
				}
				if second.BlockID.Hash != "3333333333333333333333333333333333333333333333333333333333333333" {
					t.Fatalf("expected alternate block hash, got %s", second.BlockID.Hash)
				}
				if second.BlockID.PrevHash != "4444444444444444444444444444444444444444444444444444444444444444" {
					t.Fatalf("expected alternate prev hash, got %s", second.BlockID.PrevHash)
				}
				if second.Signature != "sig-2" {
					t.Fatalf("expected alternate signature, got %s", second.Signature)
				}
				expectedTimestamp := proposalCanonical.Timestamp.Add(1 * time.Millisecond)
				if !second.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, second.Timestamp)
				}
			},
		},
		{
			name:    "alter validator encoding",
			msg:     voteCanonical,
			action:  ByzantineActionAlterValidator,
			opts:    ByzantineOptions{AlternateValidator: "validator-2"},
			wantLen: 1,
			assert: func(t *testing.T, raws []*abstraction.RawConsensusMessage) {
				var message CometBFTConsensusMessage
				if err := json.Unmarshal(raws[0].Payload, &message); err != nil {
					t.Fatalf("failed to unmarshal payload: %v", err)
				}
				if message.ValidatorAddress != "validator-2" {
					t.Fatalf("expected validator override, got %s", message.ValidatorAddress)
				}
			},
		},
		{
			name:    "alter proposer encoding",
			msg:     proposalCanonical,
			action:  ByzantineActionAlterValidator,
			opts:    ByzantineOptions{AlternateValidator: "proposer-2"},
			wantLen: 1,
			assert: func(t *testing.T, raws []*abstraction.RawConsensusMessage) {
				var message CometBFTConsensusMessage
				if err := json.Unmarshal(raws[0].Payload, &message); err != nil {
					t.Fatalf("failed to unmarshal payload: %v", err)
				}
				if message.ProposerAddress != "proposer-2" {
					t.Fatalf("expected proposer override, got %s", message.ProposerAddress)
				}
			},
		},
		{
			name:   "drop signature encoding",
			msg:    voteCanonical,
			action: ByzantineActionDropSignature,
			opts: ByzantineOptions{
				RoundOffset:    2,
				HeightOffset:   -1,
				TimestampShift: 5 * time.Millisecond,
			},
			wantLen: 1,
			assert: func(t *testing.T, raws []*abstraction.RawConsensusMessage) {
				var message CometBFTConsensusMessage
				if err := json.Unmarshal(raws[0].Payload, &message); err != nil {
					t.Fatalf("failed to unmarshal payload: %v", err)
				}
				if message.Signature != "" {
					t.Fatalf("expected signature to be dropped, got %s", message.Signature)
				}
				if message.Height != "4" {
					t.Fatalf("expected mutated height 4, got %s", message.Height)
				}
				if message.Round != "3" {
					t.Fatalf("expected mutated round 3, got %s", message.Round)
				}
				expectedTimestamp := voteCanonical.Timestamp.Add(5 * time.Millisecond)
				if !message.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, message.Timestamp)
				}
			},
		},
		{
			name:    "timestamp skew encoding",
			msg:     proposalCanonical,
			action:  ByzantineActionTimestampSkew,
			opts:    ByzantineOptions{TimestampShift: -3 * time.Second},
			wantLen: 1,
			assert: func(t *testing.T, raws []*abstraction.RawConsensusMessage) {
				var message CometBFTConsensusMessage
				if err := json.Unmarshal(raws[0].Payload, &message); err != nil {
					t.Fatalf("failed to unmarshal payload: %v", err)
				}
				expectedTimestamp := proposalCanonical.Timestamp.Add(-3 * time.Second)
				if !message.Timestamp.Equal(expectedTimestamp) {
					t.Fatalf("expected timestamp %s, got %s", expectedTimestamp, message.Timestamp)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raws, err := mapper.FromCanonicalByzantine(tc.msg, tc.action, tc.opts)
			if err != nil {
				t.Fatalf("FromCanonicalByzantine returned error: %v", err)
			}
			if len(raws) != tc.wantLen {
				t.Fatalf("expected %d raw messages, got %d", tc.wantLen, len(raws))
			}
			if tc.assert != nil {
				tc.assert(t, raws)
			}
		})
	}
}
