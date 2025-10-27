# Byzantine Proxy CLI

The `byzproxy` command sits between CometBFT peers and an upstream validator, forwarding consensus traffic while optionally mutating envelopes according to the canonical-byzantine pipeline.

## Features

- Establishes a secure connection to the upstream validator via `MakeSecretConnection` using the supplied node key.
- Mirrors the consensus channels (proposal, block-part, vote, vote-set-bits) so inbound and outbound packets remain in sync.
- Converts each consensus message into the canonical representation, applies a configured `ByzantineAction`, then re-encodes it before forwarding.
- Supports drop, delay, and duplicate hooks that activate once the configured height/round/step trigger matches the envelope metadata.
- Emits structured JSON logs and Prometheus metrics describing the forwarding and mutation lifecycle.

## Usage

```bash
# Basic proxy that mutates prevotes at height 100
byzproxy \
  --listen tcp://0.0.0.0:26656 \
  --upstream tcp://127.0.0.1:26657 \
  --node-key $HOME/.byzproxy/config/node_key.json \
  --attack double_prevote \
  --trigger-height 100 \
  --trigger-step prevote
```

Useful flags:

- `--trigger-round`: Require a specific round before firing the mutation.
- `--mutate-direction`: `upstream`, `downstream`, or `both` to control where mutations apply.
- `--delay`, `--drop`, `--duplicate`: Runtime hooks for delaying, dropping, or duplicating triggered envelopes.
- `--alternate-block`, `--alternate-prev-hash`, `--alternate-signature`, `--alternate-validator`: Override canonical fields used during mutation.
- `--round-offset`, `--height-offset`, `--timestamp-skew`: Adjust consensus metadata when forging payloads.

The binary exits with a non-zero status when configuration or runtime errors occur. All operational logs are emitted as JSON to `stdout` and can be scraped for auditing or analysis.

## Development

Run the proxy directly with Go:

```bash
go run cmd/byzproxy/main.go --node-key $HOME/.byzproxy/config/node_key.json
```

Unit tests for the proxy live in `proxy/engine/engine_test.go` and use in-memory pipes to cover duplication, dropping, and delayed proposal scenarios.
