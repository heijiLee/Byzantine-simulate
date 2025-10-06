package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"codec/message/abstraction"
	"codec/message/abstraction/validator"

	cometbftAdapter "codec/cometbft/adapter"
	besuAdapter "codec/hyperledger/besu/adapter"
	kaiaAdapter "codec/kaia/adapter"
)

func main() {
	fmt.Println("=== Byzantine Message Bridge Test Suite ===")

	// Test message mappers
	testMappers()

	// Test validators
	testValidators()

	// Test round-trip conversions
	testRoundTripConversions()

	fmt.Println("\n=== All Tests Completed ===")
}

func testMappers() {
	fmt.Println("\n--- Testing Message Mappers ---")

	// Test CometBFT mapper
	testCometBFTMapper()

	// Test Besu mapper
	testBesuMapper()

	// Test Kaia mapper
	testKaiaMapper()
}

func testCometBFTMapper() {
	fmt.Println("\nTesting CometBFT Mapper:")
	mapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

	// Test supported types
	supportedTypes := mapper.GetSupportedTypes()
	fmt.Printf("  Supported types: %v\n", supportedTypes)

	// Test chain type
	chainType := mapper.GetChainType()
	fmt.Printf("  Chain type: %s\n", chainType)

	// Create sample raw message
	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	// Convert to canonical
	canonical, err := mapper.ToCanonical(rawMsg)
	if err != nil {
		log.Printf("  Error converting to canonical: %v", err)
		return
	}

	fmt.Printf("  Converted to canonical: chain=%s, type=%s, height=%v\n",
		canonical.ChainID, canonical.Type, canonical.Height)

	// Convert back to raw
	rawBack, err := mapper.FromCanonical(canonical)
	if err != nil {
		log.Printf("  Error converting back to raw: %v", err)
		return
	}

	fmt.Printf("  Converted back to raw: chain=%s, type=%s\n",
		rawBack.ChainID, rawBack.MessageType)
}

func testBesuMapper() {
	fmt.Println("\nTesting Besu Mapper:")
	mapper := besuAdapter.NewBesuMapper("testnet-besu")

	// Test supported types
	supportedTypes := mapper.GetSupportedTypes()
	fmt.Printf("  Supported types: %v\n", supportedTypes)

	// Create sample raw message
	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeHyperledger,
		ChainID:     "testnet-besu",
		MessageType: "PROPOSAL",
		Payload:     []byte(`{"block_number":1000,"round_number":1,"type":"PROPOSAL","block_hash":"0x789abc","proposer":"validator1","gas_limit":8000000,"timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	// Convert to canonical
	canonical, err := mapper.ToCanonical(rawMsg)
	if err != nil {
		log.Printf("  Error converting to canonical: %v", err)
		return
	}

	fmt.Printf("  Converted to canonical: chain=%s, type=%s, height=%v\n",
		canonical.ChainID, canonical.Type, canonical.Height)

	// Check extensions
	if canonical.Extensions != nil {
		if gasLimit, ok := canonical.Extensions["gas_limit"].(float64); ok {
			fmt.Printf("  Gas Limit: %v\n", gasLimit)
		}
	}
}

func testKaiaMapper() {
	fmt.Println("\nTesting Kaia Mapper:")
	mapper := kaiaAdapter.NewKaiaMapper("testnet-kaia")

	// Test supported types
	supportedTypes := mapper.GetSupportedTypes()
	fmt.Printf("  Supported types: %v\n", supportedTypes)

	// Create sample raw message
	rawMsg := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeKaia,
		ChainID:     "testnet-kaia",
		MessageType: "PROPOSAL",
		Payload:     []byte(`{"block_number":1000,"round_number":1,"type":"PROPOSAL","block_hash":"0x456def","proposer":"validator1","gas_limit":8000000,"consensus_type":"istanbul","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	// Convert to canonical
	canonical, err := mapper.ToCanonical(rawMsg)
	if err != nil {
		log.Printf("  Error converting to canonical: %v", err)
		return
	}

	fmt.Printf("  Converted to canonical: chain=%s, type=%s, height=%v\n",
		canonical.ChainID, canonical.Type, canonical.Height)

	// Check extensions
	if canonical.Extensions != nil {
		if consensusType, ok := canonical.Extensions["consensus_type"].(string); ok {
			fmt.Printf("  Consensus Type: %s\n", consensusType)
		}
	}
}

func testValidators() {
	fmt.Println("\n--- Testing Message Validators ---")

	// Test CometBFT validator
	testCometBFTValidator()

	// Test Besu validator
	testBesuValidator()

	// Test Kaia validator
	testKaiaValidator()
}

func testCometBFTValidator() {
	fmt.Println("\nTesting CometBFT Validator:")
	validator := validator.NewValidator(abstraction.ChainTypeCometBFT)

	// Create valid message
	validMsg := &abstraction.CanonicalMessage{
		ChainID:   "testnet-cometbft",
		Height:    big.NewInt(1000),
		Round:     big.NewInt(1),
		Timestamp: time.Now(),
		Type:      abstraction.MsgTypeProposal,
		Proposer:  "node1",
		Signature: "sig123",
	}

	err := validator.Validate(validMsg)
	if err != nil {
		fmt.Printf("  Valid message failed validation: %v\n", err)
	} else {
		fmt.Println("  Valid message passed validation")
	}

	// Create invalid message (missing required fields)
	invalidMsg := &abstraction.CanonicalMessage{
		ChainID: "testnet-cometbft",
		// Missing height, round, timestamp, type
	}

	err = validator.Validate(invalidMsg)
	if err != nil {
		fmt.Printf("  Invalid message correctly failed validation: %v\n", err)
	} else {
		fmt.Println("  Invalid message incorrectly passed validation")
	}
}

func testKaiaValidator() {
	fmt.Println("\nTesting Kaia Validator:")
	validator := validator.NewValidator(abstraction.ChainTypeKaia)

	// Create valid message
	validMsg := &abstraction.CanonicalMessage{
		ChainID:   "testnet-kaia",
		Height:    big.NewInt(1000),
		Round:     big.NewInt(1),
		Timestamp: time.Now(),
		Type:      abstraction.MsgTypeProposal,
		Proposer:  "validator1",
	}

	err := validator.Validate(validMsg)
	if err != nil {
		fmt.Printf("  Valid message failed validation: %v\n", err)
	} else {
		fmt.Println("  Valid message passed validation")
	}
}

func testRoundTripConversions() {
	fmt.Println("\n--- Testing Round-Trip Conversions ---")

	// Test CometBFT -> Canonical -> Fabric
	testCrossChainConversion()
}

func testCrossChainConversion() {
	fmt.Println("\nTesting Cross-Chain Conversion (CometBFT -> Fabric):")

	// Create CometBFT mapper
	cometbftMapper := cometbftAdapter.NewCometBFTMapper("testnet-cometbft")

	// Create Fabric mapper
	fabricMapper := fabricAdapter.NewFabricMapper("testnet-fabric")

	// Create sample CometBFT message
	cometbftRaw := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     "testnet-cometbft",
		MessageType: "Proposal",
		Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1","timestamp":"2024-01-01T00:00:00Z"}`),
		Encoding:    "json",
		Timestamp:   time.Now(),
	}

	// Convert to canonical
	canonical, err := cometbftMapper.ToCanonical(cometbftRaw)
	if err != nil {
		log.Printf("  Error converting CometBFT to canonical: %v", err)
		return
	}

	fmt.Printf("  CometBFT -> Canonical: type=%s, height=%v\n", canonical.Type, canonical.Height)

	// Convert canonical to Fabric
	fabricRaw, err := fabricMapper.FromCanonical(canonical)
	if err != nil {
		log.Printf("  Error converting canonical to Fabric: %v", err)
		return
	}

	fmt.Printf("  Canonical -> Fabric: type=%s\n", fabricRaw.MessageType)

	// Verify the conversion
	fabricCanonical, err := fabricMapper.ToCanonical(*fabricRaw)
	if err != nil {
		log.Printf("  Error converting Fabric back to canonical: %v", err)
		return
	}

	fmt.Printf("  Fabric -> Canonical: type=%s, height=%v\n", fabricCanonical.Type, fabricCanonical.Height)

	// Check if heights match
	if canonical.Height.Cmp(fabricCanonical.Height) == 0 {
		fmt.Println("  ✓ Height preserved across conversion")
	} else {
		fmt.Println("  ✗ Height mismatch across conversion")
	}
}

type CompareProfile struct {
	Type, Height, Round, View, Timestamp bool
	BlockHash, PrevHash                  bool
	Proposer, Validator, Signature       bool
	CommitSeals, ViewChanges             bool
	Extras, RawPayload                   bool
}

// 포맷별로 parsing 결과를 비교할 수 있게 정규화된 형태로 변환
func canonicalizeForCompare(m *abstraction.AbstractMessage, formatName string) *abstraction.AbstractMessage {
	//시간 정규화
	if !m.Timestamp.IsZero() {
		m.Timestamp = m.Timestamp.UTC().Truncate(time.Second)
	}
	//nil slice → empty slice
	if m.CommitSeals == nil {
		m.CommitSeals = []string{}
	}
	if m.ViewChanges == nil {
		m.ViewChanges = []abstraction.ViewChangeEntry{}
	}
	if m.Extras == nil {
		m.Extras = map[string][]byte{}
	}
	//big.Int → 0
	if m.Height == nil {
		m.Height = big.NewInt(0)
	}
	if m.Round == nil {
		m.Round = big.NewInt(0)
	}
	if m.View == nil {
		m.View = big.NewInt(0)
	}
	m.RawPayload = nil
	switch formatName {
	case "json", "rlp", "msgpack", "bcs", "generic":
		for k, v := range m.Extras {
			if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
				m.Extras[k] = bytes.Trim(v, "\"")
			}
		}
	}

	return m
}

func compareWithProfile(a, b *abstraction.AbstractMessage, p CompareProfile) (bool, string) {
	var sb strings.Builder
	ok := true
	if p.Type && string(a.Type) != string(b.Type) {
		ok = false
		sb.WriteString(fmt.Sprintf("Type: %s != %s\n", a.Type, b.Type))
	}
	if p.Height && !bigIntEqual(a.Height, b.Height) {
		ok = false
		sb.WriteString(fmt.Sprintf("Height: %v != %v\n", a.Height, b.Height))
	}
	if p.Round && !bigIntEqual(a.Round, b.Round) {
		ok = false
		sb.WriteString(fmt.Sprintf("Round: %v != %v\n", a.Round, b.Round))
	}
	if p.View && !bigIntEqual(a.View, b.View) {
		ok = false
		sb.WriteString(fmt.Sprintf("View: %v != %v\n", a.View, b.View))
	}
	if p.Timestamp && !a.Timestamp.Equal(b.Timestamp) {
		ok = false
		sb.WriteString(fmt.Sprintf("Timestamp: %s != %s\n", a.Timestamp, b.Timestamp))
	}
	if p.BlockHash && a.BlockHash != b.BlockHash {
		ok = false
		sb.WriteString(fmt.Sprintf("BlockHash: %s != %s\n", a.BlockHash, b.BlockHash))
	}
	if p.PrevHash && a.PrevHash != b.PrevHash {
		ok = false
		sb.WriteString(fmt.Sprintf("PrevHash: %s != %s\n", a.PrevHash, b.PrevHash))
	}
	if p.Proposer && a.Proposer != b.Proposer {
		ok = false
		sb.WriteString(fmt.Sprintf("Proposer: %s != %s\n", a.Proposer, b.Proposer))
	}
	if p.Validator && a.Validator != b.Validator {
		ok = false
		sb.WriteString(fmt.Sprintf("Validator: %s != %s\n", a.Validator, b.Validator))
	}
	if p.Signature && a.Signature != b.Signature {
		ok = false
		sb.WriteString(fmt.Sprintf("Signature: %q != %q\n", a.Signature, b.Signature))
	}
	if p.CommitSeals {
		if len(a.CommitSeals) != len(b.CommitSeals) {
			ok = false
			sb.WriteString(fmt.Sprintf("CommitSeals length: %d != %d\n", len(a.CommitSeals), len(b.CommitSeals)))
		} else {
			for i := range a.CommitSeals {
				if a.CommitSeals[i] != b.CommitSeals[i] {
					ok = false
					sb.WriteString(fmt.Sprintf("CommitSeals[%d]: %q != %q\n", i, a.CommitSeals[i], b.CommitSeals[i]))
				}
			}
		}
	}
	if p.ViewChanges {
		if len(a.ViewChanges) != len(b.ViewChanges) {
			ok = false
			sb.WriteString(fmt.Sprintf("ViewChanges length: %d != %d\n", len(a.ViewChanges), len(b.ViewChanges)))
		} else {
			for i := range a.ViewChanges {
				va, vb := a.ViewChanges[i], b.ViewChanges[i]
				if !bigIntEqual(va.Height, vb.Height) || !bigIntEqual(va.View, vb.View) ||
					va.Validator != vb.Validator || va.Signature != vb.Signature {
					ok = false
					sb.WriteString(fmt.Sprintf("ViewChanges[%d] mismatch: %+v != %+v\n", i, va, vb))
				}
			}
		}
	}
	if p.Extras {
		if len(a.Extras) != len(b.Extras) {
			ok = false
			sb.WriteString(fmt.Sprintf("Extras length: %d != %d\n", len(a.Extras), len(b.Extras)))
		} else {
			for k, va := range a.Extras {
				vb, okk := b.Extras[k]
				if !okk || !bytes.Equal(va, vb) {
					ok = false
					sb.WriteString(fmt.Sprintf("Extras[%s] mismatch: %v != %v\n", k, va, vb))
				}
			}
		}
	}
	if p.RawPayload && !bytes.Equal(a.RawPayload, b.RawPayload) {
		ok = false
		sb.WriteString(fmt.Sprintf("RawPayload mismatch: %v != %v\n", a.RawPayload, b.RawPayload))
	}
	return ok, sb.String()
}

func sampleMessage() *abstraction.AbstractMessage {
	return &abstraction.AbstractMessage{
		Type:      abstraction.MsgTypeProposal,
		Height:    big.NewInt(1000),
		Round:     big.NewInt(2),
		View:      big.NewInt(0),
		Timestamp: time.Now().UTC().Truncate(time.Second),
		BlockHash: "0xdeadbeef",
		PrevHash:  "0xfeedbead",
		Proposer:  "node 1",
		Validator: "node 1",
		Signature: "SIG_ORIG",
		CommitSeals: []string{
			"seal A", "seal B",
		},
		ViewChanges: []abstraction.ViewChangeEntry{
			{
				View:      big.NewInt(1),
				Height:    big.NewInt(1000),
				Validator: "node 2",
				Signature: "vc_sig",
			},
		},
		Extras:     map[string][]byte{"payload": []byte("hello")},
		RawPayload: []byte("raw-bytes"),
	}
}

func copyAM(m *abstraction.AbstractMessage) *abstraction.AbstractMessage {
	c := *m
	if m.Height != nil {
		c.Height = new(big.Int).Set(m.Height)
	}
	if m.Round != nil {
		c.Round = new(big.Int).Set(m.Round)
	}
	if m.View != nil {
		c.View = new(big.Int).Set(m.View)
	}
	if m.CommitSeals != nil {
		c.CommitSeals = append([]string(nil), m.CommitSeals...)
	}
	if m.ViewChanges != nil {
		c.ViewChanges = append([]abstraction.ViewChangeEntry(nil), m.ViewChanges...)
		for i := range c.ViewChanges {
			if c.ViewChanges[i].Height != nil {
				c.ViewChanges[i].Height = new(big.Int).Set(c.ViewChanges[i].Height)
			}
			if c.ViewChanges[i].View != nil {
				c.ViewChanges[i].View = new(big.Int).Set(c.ViewChanges[i].View)
			}
		}
	}
	if m.Extras != nil {
		c.Extras = make(map[string][]byte, len(m.Extras))
		for k, v := range m.Extras {
			c.Extras[k] = append([]byte(nil), v...)
		}
	}
	if m.RawPayload != nil {
		c.RawPayload = append([]byte(nil), m.RawPayload...)
	}
	if m.OriginalFieldNames != nil {
		c.OriginalFieldNames = make(map[string]string, len(m.OriginalFieldNames))
		for k, v := range m.OriginalFieldNames {
			c.OriginalFieldNames[k] = v
		}
	}
	return &c
}

func bigIntEqual(a, b *big.Int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Cmp(b) == 0
}

func previewHex(b []byte, n int) string {
	if len(b) == 0 {
		return "<empty>"
	}
	if n > len(b) {
		n = len(b)
	}
	return hex.EncodeToString(b[:n])
}

func runSynonymTests() {
	fmt.Println("\n=== Synonym mapping tests ===")
	phaseInputs := []string{"Propose", "PrePrepare", "Announce", "Vote_Commit"}
	for _, in := range phaseInputs {
		normalized := codec.PhaseSynonyms[in]
		if normalized == "" {
			fmt.Printf("Phase synonym: %-12s -> Not found\n", in)
		} else {
			fmt.Printf("Phase synonym: %-12s -> %s\n", in, normalized)
		}
	}
	fieldInputs := []string{"seq_num", "block_digest", "leader", "sig", "vc_entries"}
	for _, in := range fieldInputs {
		normalized := codec.FieldSynonyms[in]
		if normalized == "" {
			fmt.Printf("Field synonym: %-15s -> Not found\n", in)
		} else {
			fmt.Printf("Field synonym: %-15s -> %s\n", in, normalized)
		}
	}
	am := sampleMessage()
	js, err := codec.Serialize(am, codec.SerializeOptions{Format: codec.FormatJSON})
	if err != nil {
		log.Printf("[ERROR] Serialize JSON: %v", err)
		return
	}
	fmt.Printf("\nOriginal JSON: %s\n", string(js))
	modified := strings.ReplaceAll(string(js), "Height", "seq_num")
	modified = strings.ReplaceAll(modified, "Signature", "sig")

	parsed, err := codec.Parse([]byte(modified), codec.ParseOptions{Format: codec.FormatJSON})
	if err != nil {
		log.Printf("[ERROR] Parse JSON (synonym keys): %v", err)
		return
	}
	fmt.Printf("Parsed from modified JSON:\n  Height=%v, Signature=%q\n",
		parsed.Height, parsed.Signature)
}
