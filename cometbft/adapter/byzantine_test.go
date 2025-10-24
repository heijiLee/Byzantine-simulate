package adapter

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"codec/message/abstraction"
)

func TestDoubleVoteMessages(t *testing.T) {
	mapper := NewCometBFTMapper("test-chain")
	canonical := &abstraction.CanonicalMessage{
		ChainID:   "test-chain",
		Height:    big.NewInt(5),
		Round:     big.NewInt(1),
		Timestamp: time.Now().UTC(),
		Type:      abstraction.MsgTypePrevote,
		BlockHash: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Validator: "validator-1",
		Signature: "sig-1",
	}

	opts := ByzantineOptions{AlternateBlockHash: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"}
	raws, err := mapper.FromCanonicalByzantine(canonical, ByzantineActionDoubleVote, opts)
	if err != nil {
		t.Fatalf("double vote conversion failed: %v", err)
	}
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

	if first.MessageType != "Vote" || second.MessageType != "Vote" {
		t.Fatalf("expected vote message types, got %s and %s", first.MessageType, second.MessageType)
	}
	if first.BlockID.Hash != canonical.BlockHash {
		t.Fatalf("expected original hash in first vote, got %s", first.BlockID.Hash)
	}
	if second.BlockID.Hash != opts.AlternateBlockHash {
		t.Fatalf("expected alternate hash in second vote, got %s", second.BlockID.Hash)
	}
	if first.Type != 1 || second.Type != 1 {
		t.Fatalf("expected prevote type 1 for both messages, got %d and %d", first.Type, second.Type)
	}
}

func TestDoubleProposalMessages(t *testing.T) {
	mapper := NewCometBFTMapper("test-chain")
	canonical := &abstraction.CanonicalMessage{
		ChainID:   "test-chain",
		Height:    big.NewInt(10),
		Round:     big.NewInt(0),
		Timestamp: time.Now().UTC(),
		Type:      abstraction.MsgTypeProposal,
		BlockHash: "1111111111111111111111111111111111111111111111111111111111111111",
		PrevHash:  "2222222222222222222222222222222222222222222222222222222222222222",
		Proposer:  "proposer-1",
		Signature: "sig-1",
	}

	opts := ByzantineOptions{
		AlternateBlockHash: "3333333333333333333333333333333333333333333333333333333333333333",
		AlternatePrevHash:  "4444444444444444444444444444444444444444444444444444444444444444",
		AlternateSignature: "sig-2",
	}
	raws, err := mapper.FromCanonicalByzantine(canonical, ByzantineActionDoubleProposal, opts)
	if err != nil {
		t.Fatalf("double proposal conversion failed: %v", err)
	}
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

	if first.BlockID.Hash != canonical.BlockHash {
		t.Fatalf("expected original block hash, got %s", first.BlockID.Hash)
	}
	if first.BlockID.PrevHash != canonical.PrevHash {
		t.Fatalf("expected original prev hash, got %s", first.BlockID.PrevHash)
	}
	if first.Signature != canonical.Signature {
		t.Fatalf("expected original signature, got %s", first.Signature)
	}

	if second.BlockID.Hash != opts.AlternateBlockHash {
		t.Fatalf("expected alternate block hash, got %s", second.BlockID.Hash)
	}
	if second.BlockID.PrevHash != opts.AlternatePrevHash {
		t.Fatalf("expected alternate prev hash, got %s", second.BlockID.PrevHash)
	}
	if second.Signature != opts.AlternateSignature {
		t.Fatalf("expected alternate signature, got %s", second.Signature)
	}
}

func TestInvalidByzantineActionOnProposal(t *testing.T) {
	mapper := NewCometBFTMapper("test-chain")
	canonical := &abstraction.CanonicalMessage{
		ChainID:   "test-chain",
		Height:    big.NewInt(10),
		Round:     big.NewInt(0),
		Timestamp: time.Now().UTC(),
		Type:      abstraction.MsgTypeProposal,
	}

	if _, err := mapper.FromCanonicalByzantine(canonical, ByzantineActionDoubleVote, ByzantineOptions{}); err == nil {
		t.Fatalf("expected error when applying double vote to a proposal message")
	}
}
