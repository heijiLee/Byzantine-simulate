# Byzantine Message Bridge

A comprehensive message transformation and routing system for blockchain consensus protocols, supporting CometBFT, Hyperledger Fabric/Besu, and Kaia networks.

## 🎯 Overview

The Byzantine Message Bridge provides bidirectional message transformation between different blockchain consensus protocols through a standardized `CanonicalMessage` format, enabling:

- **Protocol normalization**: Convert chain-specific messages to a unified `CanonicalMessage` format
- **Independent chain mapping**: Each blockchain has its own independent mapper (Chain ↔ Canonical)
- **Message validation**: Ensure message integrity and apply chain-specific constraints
- **Extensible architecture**: Easy integration of new blockchain protocols
- **Real message simulation**: Generate and test with realistic blockchain message patterns

## 🏗️ Architecture

### Core Components

| Component | Description |
|-----------|-------------|
| **Chain Adapters** | Protocol-specific mappers (`cometbft/`, `hyperledger/`, `kaia/`) |
| **Message Abstraction** | Unified message format (`CanonicalMessage`) and validation |
| **Message Examples** | Real blockchain message patterns for testing |
| **Message Parser** | JSON-based message file parsing and conversion |
| **Configuration** | YAML-based chain and routing configuration |

### Supported Protocols

- **CometBFT**: Tendermint consensus with Proposal/Prevote/Precommit messages
- **Hyperledger Fabric**: PBFT-style consensus with channel-specific messages  
- **Hyperledger Besu**: IBFT2 consensus with Ethereum-compatible messages
- **Kaia**: Istanbul BFT consensus with governance features

## 🔄 Message Flow Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  CometBFT   │    │ Hyperledger │    │ Hyperledger │    │    Kaia     │
│             │    │   Fabric    │    │    Besu     │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ CometBFT    │    │ Fabric      │    │ Besu        │    │ Kaia        │
│ Mapper      │    │ Mapper      │    │ Mapper      │    │ Mapper      │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Canonical Message                                │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ • ChainID: string                                               │   │
│  │ • Height: int64                                                 │   │
│  │ • Round: int32                                                  │   │
│  │ • Type: MsgType                                                 │   │
│  │ • Proposer/Validator: string                                     │   │
│  │ • BlockHash: string                                             │   │
│  │ • Extensions: map[string]interface{} (체인별 특화 데이터)        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Core Principles

1. **Independent Mappers**: Each blockchain has its own independent mapper
   - CometBFT Mapper: CometBFT ↔ Canonical
   - Fabric Mapper: Fabric ↔ Canonical
   - Besu Mapper: Besu ↔ Canonical
   - Kaia Mapper: Kaia ↔ Canonical

2. **Canonical Message Centric**: All transformations go through Canonical Message
   - Canonical Message is the standardized intermediate format
   - Extensions field preserves chain-specific data

3. **Simple Conversion Logic**: Each mapper implements only two methods
   - `ToCanonical()`: Chain message → Canonical
   - `FromCanonical()`: Canonical → Chain message

## 🚀 Quick Start

### Prerequisites
- Go 1.23.6+
- Protocol Buffers compiler (`protoc`)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd Byzantine-simulate

# Install dependencies
go mod tidy
```

### Step-by-Step Usage Guide

#### 1. Generate Message Examples
```bash
# Create realistic CometBFT message examples
go run cmd/demo/message_example_generator.go
```
This creates:
- `examples/cometbft/samples.json` - 6 sample messages
- `examples/cometbft/all_messages.json` - 15 complete messages
- `examples/cometbft/[MessageType].json` - Type-specific files

#### 2. Test Message Conversion
```bash
# Test sample messages (recommended for beginners)
go run cmd/demo/message_file_parser.go examples/cometbft/samples.json

# Test specific message type
go run cmd/demo/message_file_parser.go examples/cometbft/Vote.json

# Test all messages
go run cmd/demo/message_file_parser.go examples/cometbft/all_messages.json
```

#### 3. Run Integration Tests
```bash
# CometBFT advanced mapper tests
go test -v cometbft_advanced_test.go

# CometBFT integration tests
go test -v cometbft_integration_test.go

# Performance benchmarks
go test -bench=. -benchmem benchmark_test.go
```

#### 4. Simulate Real Consensus Flow
```bash
# Simulate real CometBFT consensus process
go run cmd/demo/real_message_simulator.go
```

## 💡 Practical Usage Examples

### Example 1: Convert CometBFT Message to Canonical
```go
package main

import (
    "fmt"
    "codec/message/abstraction"
    cometbftAdapter "codec/cometbft/adapter"
)

func main() {
    // Create CometBFT mapper
    mapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")
    
    // Create a CometBFT raw message
    rawMsg := abstraction.RawConsensusMessage{
        ChainType:   abstraction.ChainTypeCometBFT,
        ChainID:     "cosmos-hub-4",
        MessageType: "Proposal",
        Payload:     []byte(`{"height":1000000,"round":0,"proposer_address":"cosmos1abc123..."}`),
        Encoding:    "json",
    }
    
    // Convert to Canonical
    canonical, err := mapper.ToCanonical(rawMsg)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Height: %v\n", canonical.Height)
    fmt.Printf("Proposer: %s\n", canonical.Proposer)
    fmt.Printf("Type: %s\n", canonical.Type)
}
```

### Example 2: Convert Canonical to Fabric
```go
package main

import (
    "fmt"
    "codec/message/abstraction"
    fabricAdapter "codec/hyperledger/fabric/adapter"
)

func main() {
    // Create Fabric mapper
    mapper := fabricAdapter.NewFabricMapper("channel1")
    
    // Create a Canonical message
    canonical := &abstraction.CanonicalMessage{
        ChainID:    "channel1",
        Height:     big.NewInt(1000000),
        Round:      big.NewInt(0),
        Type:       abstraction.MsgTypeProposal,
        Proposer:   "peer0.org1.example.com",
        BlockHash:  "0x1234567890abcdef",
        Extensions: map[string]interface{}{
            "channel_id": "channel1",
        },
    }
    
    // Convert to Fabric
    fabricMsg, err := mapper.FromCanonical(canonical)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Fabric Message Type: %s\n", fabricMsg.MessageType)
    fmt.Printf("Channel: %s\n", fabricMsg.ChainID)
}
```

### Example 3: Cross-Chain Message Bridge
```go
package main

import (
    "fmt"
    "codec/message/abstraction"
    cometbftAdapter "codec/cometbft/adapter"
    fabricAdapter "codec/hyperledger/fabric/adapter"
)

func main() {
    // Create mappers
    cometbftMapper := cometbftAdapter.NewCometBFTMapper("cosmos-hub-4")
    fabricMapper := fabricAdapter.NewFabricMapper("channel1")
    
    // Step 1: CometBFT → Canonical
    cometbftMsg := createCometBFTMessage()
    canonical, err := cometbftMapper.ToCanonical(cometbftMsg)
    if err != nil {
        panic(err)
    }
    
    // Step 2: Canonical → Fabric
    fabricMsg, err := fabricMapper.FromCanonical(canonical)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Successfully converted CometBFT → Fabric\n")
    fmt.Printf("Original: %s\n", cometbftMsg.MessageType)
    fmt.Printf("Target: %s\n", fabricMsg.MessageType)
}
```

### Configuration

Edit `configs/bridge.yaml` to configure chains and routing rules:

```yaml
chains:
  - name: cometbft
    enabled: true
    endpoint: grpc://localhost:9090
    ingress:
      type: collector
      decoder: proto
    egress:
      targets:
        - type: kafka
          topic: cometbft.consensus

router:
  rules:
    - match: {chain: cometbft, message_type: vote}
      forward:
        - chain: fabric
        - sink: kafka://consensus.vote
```

## 📁 Project Structure

```
Byzantine-simulate/
├── cometbft/
│   ├── adapter/              # CometBFT message mapper
│   └── consensus_engine.go   # CometBFT consensus simulation
├── hyperledger/
│   ├── fabric/adapter/       # Fabric message mapper
│   └── besu/adapter/         # Besu message mapper
├── kaia/adapter/             # Kaia message mapper
├── message/
│   └── abstraction/          # Core message types and validation
├── cmd/demo/                 # Demo applications
│   ├── message_example_generator.go  # Generate message examples
│   ├── message_file_parser.go       # Parse message files
│   ├── real_message_simulator.go    # Real message simulation
│   └── wal_reader.go                # WAL file reader
├── examples/cometbft/        # CometBFT message examples
│   ├── samples.json          # Sample messages
│   ├── all_messages.json     # All message types
│   └── [MessageType].json    # Type-specific messages
└── configs/
    └── bridge.yaml           # Configuration file
```

## 🚀 Quick Demo (5 minutes)

Want to see it in action? Follow these steps:

```bash
# 1. Generate realistic message examples
go run cmd/demo/message_example_generator.go

# 2. Test message conversion (you'll see 100% success rate!)
go run cmd/demo/message_file_parser.go examples/cometbft/samples.json

# 3. Run integration tests
go test -v cometbft_integration_test.go

# 4. Simulate real consensus flow
go run cmd/demo/real_message_simulator.go
```

Expected output:
- ✅ 6/6 messages successfully converted to Canonical format
- ✅ All CometBFT message types supported
- ✅ Cross-chain conversion working
- ✅ Real consensus protocol simulation

## 📦 Message Examples & Testing

### Generate Message Examples
```bash
# Generate realistic CometBFT message examples
go run cmd/demo/message_example_generator.go
```

This creates JSON files with realistic blockchain message patterns:
- `examples/cometbft/samples.json` - 6 sample messages (one per type)
- `examples/cometbft/all_messages.json` - 15 complete messages
- `examples/cometbft/[MessageType].json` - Type-specific messages

### Parse and Test Messages
```bash
# Test sample messages
go run cmd/demo/message_file_parser.go examples/cometbft/samples.json

# Test specific message type
go run cmd/demo/message_file_parser.go examples/cometbft/Vote.json

# Test all messages
go run cmd/demo/message_file_parser.go examples/cometbft/all_messages.json
```

### Real Message Simulation
```bash
# Simulate real CometBFT consensus flow
go run cmd/demo/real_message_simulator.go
```

## 🔄 Message Flow

1. **Message Generation**: Create realistic blockchain message patterns
2. **Normalization**: Convert to `CanonicalMessage` format using chain-specific mappers
3. **Validation**: Apply chain-specific validation rules
4. **Testing**: Verify conversion accuracy and data preservation
5. **Cross-chain**: Convert Canonical messages to other blockchain formats

## 📋 Message Types

### CanonicalMessage
```go
type CanonicalMessage struct {
    ChainID    string                 // Chain identifier
    Height     *big.Int              // Block height
    Round      *big.Int              // Consensus round
    Type       MsgType               // Message type
    Timestamp  time.Time             // Creation time
    BlockHash  string                // Block hash
    Proposer   string                // Proposer ID
    Validator  string                // Validator ID
    Signature  string                // Message signature
    Extensions map[string]interface{} // Chain-specific data
}
```

### Supported Message Types
- `proposal` - Block proposals
- `vote` - Consensus votes
- `prepare` - PBFT prepare messages
- `commit` - PBFT commit messages
- `view_change` - View change messages
- `block` - Block data

## 🛠️ Development

### Adding a New Chain

1. **Create adapter**: Implement `Mapper` interface in `{chain}/adapter/`
2. **Define message types**: Add chain-specific message structures
3. **Add validation**: Create validation rules in `validator/`
4. **Update configuration**: Add chain config to `bridge.yaml`
5. **Test**: Run integration tests

### Example: Adding Ethereum

```go
// ethereum/adapter/mapper.go
type EthereumMapper struct {
    chainID string
}

func (m *EthereumMapper) ToCanonical(raw RawConsensusMessage) (*CanonicalMessage, error) {
    // Convert Ethereum message to canonical format
}

func (m *EthereumMapper) FromCanonical(msg *CanonicalMessage) (*RawConsensusMessage, error) {
    // Convert canonical message to Ethereum format
}
```

## 🧪 Testing & Experiments

### Integration Tests
```bash
# CometBFT advanced mapper tests
go test -v cometbft_advanced_test.go

# CometBFT integration tests
go test -v cometbft_integration_test.go

# Performance benchmarks
go test -bench=. -benchmem benchmark_test.go
```

### Real Message Testing
```bash
# Generate and test realistic messages
go run cmd/demo/message_example_generator.go
go run cmd/demo/message_file_parser.go examples/cometbft/samples.json

# Simulate real consensus flow
go run cmd/demo/real_message_simulator.go
```

### CometBFT Protocol Implementation

This project implements real CometBFT protocol buffer structures:

#### Supported Message Types
- **NewRoundStep**: Consensus round step transitions
- **Proposal**: Block proposals
- **Vote**: Prevote/Precommit votes
- **BlockPart**: Block part transmission
- **NewValidBlock**: Valid block notifications
- **VoteSetBits**: Vote bitmap messages
- **HasVote**: Vote receipt confirmations
- **ProposalPOL**: Proposal POL evidence

#### Real CometBFT Structures
- **BlockID**: Hash and PartSetHeader inclusion
- **PartSetHeader**: Block part information
- **SignedMsgType**: Prevote/Precommit distinction
- **ValidatorSet**: Validator list and voting power
- **ConsensusState**: Height, round, step state

#### Consensus Engine Simulation
- Real CometBFT consensus protocol flow implementation
- Validator voting power-based consensus simulation
- Round-robin proposer selection
- Block finalization process

### What You Can Verify in Tests

1. **Message Conversion Accuracy**: Each chain's messages correctly convert to Canonical format
2. **Cross-chain Compatibility**: CometBFT messages can be converted to Fabric, Besu, Kaia
3. **Data Preservation**: Core data like height, round, timestamp are preserved
4. **Validation System**: Invalid messages are properly rejected
5. **Performance**: Conversion speed and memory usage
6. **Real-time Processing**: Continuous message stream processing capability
7. **CometBFT Protocol**: Real consensus protocol flow simulation
8. **Validator Management**: Voting power-based consensus simulation

## 📊 Monitoring

The bridge provides metrics for:
- Message processing rates
- Validation failures
- Cross-chain routing success/failure
- Chain connectivity status

## 🔧 Configuration Reference

### Chain Configuration
- `name`: Chain identifier
- `enabled`: Enable/disable chain
- `endpoint`: Connection endpoint
- `ingress`: Input configuration
- `egress`: Output targets

### Routing Rules
- `match`: Message matching criteria
- `forward`: Target destinations

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Submit a pull request

## 📄 License

[License information]

## 🎉 Key Benefits

1. **Simplicity**: Each mapper handles only one blockchain
2. **Scalability**: Easy to add new blockchains with independent mappers
3. **Maintainability**: Chain-specific logic is separated for easy management
4. **Testability**: Each mapper can be tested independently
5. **Performance**: Direct conversion minimizes intermediate steps
6. **Real-world Testing**: Uses realistic blockchain message patterns

## 🆘 Support

For issues and questions:
- Create an issue in the repository
- Check the documentation
- Review test cases for examples
- Use message examples for testing: `examples/cometbft/`
