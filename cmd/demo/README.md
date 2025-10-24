# CometBFT Demo CLI

This CLI showcases how the PBFT canonical mapper powers different kinds of CometBFT experiments. It exposes three scenarios:

1. **simulation** – Streams randomly generated CometBFT messages through the canonical mapper so you can inspect the round-trip flow.
2. **vote-batch** – Replays fixtures from `examples/cometbft/Vote.json` and verifies that they survive a canonical round-trip.
3. **byzantine** – Loads a canonical message (either from a file or derived from the fixtures) and emits forged CometBFT payloads using the byzantine pipeline.

## Usage

```bash
# Show the available scenarios
go run cmd/demo/main.go

# Run the live simulator for 15 seconds
go run cmd/demo/main.go -scenario=simulation -duration=15s

# Replay and verify the bundled vote fixtures
go run cmd/demo/main.go -scenario=vote-batch

# Forge double proposals derived from the fixtures
go run cmd/demo/main.go -scenario=byzantine -action=double-proposal
```

You can provide your own canonical input for the byzantine scenario using `-canonical=/path/to/canonical.json`. Optional flags `-alternate-block`, `-alternate-prev`, and `-alternate-signature` override the forged fields when you need explicit values.

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
  go run cmd/demo/main.go -scenario=byzantine -action=double-proposal

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
