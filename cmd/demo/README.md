# CometBFT Demo CLI

This CLI showcases how the PBFT canonical mapper powers different kinds of CometBFT experiments. It exposes three scenarios:

1. **simulation** – Streams randomly generated CometBFT messages through the canonical mapper so you can inspect the round-trip flow.
2. **vote-batch** – Replays fixtures from `examples/cometbft/Vote.json` and verifies that they survive a canonical round-trip.
3. **byzantine** – Loads a canonical message (either from a file or derived from the fixtures), materializes one or more **byz-canonical** mutations (double votes, proposal forks, validator swaps, signature drops, timestamp skews, etc.), and then re-encodes them into forged CometBFT payloads.

## Usage

```bash
# Show the available scenarios
go run cmd/demo/main.go

# Run the live simulator for 15 seconds
go run cmd/demo/main.go -scenario=simulation -duration=15s

# Replay and verify the bundled vote fixtures
go run cmd/demo/main.go -scenario=vote-batch

# Forge double proposals derived from the fixtures
go run cmd/demo/main.go -scenario=byzantine -action=double_proposal

# Swap the validator and bump round/height with a single command
go run cmd/demo/main.go -scenario=byzantine -action=alter_validator -alternate-validator=validator-9 -round-offset=1 -height-offset=2
```

You can provide your own canonical input for the byzantine scenario using `-canonical=/path/to/canonical.json`. Optional flags `-alternate-block`, `-alternate-prev`, `-alternate-signature`, `-alternate-validator`, `-round-offset`, `-height-offset`, and `-timestamp-skew` override the forged fields when you need explicit values. During execution the CLI prints the **canonical → byz-canonical → byzcomet** progression so you can inspect each stage of the mutation.

## Demonstrating the byzantine proxy

The `cmd/byzproxy` CLI reuses the same canonical-byzantine pipeline to mutate **live** consensus traffic. A lightweight way to observe the behaviour without a full network is to wire two local CometBFT nodes together through the proxy.

```bash
# Terminal 1 – start an upstream validator (or use the localnet harness)
cometbft start --home $HOME/.cometbft-upstream

# Terminal 2 – launch the proxy with a mutation that triggers on the next prevote
go run cmd/byzproxy/main.go \
  --listen tcp://0.0.0.0:36656 \
  --upstream tcp://127.0.0.1:26656 \
  --node-key $HOME/.byzproxy/config/node_key.json \
  --attack double_prevote \
  --trigger-height 1 \
  --trigger-step prevote \
  --duplicate

# Terminal 3 – run a peer that dials the proxy instead of the validator
cometbft start --home $HOME/.cometbft-peer --p2p.laddr tcp://0.0.0.0:46656 --proxy_app tcp://127.0.0.1:36656
```

The proxy logs every consensus envelope, highlighting when a trigger fired, how the canonical message was mutated, and whether the packet travelled upstream (towards the validator) or downstream (towards external peers). Adjust the `--delay`, `--drop`, `--mutate-direction`, and canonical override flags to experiment with different failure scenarios.

### Understanding the byzantine pipeline

1. Start with a canonical vote or proposal (from fixtures or your own data).
2. Call `ApplyByzantineCanonical` to emit one or more mutated **byz-canonical** payloads.
3. Feed each byz-canonical payload through `mapper.FromCanonical` to obtain the forged CometBFT (`byzcomet`) messages.

The `cmd/demo` byzantine scenario prints each step, while `cmd/byzantine` can serialize the same structure as JSON for scripting or archival.

## Output snapshot

```
PBFT Message Abstraction Demo
============================

Scenarios:
  - simulation: Stream randomly generated CometBFT messages through the canonical mapper.
  - vote-batch: Replay vote samples from examples/cometbft/Vote.json and verify round-trips.
  - byzantine:  Emit forged CometBFT payloads from a canonical message using the byzantine pipeline.

Example usage:
  go run cmd/demo/main.go -scenario=simulation -duration=15s
  go run cmd/demo/main.go -scenario=vote-batch
  go run cmd/demo/main.go -scenario=byzantine -action=double_proposal

🧪 CometBFT Vote Round-Trip
===========================
Loaded vote fixtures from examples/cometbft/Vote.json

Case 1 → Prevote for Block
-----------
   Fixture → Raw consensus message
   Raw → Canonical
   Canonical → Raw
   Comparing original and converted payloads
   Summary
      Type: prevote
      Height: 882281
      Round: 0
      Block hash: 6D895…
      Validator: cosmosvalcons1...
      Extension count: 0
Result: success

Summary: 6/6 cases succeeded (100.0%).
```
