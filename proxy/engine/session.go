package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
	p2pconn "github.com/cometbft/cometbft/p2p/conn"
	consensuspb "github.com/cometbft/cometbft/proto/tendermint/consensus"
	evidpb "github.com/cometbft/cometbft/proto/tendermint/evidence"
	mempoolpb "github.com/cometbft/cometbft/proto/tendermint/mempool"
)

const (
	mempoolChannelID  byte = 0x30
	evidenceChannelID byte = 0x38
)

type session struct {
	ctx    context.Context
	cancel context.CancelFunc

	cfg     *Config
	mapper  *cometbftAdapter.CometBFTMapper
	metrics *Metrics

	downstream *p2pconn.MConnection
	upstream   *p2pconn.MConnection

	logger  *slog.Logger
	errOnce sync.Once
	err     error
}

func newSession(ctx context.Context, cancel context.CancelFunc, cfg *Config, mapper *cometbftAdapter.CometBFTMapper, metrics *Metrics, downstream, upstream net.Conn) *session {
	s := &session{
		ctx:     ctx,
		cancel:  cancel,
		cfg:     cfg,
		mapper:  mapper,
		metrics: metrics,
		logger:  cfg.Logger.With("remote", downstream.RemoteAddr().String()),
	}

	downRecv := func(chID byte, payload []byte) {
		s.handleDownstream(chID, payload)
	}
	upRecv := func(chID byte, payload []byte) {
		s.handleUpstream(chID, payload)
	}
	onError := func(direction flowDirection) func(any) {
		return func(err any) {
			s.recordError(fmt.Errorf("%s mconnection error: %v", direction, err))
		}
	}

	s.downstream = p2pconn.NewMConnection(downstream, defaultDescriptors(), downRecv, onError(directionDownstream))
	s.upstream = p2pconn.NewMConnection(upstream, defaultDescriptors(), upRecv, onError(directionUpstream))

	return s
}

func (s *session) run() error {
	if err := s.downstream.Start(); err != nil {
		s.recordError(err)
		return err
	}
	if err := s.upstream.Start(); err != nil {
		s.downstream.Stop()
		s.recordError(err)
		return err
	}

	defer s.downstream.FlushStop()
	defer s.upstream.FlushStop()

	<-s.ctx.Done()
	if s.err != nil {
		return s.err
	}
	return s.ctx.Err()
}

func (s *session) handleDownstream(chID byte, payload []byte) {
	if s.cfg.Direction.ShouldMutateDownstream() && isConsensusChannel(chID) {
		if err := s.processConsensus(directionDownstream, chID, payload, s.upstream); err != nil {
			s.logger.Warn("failed to process downstream consensus message", "err", err)
			s.forwardRaw(s.upstream, chID, payload)
		}
		return
	}
	s.forwardRaw(s.upstream, chID, payload)
}

func (s *session) handleUpstream(chID byte, payload []byte) {
	if s.cfg.Direction.ShouldMutateUpstream() && isConsensusChannel(chID) {
		if err := s.processConsensus(directionUpstream, chID, payload, s.downstream); err != nil {
			s.logger.Warn("failed to process upstream consensus message", "err", err)
			s.forwardRaw(s.downstream, chID, payload)
		}
		return
	}
	s.forwardRaw(s.downstream, chID, payload)
}

func (s *session) processConsensus(direction flowDirection, chID byte, payload []byte, target *p2pconn.MConnection) error {
	msg, err := decodeConsensusMessage(payload)
	if err != nil {
		return err
	}

	canonical, err := canonicalFromConsensus(s.mapper, s.cfg.ChainID, msg)
	if err != nil {
		if errors.Is(err, errUnsupportedMessage) {
			s.forwardRaw(target, chID, payload)
			return nil
		}
		return err
	}

	if !s.cfg.Trigger.Matches(canonical) {
		s.forwardRaw(target, chID, payload)
		return nil
	}

	if s.cfg.Hooks.Delay > 0 {
		s.metrics.IncDelayed()
		time.Sleep(s.cfg.Hooks.Delay)
	}

	if s.cfg.Hooks.Drop {
		s.metrics.IncDropped()
		s.logger.Info("dropped consensus message", "direction", direction, "channel", fmt.Sprintf("0x%X", chID), "height", canonicalHeight(canonical), "round", canonicalRound(canonical), "type", canonical.Type)
		return nil
	}

	raws, err := s.applyByzantineAction(canonical)
	if err != nil {
		return err
	}

	sent := 0
	duplicateCount := 0
	for _, raw := range raws {
		protoMsg, err := rawToConsensusMessage(raw)
		if err != nil {
			return err
		}
		bytes, err := marshalConsensusMessage(protoMsg)
		if err != nil {
			return err
		}
		s.forwardRaw(target, chID, bytes)
		sent++
		if s.cfg.Hooks.Duplicate {
			s.forwardRaw(target, chID, bytes)
			sent++
			duplicateCount++
		}
	}

	s.metrics.IncMutated(int64(sent))
	if duplicateCount > 0 {
		s.metrics.IncDuplicated(duplicateCount)
	}

	s.logger.Info("mutated consensus message", "direction", direction, "channel", fmt.Sprintf("0x%X", chID), "height", canonicalHeight(canonical), "round", canonicalRound(canonical), "type", canonical.Type, "count", sent, "duplicates", duplicateCount)

	return nil
}

func (s *session) applyByzantineAction(canonical *abstraction.CanonicalMessage) ([]*abstraction.RawConsensusMessage, error) {
	if s.cfg.Action == cometbftAdapter.ByzantineActionNone {
		raw, err := s.mapper.FromCanonical(canonical)
		if err != nil {
			return nil, err
		}
		return []*abstraction.RawConsensusMessage{raw}, nil
	}
	return s.mapper.FromCanonicalByzantine(canonical, s.cfg.Action, s.cfg.Options)
}

func (s *session) forwardRaw(target *p2pconn.MConnection, chID byte, payload []byte) {
	if ok := target.Send(chID, append([]byte(nil), payload...)); !ok {
		s.logger.Warn("failed to forward message", "channel", fmt.Sprintf("0x%X", chID))
	}
}

func (s *session) recordError(err error) {
	if err == nil {
		return
	}
	s.errOnce.Do(func() {
		s.err = err
		s.cancel()
	})
}

func canonicalHeight(msg *abstraction.CanonicalMessage) int64 {
	if msg == nil || msg.Height == nil {
		return 0
	}
	return msg.Height.Int64()
}

func canonicalRound(msg *abstraction.CanonicalMessage) int64 {
	if msg == nil || msg.Round == nil {
		return 0
	}
	return msg.Round.Int64()
}

func defaultDescriptors() []*p2pconn.ChannelDescriptor {
	return []*p2pconn.ChannelDescriptor{
		{
			ID:                  stateChannelID,
			Priority:            6,
			SendQueueCapacity:   100,
			RecvMessageCapacity: 1 << 20,
			MessageType:         &consensuspb.Message{},
		},
		{
			ID:                  dataChannelID,
			Priority:            10,
			SendQueueCapacity:   100,
			RecvMessageCapacity: 1 << 20,
			MessageType:         &consensuspb.Message{},
		},
		{
			ID:                  voteChannelID,
			Priority:            7,
			SendQueueCapacity:   100,
			RecvMessageCapacity: 1 << 20,
			MessageType:         &consensuspb.Message{},
		},
		{
			ID:                  voteSetBitsChannelID,
			Priority:            1,
			SendQueueCapacity:   10,
			RecvMessageCapacity: 1 << 20,
			MessageType:         &consensuspb.Message{},
		},
		{
			ID:                  mempoolChannelID,
			Priority:            5,
			SendQueueCapacity:   128,
			RecvMessageCapacity: 1 << 21,
			MessageType:         &mempoolpb.Message{},
		},
		{
			ID:                  evidenceChannelID,
			Priority:            4,
			SendQueueCapacity:   32,
			RecvMessageCapacity: 1 << 20,
			MessageType:         &evidpb.Message{},
		},
	}
}
