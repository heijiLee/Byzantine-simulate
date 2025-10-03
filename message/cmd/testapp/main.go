package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"codec/abstraction"
	"codec/codec"
)

func main() {
	//proto descriptor set 등록
	if err := codec.RegisterDescriptorSetFile("proto/abstraction.protoset"); err != nil {
		log.Printf("Failed to register: %v\n", err)
	} else {
		log.Println("Registered proto descriptor set")
	}
	am := sampleMessage() //샘플 메시지
	tests := []struct {
		name      string
		format    codec.Format
		serOpts   codec.SerializeOptions
		parseOpts codec.ParseOptions
		profile   CompareProfile
	}{
		{
			name:   "generic",
			format: codec.FormatGeneric,
			serOpts: codec.SerializeOptions{
				Format: codec.FormatGeneric,
			},
			parseOpts: codec.ParseOptions{
				Format: codec.FormatGeneric,
			},
			profile: CompareProfile{
				Type: true, Height: true, Round: true, View: true,
				Timestamp: true, BlockHash: true, PrevHash: true,
				Proposer: true, Validator: true, Signature: true,
				CommitSeals: false, ViewChanges: false, Extras: true, RawPayload: false,
			},
		},
		{
			name:   "json",
			format: codec.FormatJSON,
			serOpts: codec.SerializeOptions{
				Format: codec.FormatJSON,
			},
			parseOpts: codec.ParseOptions{
				Format: codec.FormatJSON,
			},
			profile: CompareProfile{
				Type: true, Height: true, Round: true, View: true,
				Timestamp: true, BlockHash: true, PrevHash: true,
				Proposer: true, Validator: true, Signature: true,
				CommitSeals: true, ViewChanges: true, Extras: true, RawPayload: false,
			},
		},
		{
			name:   "protobuf",
			format: codec.FormatProtobuf,
			serOpts: codec.SerializeOptions{
				Format:               codec.FormatProtobuf,
				ProtoMessageFullName: "pbft.AbstractMessage",
				ProtoDiscardUnknown:  true,
			},
			parseOpts: codec.ParseOptions{
				Format:               codec.FormatProtobuf,
				ProtoMessageFullName: "pbft.AbstractMessage",
				ProtoDiscardUnknown:  true,
			},
			profile: CompareProfile{
				Type: true, Height: true, Round: true, View: false,
				Timestamp: true, BlockHash: false, PrevHash: false,
				Proposer: true, Validator: true, Signature: true,
				CommitSeals: false, ViewChanges: false, Extras: false, RawPayload: false,
			},
		},
		{
			name:   "rlp",
			format: codec.FormatRLP,
			serOpts: codec.SerializeOptions{
				Format: codec.FormatRLP,
			},
			parseOpts: codec.ParseOptions{
				Format: codec.FormatRLP,
			},
			profile: CompareProfile{
				Type: true, Height: true, Round: true, View: true,
				Timestamp: true, BlockHash: true, PrevHash: true,
				Proposer: true, Validator: true, Signature: true,
				CommitSeals: true, ViewChanges: true, Extras: true, RawPayload: false,
			},
		},
		{
			name:   "msgpack",
			format: codec.FormatMsgPack,
			serOpts: codec.SerializeOptions{
				Format: codec.FormatMsgPack,
			},
			parseOpts: codec.ParseOptions{
				Format: codec.FormatMsgPack,
			},
			profile: CompareProfile{
				Type: true, Height: true, Round: true, View: true,
				Timestamp: true, BlockHash: true, PrevHash: true,
				Proposer: true, Validator: true, Signature: true,
				CommitSeals: true, ViewChanges: true, Extras: true, RawPayload: false,
			},
		},
		{
			name:   "bcs",
			format: codec.FormatBCS,
			serOpts: codec.SerializeOptions{
				Format: codec.FormatBCS,
			},
			parseOpts: codec.ParseOptions{
				Format: codec.FormatBCS,
			},
			profile: CompareProfile{
				Type: true, Height: true, Round: true, View: true,
				Timestamp: true, BlockHash: true, PrevHash: true,
				Proposer: true, Validator: true, Signature: true,
				CommitSeals: true, ViewChanges: true, Extras: true, RawPayload: false,
			},
		},
	}

	for _, t := range tests {
		fmt.Printf("\n=== Testing format: %s ===\n", t.name)
		//1) serialize
		data, err := codec.Serialize(am, t.serOpts)
		if err != nil {
			log.Printf("[ERROR] Serialize (%s): %v\n", t.name, err)
			continue
		}
		fmt.Printf("Serialized %d bytes\n", len(data))
		fmt.Printf("first 64 bytes (hex): %s\n", previewHex(data, 64))
		//2) parse
		parsed, err := codec.Parse(data, t.parseOpts)
		if err != nil {
			log.Printf("[ERROR] Parse (%s): %v\n", t.name, err)
			continue
		}
		// 3) canonicalize & compare
		aCanon := canonicalizeForCompare(copyAM(am), t.name)
		bCanon := canonicalizeForCompare(copyAM(parsed), t.name)
		ok, diff := compareWithProfile(aCanon, bCanon, t.profile)
		if ok {
			fmt.Println("Fields match (profiled)")
		} else {
			fmt.Println("Fields mismatch (profiled)")
			fmt.Println(diff)
		}
		// 4) mutation 테스트
		parsed.Signature = "CORRUPTED_SIG"
		if parsed.Extras == nil {
			parsed.Extras = map[string][]byte{}
		}
		if t.name != "protobuf" {
			parsed.Extras["injected_by_testapp"] = []byte("1")
		}
		data2, err := codec.Serialize(parsed, t.serOpts)
		if err != nil {
			log.Printf("[ERROR] Serialize (mutated) (%s): %v\n", t.name, err)
			continue
		}
		fmt.Printf("Mutated: %d bytes\n", len(data2))
		fmt.Printf("mutated first 64 bytes (hex): %s\n", previewHex(data2, 64))
		parsed2, err := codec.Parse(data2, t.parseOpts)
		if err != nil {
			log.Printf("[ERROR] Parse (mutated) (%s): %v\n", t.name, err)
			continue
		}
		if parsed2.Signature == "CORRUPTED_SIG" {
			fmt.Println("Mutation confirmed: Signature corrupted")
		} else {
			fmt.Printf("Mutation failed: Signature unchanged (%q)\n", parsed2.Signature)
		}
		if t.name != "protobuf" {
			if v, ok := parsed2.Extras["injected_by_testapp"]; ok {
				if string(v) == "1" || string(v) == "\"1\"" {
					fmt.Printf("Mutation confirmed: Extras injected (%s)\n", string(v))
				} else {
					fmt.Printf("Mutation maybe injected but normalized differently (%q)\n", string(v))
				}
			} else {
				fmt.Println("Mutation failed: Extras unchanged")
			}
		}
	}
	runSynonymTests()
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


