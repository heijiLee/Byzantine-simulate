package engine

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
	consensuspb "github.com/cometbft/cometbft/proto/tendermint/consensus"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
)

const (
	stateChannelID       byte = 0x20
	dataChannelID        byte = 0x21
	voteChannelID        byte = 0x22
	voteSetBitsChannelID byte = 0x23
)

var consensusChannelSet = map[byte]struct{}{
	stateChannelID:       {},
	dataChannelID:        {},
	voteChannelID:        {},
	voteSetBitsChannelID: {},
}

var errUnsupportedMessage = errors.New("unsupported consensus message")

type flowDirection string

const (
	directionUpstream   flowDirection = "upstream"
	directionDownstream flowDirection = "downstream"
)

func isConsensusChannel(chID byte) bool {
	_, ok := consensusChannelSet[chID]
	return ok
}

func decodeConsensusMessage(payload []byte) (*consensuspb.Message, error) {
	var msg consensuspb.Message
	if err := msg.Unmarshal(payload); err != nil {
		return nil, err
	}
	return &msg, nil
}

func canonicalFromConsensus(mapper *cometbftAdapter.CometBFTMapper, chainID string, msg *consensuspb.Message) (*abstraction.CanonicalMessage, error) {
	adapterMsg, messageType, err := adapterMessageFromConsensus(msg)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(adapterMsg)
	if err != nil {
		return nil, err
	}
	raw := abstraction.RawConsensusMessage{
		ChainType:   abstraction.ChainTypeCometBFT,
		ChainID:     chainID,
		MessageType: messageType,
		Payload:     payload,
		Encoding:    "json",
		Timestamp:   adapterMsg.Timestamp,
	}
	return mapper.ToCanonical(raw)
}

func adapterMessageFromConsensus(msg *consensuspb.Message) (*cometbftAdapter.CometBFTConsensusMessage, string, error) {
	switch payload := msg.Sum.(type) {
	case *consensuspb.Message_Proposal:
		return proposalToAdapter(payload.Proposal)
	case *consensuspb.Message_Vote:
		return voteToAdapter(payload.Vote)
	default:
		return nil, "", errUnsupportedMessage
	}
}

func proposalToAdapter(wrapper *consensuspb.Proposal) (*cometbftAdapter.CometBFTConsensusMessage, string, error) {
	if wrapper == nil || wrapper.Proposal == nil {
		return nil, "", fmt.Errorf("empty proposal payload")
	}
	p := wrapper.Proposal
	msg := &cometbftAdapter.CometBFTConsensusMessage{
		MessageType: "Proposal",
		Height:      strconv.FormatInt(p.Height, 10),
		Round:       strconv.FormatInt(int64(p.Round), 10),
		Timestamp:   p.Timestamp,
		POLRound:    p.POLRound,
		Signature:   encodeBase64(p.Signature),
		BlockID: cometbftAdapter.BlockID{
			Hash:          hex.EncodeToString(p.BlockID.Hash),
			PartSetHeader: cometbftAdapter.PartSetHeader{Total: uint32(p.BlockID.PartSetHeader.Total), Hash: append([]byte(nil), p.BlockID.PartSetHeader.Hash...)},
		},
	}
	return msg, msg.MessageType, nil
}

func voteToAdapter(wrapper *consensuspb.Vote) (*cometbftAdapter.CometBFTConsensusMessage, string, error) {
	if wrapper == nil || wrapper.Vote == nil {
		return nil, "", fmt.Errorf("empty vote payload")
	}
	v := wrapper.Vote
	msg := &cometbftAdapter.CometBFTConsensusMessage{
		MessageType:        "Vote",
		Type:               int32(v.Type),
		Height:             strconv.FormatInt(v.Height, 10),
		Round:              strconv.FormatInt(int64(v.Round), 10),
		Timestamp:          v.Timestamp,
		BlockID:            cometbftAdapter.BlockID{Hash: hex.EncodeToString(v.BlockID.Hash), PartSetHeader: cometbftAdapter.PartSetHeader{Total: uint32(v.BlockID.PartSetHeader.Total), Hash: append([]byte(nil), v.BlockID.PartSetHeader.Hash...)}},
		ValidatorAddress:   hex.EncodeToString(v.ValidatorAddress),
		ValidatorIndex:     v.ValidatorIndex,
		Signature:          encodeBase64(v.Signature),
		Extension:          encodeBase64(v.Extension),
		ExtensionSignature: encodeBase64(v.ExtensionSignature),
	}
	return msg, msg.MessageType, nil
}

func rawToConsensusMessage(raw *abstraction.RawConsensusMessage) (*consensuspb.Message, error) {
	var adapterMsg cometbftAdapter.CometBFTConsensusMessage
	if err := json.Unmarshal(raw.Payload, &adapterMsg); err != nil {
		return nil, err
	}
	return protoFromAdapterMessage(&adapterMsg)
}

func protoFromAdapterMessage(msg *cometbftAdapter.CometBFTConsensusMessage) (*consensuspb.Message, error) {
	switch strings.ToLower(msg.MessageType) {
	case "proposal":
		height, err := strconv.ParseInt(msg.Height, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid proposal height: %w", err)
		}
		round, err := strconv.ParseInt(msg.Round, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid proposal round: %w", err)
		}
		proposal := &cmttypes.Proposal{
			Type:      cmtproto.ProposalType,
			Height:    height,
			Round:     int32(round),
			POLRound:  msg.POLRound,
			BlockID:   typesBlockIDFromAdapter(msg.BlockID),
			Timestamp: ensureTime(msg.Timestamp),
			Signature: decodeString(msg.Signature),
		}
		return &consensuspb.Message{Sum: &consensuspb.Message_Proposal{Proposal: &consensuspb.Proposal{Proposal: proposal}}}, nil
	case "vote":
		height, err := strconv.ParseInt(msg.Height, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid vote height: %w", err)
		}
		round, err := strconv.ParseInt(msg.Round, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid vote round: %w", err)
		}
		vote := &cmttypes.Vote{
			Type:               cmttypes.SignedMsgType(msg.Type),
			Height:             height,
			Round:              int32(round),
			Timestamp:          ensureTime(msg.Timestamp),
			BlockID:            typesBlockIDFromAdapter(msg.BlockID),
			ValidatorAddress:   decodeHexString(msg.ValidatorAddress),
			ValidatorIndex:     msg.ValidatorIndex,
			Signature:          decodeString(msg.Signature),
			Extension:          decodeString(msg.Extension),
			ExtensionSignature: decodeString(msg.ExtensionSignature),
		}
		return &consensuspb.Message{Sum: &consensuspb.Message_Vote{Vote: &consensuspb.Vote{Vote: vote}}}, nil
	default:
		return nil, errUnsupportedMessage
	}
}

func typesBlockIDFromAdapter(block cometbftAdapter.BlockID) cmttypes.BlockID {
	return cmttypes.BlockID{
		Hash: hexDecodeOrCopy(block.Hash),
		PartSetHeader: cmttypes.PartSetHeader{
			Total: int(block.PartSetHeader.Total),
			Hash:  append([]byte(nil), block.PartSetHeader.Hash...),
		},
	}
}

func encodeBase64(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func decodeString(value string) []byte {
	if value == "" {
		return nil
	}
	if data, err := base64.StdEncoding.DecodeString(value); err == nil {
		return data
	}
	return hexDecodeOrCopy(value)
}

func hexDecodeOrCopy(value string) []byte {
	if value == "" {
		return nil
	}
	if data, err := hex.DecodeString(strings.TrimPrefix(value, "0x")); err == nil {
		return data
	}
	return []byte(value)
}

func decodeHexString(value string) []byte {
	if value == "" {
		return nil
	}
	if data, err := hex.DecodeString(strings.TrimPrefix(value, "0x")); err == nil {
		return data
	}
	// addresses are expected to be hex, but fall back to raw bytes
	return []byte(value)
}

func ensureTime(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts
}

func marshalConsensusMessage(msg *consensuspb.Message) ([]byte, error) {
	return gogoproto.Marshal(msg)
}
