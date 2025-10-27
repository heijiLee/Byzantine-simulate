package engine

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	cometbftAdapter "codec/cometbft/adapter"
	p2pconn "github.com/cometbft/cometbft/p2p/conn"
)

// Engine runs the proxy.
type Engine struct {
	cfg     *Config
	mapper  *cometbftAdapter.CometBFTMapper
	metrics *Metrics
}

// New constructs a proxy engine from the configuration.
func New(cfg *Config) *Engine {
	if cfg == nil {
		panic("engine config cannot be nil")
	}
	mapper := cometbftAdapter.NewCometBFTMapper(cfg.ChainID)
	return &Engine{
		cfg:     cfg,
		mapper:  mapper,
		metrics: NewMetrics(),
	}
}

// Run starts accepting peers until the context is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	ln, err := net.Listen(e.cfg.ListenNetwork, e.cfg.ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s://%s: %w", e.cfg.ListenNetwork, e.cfg.ListenAddress, err)
	}
	e.cfg.Logger.Info("proxy listening", "network", e.cfg.ListenNetwork, "address", e.cfg.ListenAddress)

	var wg sync.WaitGroup
	defer func() {
		_ = ln.Close()
		wg.Wait()
	}()

	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				e.cfg.Logger.Warn("temporary accept error", "err", err)
				continue
			}
			errCh <- err
			break
		}

		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			if err := e.handleConnection(ctx, c); err != nil {
				if ctx.Err() == nil {
					e.cfg.Logger.Error("connection handler exited", "err", err, "remote", c.RemoteAddr().String())
				}
			}
		}(conn)
	}

	select {
	case err := <-errCh:
		return err
	default:
		return ctx.Err()
	}
}

func (e *Engine) handleConnection(ctx context.Context, downstreamConn net.Conn) error {
	defer downstreamConn.Close()
	remote := downstreamConn.RemoteAddr().String()
	e.cfg.Logger.Info("accepted peer", "remote", remote)

	downstreamSecret, err := p2pconn.MakeSecretConnection(downstreamConn, e.cfg.NodeKey.PrivKey)
	if err != nil {
		return fmt.Errorf("handshake with downstream peer failed: %w", err)
	}

	upstreamConn, err := net.DialTimeout(e.cfg.UpstreamNetwork, e.cfg.UpstreamAddress, e.cfg.DialTimeout)
	if err != nil {
		return fmt.Errorf("failed to dial upstream %s://%s: %w", e.cfg.UpstreamNetwork, e.cfg.UpstreamAddress, err)
	}

	upstreamSecret, err := p2pconn.MakeSecretConnection(upstreamConn, e.cfg.NodeKey.PrivKey)
	if err != nil {
		upstreamConn.Close()
		return fmt.Errorf("handshake with upstream failed: %w", err)
	}

	sessionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sess := newSession(sessionCtx, cancel, e.cfg, e.mapper, e.metrics, downstreamSecret, upstreamSecret)
	if err := sess.run(); err != nil {
		if !strings.Contains(err.Error(), "closed network connection") {
			return err
		}
	}
	return nil
}
