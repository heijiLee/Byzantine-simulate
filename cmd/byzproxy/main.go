package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cometbftAdapter "codec/cometbft/adapter"
	"codec/proxy/engine"

	"github.com/cometbft/cometbft/p2p"
)

func main() {
	var (
		listenAddr         = flag.String("listen", "tcp://0.0.0.0:26656", "address to accept external peers (tcp://host:port)")
		upstreamAddr       = flag.String("upstream", "tcp://127.0.0.1:26657", "address of the upstream validator (tcp://host:port)")
		nodeKeyPath        = flag.String("node-key", "", "path to the node_key.json used for the proxy handshake")
		chainID            = flag.String("chain-id", "proxy-chain", "chain identifier used for canonical mapping")
		attack             = flag.String("attack", string(cometbftAdapter.ByzantineActionNone), "byzantine action to apply")
		triggerHeight      = flag.Int64("trigger-height", 0, "height at which mutations activate (0 disables)")
		triggerRound       = flag.Int64("trigger-round", 0, "round at which mutations activate (0 disables)")
		triggerStep        = flag.String("trigger-step", "", "canonical message type (proposal|prevote|precommit) required for mutation")
		delayDur           = flag.Duration("delay", 0, "delay applied to triggered messages before forwarding")
		dropMessages       = flag.Bool("drop", false, "drop triggered messages instead of forwarding")
		duplicate          = flag.Bool("duplicate", false, "duplicate triggered messages after mutation")
		alternateBlock     = flag.String("alternate-block", "", "alternate block hash used during mutation")
		alternatePrev      = flag.String("alternate-prev-hash", "", "alternate previous block hash used during mutation")
		alternateSig       = flag.String("alternate-signature", "", "alternate signature for forged messages")
		alternateValidator = flag.String("alternate-validator", "", "alternate validator/proposer identifier")
		roundOffset        = flag.Int64("round-offset", 0, "offset applied to canonical round when mutating")
		heightOffset       = flag.Int64("height-offset", 0, "offset applied to canonical height when mutating")
		timestampShift     = flag.Duration("timestamp-skew", 0, "duration applied to canonical timestamps when mutating")
		dialTimeout        = flag.Duration("dial-timeout", 5*time.Second, "timeout used when dialing the upstream validator")
		mutateDir          = flag.String("mutate-direction", "upstream", "direction to apply mutations (upstream|downstream|both)")
	)

	flag.Parse()

	if strings.TrimSpace(*nodeKeyPath) == "" {
		fmt.Fprintln(os.Stderr, "--node-key is required")
		os.Exit(1)
	}

	nodeKey, err := p2p.LoadNodeKey(*nodeKeyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load node key: %v\n", err)
		os.Exit(1)
	}

	byzAction, err := cometbftAdapter.ParseByzantineAction(*attack)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid attack type: %v\n", err)
		os.Exit(1)
	}

	opts := cometbftAdapter.ByzantineOptions{
		AlternateBlockHash: *alternateBlock,
		AlternatePrevHash:  *alternatePrev,
		AlternateSignature: *alternateSig,
		AlternateValidator: *alternateValidator,
		RoundOffset:        *roundOffset,
		HeightOffset:       *heightOffset,
		TimestampShift:     *timestampShift,
	}

	trigger := engine.Trigger{}
	if *triggerHeight > 0 {
		trigger.Height = triggerHeight
	}
	if *triggerRound > 0 {
		trigger.Round = triggerRound
	}
	if step := strings.TrimSpace(*triggerStep); step != "" {
		trigger.Step = strings.ToLower(step)
	}

	hooks := engine.Hooks{
		Delay:     *delayDur,
		Drop:      *dropMessages,
		Duplicate: *duplicate,
	}

	direction, err := engine.ParseDirection(*mutateDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid mutate direction: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := engine.NewConfig(engine.ConfigOptions{
		ListenAddress:  *listenAddr,
		UpstreamTarget: *upstreamAddr,
		ChainID:        *chainID,
		NodeKey:        nodeKey,
		Action:         byzAction,
		Options:        opts,
		Trigger:        trigger,
		Hooks:          hooks,
		Direction:      direction,
		DialTimeout:    *dialTimeout,
		Logger:         logger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build config: %v\n", err)
		os.Exit(1)
	}

	eng := engine.New(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := eng.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "proxy exited with error: %v\n", err)
		os.Exit(1)
	}
}
