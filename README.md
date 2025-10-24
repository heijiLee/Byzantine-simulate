# PBFT Message Abstraction Playground

## Overview
This project explores how to consolidate PBFT-family consensus messages into a single **canonical representation** and then re-materialize them into the original wire format of each blockchain client. The current implementation centers on the **CometBFT** pipeline, and we are extending the same canonical model to cover **Kaia IBFT** and **Hyperledger Besu (IBFT/QBFT)** so that messages can be translated across engines without losing intent.

## Objectives
1. Convert PBFT-style consensus messages from multiple engines into a common `CanonicalMessage` format.
2. Inspect, mutate, or simulate the canonical data for research and fault-injection experiments.
3. Serialize the modified canonical messages back into chain-specific vote or proposal objects for validation or reinjection.

## Key Features
- **Canonical message model**: The `message/abstraction` package defines the shared structure that captures proposal, vote, precommit, and related PBFT semantics.
- **Chain-specific mappers**: Adapters in `cometbft/`, `kaia/`, and `hyperledger/besu/` implement the `Mapper` interface (`ToCanonical` / `FromCanonical`) to bridge native data structures with the canonical model.
- **Raw message wrappers**: On-chain WAL entries, RPC responses, or network packets can be wrapped into `RawConsensusMessage` for uniform processing.
- **Conversion simulators**: Utilities under `cmd/demo` demonstrate how real CometBFT messages round-trip through the canonical bridge.
- **Codec experiments**: The `message/codec` package contains JSON, Protobuf, RLP, and other serialization experiments that stress-test interoperability.

## Repository Layout
```
.
├── cmd/                # CLI tools and conversion demos
│   └── demo/           # CometBFT message simulator and round-trip checker
├── cometbft/           # CometBFT mapper and consensus adapters
├── hyperledger/besu/   # Besu IBFT/QBFT mapper (work in progress)
├── kaia/               # Kaia IBFT mapper (work in progress)
├── message/            # Canonical models, codecs, and protobuf definitions
└── examples/           # Sample WAL-derived consensus messages
```

## Quick Start
### 1. Verify prerequisites
```bash
go version    # Go 1.21 or later is recommended
protoc --version
```

### 2. Install dependencies
```bash
git clone <repository-url>
cd Byzantine-simulate
go mod tidy
```

### 3. Explore the CometBFT demo CLI
```bash
go run cmd/demo/main.go
```
- Lists the available scenarios (`simulation`, `vote-batch`, `byzantine`).
- `-scenario=simulation` streams synthetic CometBFT messages through the canonical mapper.
- `-scenario=vote-batch` replays fixtures from `examples/cometbft/Vote.json` and validates the round-trip.
- `-scenario=byzantine` forges mutated payloads via the **canonical → byz-canonical → byzcomet** pipeline and prints each stage of the mutation.
- Actions supported by the byzantine pipeline include `double_vote`, `double_proposal`, `alter_validator`, `drop_signature`, `timestamp_skew`, and `none`.
- Tunable flags such as `-alternate-block`, `-alternate-prev`, `-alternate-signature`, `-alternate-validator`, `-round-offset`, `-height-offset`, and `-timestamp-skew` control the resulting forged payloads.

Example explorations:

```bash
# Emit a conflicting prevote that bumps height/round and swaps the validator
go run cmd/demo/main.go -scenario=byzantine -action=alter_validator -alternate-validator=validator-9 -round-offset=1 -height-offset=2

# Produce a timestamp-skewed proposal with custom hashes
go run cmd/demo/main.go -scenario=byzantine -action=timestamp_skew -timestamp-skew=250ms -alternate-block=0xDEADBEEF -alternate-prev=0xFEEDFACE

# Generate a double vote while forcing an explicit signature override
go run cmd/demo/main.go -scenario=byzantine -action=double_vote -alternate-signature=fake-signature
```

To script the same pipeline, use `cmd/byzantine` which emits JSON containing both the byz-canonical mutations and their encoded CometBFT counterparts.

### 4. Execute tests
```bash
go test ./...
```
- Validates transformation logic, verification helpers, and simulator behaviors.

### 5. (Optional) Regenerate protobuf descriptors
```bash
protoc \
  --proto_path=message/proto \
  --descriptor_set_out=message/proto/abstraction.protoset \
  --include_imports --include_source_info \
  message/proto/abstraction.proto
```

## Canonical Flow
1. **Collect raw data**: Read WAL entries, RPC responses, or network packets and wrap them as `RawConsensusMessage` objects.
2. **Normalize**: Use a chain-specific mapper to convert the raw data into `CanonicalMessage` instances.
3. **Analyze or mutate**: Apply filtering, field edits, or re-signing steps against the canonical representation.
4. **Rehydrate**: Serialize the modified canonical messages back into the original chain-specific structures.
5. **Inject or simulate**: Feed the resulting messages into the consensus engine or exercise them on a controlled test network.

The CometBFT pipeline is production-ready, while Kaia and Besu adapters are being aligned with the canonical schema, signature semantics, and codec expectations.

## Additional Documentation
- `verify_vote_conversion.md`: Walkthrough of the CometBFT vote conversion experiment.
- `verify_conversion.md`: Canonical conversion rules and testing strategy overview.
- `message/README.md`: Usage notes for the codec experimentation tools.

## Contributing
1. Open an issue to discuss new ideas or report a bug.
2. Create a feature branch for your changes.
3. Run `go test ./...` to ensure the core suites pass.
4. Submit a pull request summarizing the change and test results.

## License
Refer to the repository's license file for detailed terms.
