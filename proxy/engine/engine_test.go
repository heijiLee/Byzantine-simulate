package engine

import (
	"context"
	"encoding/hex"
	"net"
	"testing"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	p2pconn "github.com/cometbft/cometbft/p2p/conn"
	consensuspb "github.com/cometbft/cometbft/proto/tendermint/consensus"
	cmttypes "github.com/cometbft/cometbft/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
)

func TestSessionDoublePrevoteMutation(t *testing.T) {
	nodeKey := &p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}
	height := int64(11)
	trigger := Trigger{Height: &height, Step: "prevote"}

	cfg, err := NewConfig(ConfigOptions{
		ListenAddress:  "tcp://0.0.0.0:0",
		UpstreamTarget: "tcp://0.0.0.0:0",
		ChainID:        "test-chain",
		NodeKey:        nodeKey,
		Action:         cometbftAdapter.ByzantineActionDoubleVote,
		Trigger:        trigger,
		Hooks:          Hooks{},
		Direction:      DirectionUpstream,
		DialTimeout:    time.Second,
	})
	if err != nil {
		t.Fatalf("config error: %v", err)
	}

	harness := newProxyHarness(t, cfg)
	defer harness.Close()

	voteBytes := make([]byte, 32)
	for i := range voteBytes {
		voteBytes[i] = 0xAA
	}
	vote := &cmttypes.Vote{
		Type:             cmttypes.PrevoteType,
		Height:           height,
		Round:            2,
		Timestamp:        time.Now().UTC(),
		BlockID:          cmttypes.BlockID{Hash: voteBytes, PartSetHeader: cmttypes.PartSetHeader{Total: 1, Hash: []byte{0x01}}},
		ValidatorAddress: []byte("validator-1"),
		ValidatorIndex:   7,
		Signature:        []byte("sig"),
	}
	payload := &consensuspb.Message{Sum: &consensuspb.Message_Vote{Vote: &consensuspb.Vote{Vote: vote}}}
	raw, err := gogoproto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal vote: %v", err)
	}

	harness.sendUpstream(voteChannelID, raw)

	msg1 := harness.mustReceive()
	msg2 := harness.mustReceive()

	vote1 := decodeVote(t, msg1)
	vote2 := decodeVote(t, msg2)

	hash1 := hex.EncodeToString(vote1.BlockID.Hash)
	hash2 := hex.EncodeToString(vote2.BlockID.Hash)
	original := hex.EncodeToString(vote.BlockID.Hash)

	if hash1 == hash2 {
		t.Fatalf("expected distinct hashes, got %s and %s", hash1, hash2)
	}
	if hash1 != original && hash2 != original {
		t.Fatalf("expected one hash to match original %s, got %s and %s", original, hash1, hash2)
	}
	if hash1 == original {
		ensureHashMutated(t, hash2, original)
	} else {
		ensureHashMutated(t, hash1, original)
	}

	harness.assertNoExtraMessages()
}

func TestSessionDropTriggeredMessage(t *testing.T) {
	nodeKey := &p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}
	height := int64(5)
	trigger := Trigger{Height: &height, Step: "prevote"}

	cfg, err := NewConfig(ConfigOptions{
		ListenAddress:  "tcp://0.0.0.0:0",
		UpstreamTarget: "tcp://0.0.0.0:0",
		ChainID:        "drop-chain",
		NodeKey:        nodeKey,
		Action:         cometbftAdapter.ByzantineActionNone,
		Trigger:        trigger,
		Hooks: Hooks{
			Drop: true,
		},
		Direction:   DirectionUpstream,
		DialTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("config error: %v", err)
	}

	harness := newProxyHarness(t, cfg)
	defer harness.Close()

	vote := &cmttypes.Vote{
		Type:             cmttypes.PrevoteType,
		Height:           height,
		Round:            1,
		Timestamp:        time.Now().UTC(),
		BlockID:          cmttypes.BlockID{Hash: []byte{0x01}, PartSetHeader: cmttypes.PartSetHeader{Total: 1, Hash: []byte{0x02}}},
		ValidatorAddress: []byte("validator"),
		ValidatorIndex:   3,
		Signature:        []byte("sig"),
	}
	payload := &consensuspb.Message{Sum: &consensuspb.Message_Vote{Vote: &consensuspb.Vote{Vote: vote}}}
	raw, err := gogoproto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal vote: %v", err)
	}

	harness.sendUpstream(voteChannelID, raw)

	select {
	case <-harness.received:
		t.Fatalf("expected message to be dropped")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestSessionDelayProposal(t *testing.T) {
	nodeKey := &p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}
	height := int64(7)
	trigger := Trigger{Height: &height, Step: "proposal"}
	delay := 200 * time.Millisecond

	cfg, err := NewConfig(ConfigOptions{
		ListenAddress:  "tcp://0.0.0.0:0",
		UpstreamTarget: "tcp://0.0.0.0:0",
		ChainID:        "delay-chain",
		NodeKey:        nodeKey,
		Action:         cometbftAdapter.ByzantineActionNone,
		Trigger:        trigger,
		Hooks: Hooks{
			Delay: delay,
		},
		Direction:   DirectionUpstream,
		DialTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("config error: %v", err)
	}

	harness := newProxyHarness(t, cfg)
	defer harness.Close()

	proposal := &cmttypes.Proposal{
		Type:      cmttypes.ProposalType,
		Height:    height,
		Round:     1,
		POLRound:  0,
		BlockID:   cmttypes.BlockID{Hash: []byte{0xAA}, PartSetHeader: cmttypes.PartSetHeader{Total: 1, Hash: []byte{0xBB}}},
		Timestamp: time.Now().UTC(),
		Signature: []byte("sig"),
	}
	payload := &consensuspb.Message{Sum: &consensuspb.Message_Proposal{Proposal: &consensuspb.Proposal{Proposal: proposal}}}
	raw, err := gogoproto.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal proposal: %v", err)
	}

	start := time.Now()
	harness.sendUpstream(dataChannelID, raw)

	<-harness.received
	elapsed := time.Since(start)
	if elapsed < delay {
		t.Fatalf("expected delay of at least %v, got %v", delay, elapsed)
	}
}

// proxyHarness manages a session and associated peer connections for tests.
type proxyHarness struct {
	t       *testing.T
	cfg     *Config
	mapper  *cometbftAdapter.CometBFTMapper
	metrics *Metrics
	ctx     context.Context
	cancel  context.CancelFunc
	session *session
	errCh   chan error

	upstreamPeer   *p2pconn.MConnection
	downstreamPeer *p2pconn.MConnection

	upstreamSecret   *p2pconn.SecretConnection
	downstreamSecret *p2pconn.SecretConnection

	received chan []byte
}

func newProxyHarness(t *testing.T, cfg *Config) *proxyHarness {
	t.Helper()

	mapper := cometbftAdapter.NewCometBFTMapper(cfg.ChainID)
	metrics := NewMetrics()

	localPriv, ok := cfg.NodeKey.PrivKey.(ed25519.PrivKey)
	if !ok {
		t.Fatalf("unexpected private key type %T", cfg.NodeKey.PrivKey)
	}

	downstreamSecret, downstreamPeer := makeSecretConnPair(t, localPriv, ed25519.GenPrivKey())
	upstreamSecret, upstreamPeer := makeSecretConnPair(t, localPriv, ed25519.GenPrivKey())

	ctx, cancel := context.WithCancel(context.Background())

	sess := newSession(ctx, cancel, cfg, mapper, metrics, downstreamSecret, upstreamSecret)

	received := make(chan []byte, 10)

	downRecv := func(chID byte, msg []byte) {
		if chID == stateChannelID || chID == dataChannelID || chID == voteChannelID || chID == voteSetBitsChannelID {
			received <- append([]byte(nil), msg...)
		}
	}
	upRecv := func(chID byte, msg []byte) {}

	downstreamPeerConn := p2pconn.NewMConnection(downstreamPeer, defaultDescriptors(), downRecv, func(any) {})
	upstreamPeerConn := p2pconn.NewMConnection(upstreamPeer, defaultDescriptors(), upRecv, func(any) {})

	if err := downstreamPeerConn.Start(); err != nil {
		t.Fatalf("downstream peer start: %v", err)
	}
	if err := upstreamPeerConn.Start(); err != nil {
		t.Fatalf("upstream peer start: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- sess.run()
	}()

	return &proxyHarness{
		t:                t,
		cfg:              cfg,
		mapper:           mapper,
		metrics:          metrics,
		ctx:              ctx,
		cancel:           cancel,
		session:          sess,
		errCh:            errCh,
		upstreamPeer:     upstreamPeerConn,
		downstreamPeer:   downstreamPeerConn,
		upstreamSecret:   upstreamSecret,
		downstreamSecret: downstreamSecret,
		received:         received,
	}
}

func (h *proxyHarness) sendUpstream(chID byte, payload []byte) {
	if ok := h.upstreamPeer.Send(chID, payload); !ok {
		h.t.Fatalf("failed to send upstream payload")
	}
}

func (h *proxyHarness) mustReceive() []byte {
	select {
	case msg := <-h.received:
		return msg
	case <-time.After(2 * time.Second):
		h.t.Fatalf("timed out waiting for message")
		return nil
	}
}

func (h *proxyHarness) assertNoExtraMessages() {
	select {
	case msg := <-h.received:
		h.t.Fatalf("unexpected extra message: %x", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func (h *proxyHarness) Close() {
	h.cancel()
	<-h.errCh
	h.upstreamPeer.FlushStop()
	h.downstreamPeer.FlushStop()
	h.upstreamSecret.Close()
	h.downstreamSecret.Close()
}

func decodeVote(t *testing.T, payload []byte) *cmttypes.Vote {
	t.Helper()
	var msg consensuspb.Message
	if err := msg.Unmarshal(payload); err != nil {
		t.Fatalf("decode vote: %v", err)
	}
	if msg.GetVote() == nil || msg.GetVote().GetVote() == nil {
		t.Fatalf("expected vote message")
	}
	return msg.GetVote().GetVote()
}

func ensureHashMutated(t *testing.T, got, original string) {
	t.Helper()
	if got == original {
		t.Fatalf("expected mutated hash different from original %s", original)
	}
}

func makeSecretConnPair(t *testing.T, localKey, remoteKey ed25519.PrivKey) (*p2pconn.SecretConnection, *p2pconn.SecretConnection) {
	t.Helper()
	a, b := net.Pipe()
	remoteCh := make(chan *p2pconn.SecretConnection, 1)
	errCh := make(chan error, 1)

	go func() {
		conn, err := p2pconn.MakeSecretConnection(b, remoteKey)
		if err != nil {
			errCh <- err
			return
		}
		remoteCh <- conn
		errCh <- nil
	}()

	localConn, err := p2pconn.MakeSecretConnection(a, localKey)
	if err != nil {
		t.Fatalf("local secret connection: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("remote secret connection: %v", err)
	}
	return localConn, <-remoteCh
}
