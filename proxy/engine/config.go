package engine

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/message/abstraction"
	"github.com/cometbft/cometbft/p2p"
)

// Direction indicates which link should have mutations applied.
type Direction int

const (
	// DirectionUpstream mutates messages originating from the upstream validator before they reach external peers.
	DirectionUpstream Direction = iota
	// DirectionDownstream mutates messages heading to the upstream validator.
	DirectionDownstream
	// DirectionBoth mutates messages flowing in both directions.
	DirectionBoth
)

// ParseDirection parses a textual direction value.
func ParseDirection(value string) (Direction, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "upstream":
		return DirectionUpstream, nil
	case "downstream":
		return DirectionDownstream, nil
	case "both":
		return DirectionBoth, nil
	default:
		return DirectionUpstream, fmt.Errorf("unknown direction %q", value)
	}
}

// ShouldMutateUpstream reports if messages from the upstream validator should be mutated.
func (d Direction) ShouldMutateUpstream() bool {
	return d == DirectionUpstream || d == DirectionBoth
}

// ShouldMutateDownstream reports if messages heading to the upstream validator should be mutated.
func (d Direction) ShouldMutateDownstream() bool {
	return d == DirectionDownstream || d == DirectionBoth
}

// Trigger describes when the engine should mutate traffic.
type Trigger struct {
	Height *int64
	Round  *int64
	Step   string
}

// Matches reports whether the canonical message satisfies the trigger.
func (t Trigger) Matches(msg *abstraction.CanonicalMessage) bool {
	if msg == nil {
		return false
	}
	if t.Height != nil {
		if msg.Height == nil || msg.Height.Int64() != *t.Height {
			return false
		}
	}
	if t.Round != nil {
		if msg.Round == nil || msg.Round.Int64() != *t.Round {
			return false
		}
	}
	if t.Step != "" && strings.ToLower(string(msg.Type)) != t.Step {
		return false
	}
	return true
}

// Hooks define behavioural mutations around forwarding.
type Hooks struct {
	Delay     time.Duration
	Drop      bool
	Duplicate bool
}

// Config holds the runtime configuration for the proxy engine.
type Config struct {
	ListenNetwork   string
	ListenAddress   string
	UpstreamNetwork string
	UpstreamAddress string

	ChainID string

	NodeKey *p2p.NodeKey

	Action  cometbftAdapter.ByzantineAction
	Options cometbftAdapter.ByzantineOptions

	Trigger   Trigger
	Hooks     Hooks
	Direction Direction

	DialTimeout time.Duration

	Logger *slog.Logger
}

// ConfigOptions contains inputs to build a Config.
type ConfigOptions struct {
	ListenAddress  string
	UpstreamTarget string
	ChainID        string
	NodeKey        *p2p.NodeKey
	Action         cometbftAdapter.ByzantineAction
	Options        cometbftAdapter.ByzantineOptions
	Trigger        Trigger
	Hooks          Hooks
	Direction      Direction
	DialTimeout    time.Duration
	Logger         *slog.Logger
}

// NewConfig validates and normalises proxy options.
func NewConfig(opts ConfigOptions) (*Config, error) {
	listenNetwork, listenAddr, err := parseNetworkAddress(opts.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid listen address: %w", err)
	}
	upstreamNetwork, upstreamAddr, err := parseNetworkAddress(opts.UpstreamTarget)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream address: %w", err)
	}
	if opts.NodeKey == nil {
		return nil, fmt.Errorf("node key is required")
	}
	if strings.TrimSpace(opts.ChainID) == "" {
		return nil, fmt.Errorf("chain id is required")
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	trigger := Trigger{Step: strings.ToLower(strings.TrimSpace(opts.Trigger.Step))}
	if opts.Trigger.Height != nil {
		h := *opts.Trigger.Height
		trigger.Height = &h
	}
	if opts.Trigger.Round != nil {
		r := *opts.Trigger.Round
		trigger.Round = &r
	}

	cfg := &Config{
		ListenNetwork:   listenNetwork,
		ListenAddress:   listenAddr,
		UpstreamNetwork: upstreamNetwork,
		UpstreamAddress: upstreamAddr,
		ChainID:         strings.TrimSpace(opts.ChainID),
		NodeKey:         opts.NodeKey,
		Action:          opts.Action,
		Options:         opts.Options,
		Trigger:         trigger,
		Hooks:           opts.Hooks,
		Direction:       opts.Direction,
		DialTimeout:     opts.DialTimeout,
		Logger:          logger,
	}

	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 5 * time.Second
	}

	return cfg, nil
}

func parseNetworkAddress(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("address cannot be empty")
	}
	if !strings.Contains(raw, "://") {
		return "tcp", raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	if u.Scheme == "" {
		return "", "", fmt.Errorf("missing scheme")
	}
	if u.Host == "" {
		return "", "", fmt.Errorf("missing host")
	}
	return u.Scheme, u.Host, nil
}
