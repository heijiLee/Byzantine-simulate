package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"byzantine-message-bridge/message/abstraction"
	"byzantine-message-bridge/message/abstraction/validator"

	cometbftAdapter "byzantine-message-bridge/cometbft/adapter"
	besuAdapter "byzantine-message-bridge/hyperledger/besu/adapter"
	fabricAdapter "byzantine-message-bridge/hyperledger/fabric/adapter"
	kaiaAdapter "byzantine-message-bridge/kaia/adapter"
)

// BridgeConfig represents the configuration for the message bridge
type BridgeConfig struct {
	Chains []ChainConfig `json:"chains"`
	Router RouterConfig  `json:"router"`
}

// ChainConfig represents configuration for a specific chain
type ChainConfig struct {
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Endpoint string                 `json:"endpoint"`
	Ingress  IngressConfig          `json:"ingress"`
	Egress   EgressConfig           `json:"egress"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// IngressConfig represents ingress configuration
type IngressConfig struct {
	Type    string `json:"type"`
	Decoder string `json:"decoder"`
}

// EgressConfig represents egress configuration
type EgressConfig struct {
	Targets []EgressTarget `json:"targets"`
}

// EgressTarget represents an egress target
type EgressTarget struct {
	Type   string                 `json:"type"`
	Topic  string                 `json:"topic,omitempty"`
	Chain  string                 `json:"chain,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// RouterConfig represents router configuration
type RouterConfig struct {
	Rules []RoutingRule `json:"rules"`
}

// RoutingRule represents a routing rule
type RoutingRule struct {
	Match   MatchCondition  `json:"match"`
	Forward []ForwardTarget `json:"forward"`
}

// MatchCondition represents matching conditions
type MatchCondition struct {
	Chain       string `json:"chain,omitempty"`
	MessageType string `json:"message_type,omitempty"`
}

// ForwardTarget represents a forwarding target
type ForwardTarget struct {
	Chain string `json:"chain,omitempty"`
	Sink  string `json:"sink,omitempty"`
}

// MessageBridge orchestrates message collection, normalization, and routing
type MessageBridge struct {
	config     BridgeConfig
	mappers    map[string]abstraction.Mapper
	validators map[string]*validator.Validator
	rules      []RoutingRule
}

// NewMessageBridge creates a new message bridge
func NewMessageBridge(config BridgeConfig) *MessageBridge {
	bridge := &MessageBridge{
		config:     config,
		mappers:    make(map[string]abstraction.Mapper),
		validators: make(map[string]*validator.Validator),
		rules:      config.Router.Rules,
	}

	// Initialize mappers for each enabled chain
	for _, chainConfig := range config.Chains {
		if chainConfig.Enabled {
			bridge.initializeMapper(chainConfig)
		}
	}

	return bridge
}

// initializeMapper initializes a mapper for a specific chain
func (mb *MessageBridge) initializeMapper(config ChainConfig) {
	var mapper abstraction.Mapper
	var chainType abstraction.ChainType

	switch config.Name {
	case "cometbft":
		chainType = abstraction.ChainTypeCometBFT
		mapper = cometbftAdapter.NewCometBFTMapper(config.Endpoint)
	case "fabric":
		chainType = abstraction.ChainTypeHyperledger
		mapper = fabricAdapter.NewFabricMapper(config.Endpoint)
	case "besu":
		chainType = abstraction.ChainTypeHyperledger
		mapper = besuAdapter.NewBesuMapper(config.Endpoint)
	case "kaia":
		chainType = abstraction.ChainTypeKaia
		mapper = kaiaAdapter.NewKaiaMapper(config.Endpoint)
	default:
		log.Printf("Unknown chain type: %s", config.Name)
		return
	}

	mb.mappers[config.Name] = mapper
	mb.validators[config.Name] = validator.NewValidator(chainType)
	log.Printf("Initialized mapper for chain: %s", config.Name)
}

// ProcessMessage processes a raw consensus message
func (mb *MessageBridge) ProcessMessage(raw abstraction.RawConsensusMessage) error {
	// Find the appropriate mapper
	mapper, exists := mb.mappers[raw.ChainID]
	if !exists {
		return fmt.Errorf("no mapper found for chain: %s", raw.ChainID)
	}

	// Convert to canonical format
	canonical, err := mapper.ToCanonical(raw)
	if err != nil {
		return fmt.Errorf("failed to convert to canonical: %v", err)
	}

	// Validate the canonical message
	validator, exists := mb.validators[raw.ChainID]
	if exists {
		if err := validator.Validate(canonical); err != nil {
			return fmt.Errorf("validation failed: %v", err)
		}
	}

	// Apply routing rules
	if err := mb.routeMessage(canonical); err != nil {
		return fmt.Errorf("routing failed: %v", err)
	}

	log.Printf("Successfully processed message: chain=%s, type=%s, height=%v",
		canonical.ChainID, canonical.Type, canonical.Height)
	return nil
}

// routeMessage applies routing rules to a canonical message
func (mb *MessageBridge) routeMessage(msg *abstraction.CanonicalMessage) error {
	for _, rule := range mb.rules {
		if mb.matchesRule(msg, rule.Match) {
			for _, target := range rule.Forward {
				if err := mb.forwardMessage(msg, target); err != nil {
					log.Printf("Failed to forward message: %v", err)
				}
			}
		}
	}
	return nil
}

// matchesRule checks if a message matches a routing rule
func (mb *MessageBridge) matchesRule(msg *abstraction.CanonicalMessage, match MatchCondition) bool {
	if match.Chain != "" && match.Chain != msg.ChainID {
		return false
	}
	if match.MessageType != "" && match.MessageType != string(msg.Type) {
		return false
	}
	return true
}

// forwardMessage forwards a message to a target
func (mb *MessageBridge) forwardMessage(msg *abstraction.CanonicalMessage, target ForwardTarget) error {
	if target.Chain != "" {
		// Forward to another chain
		return mb.forwardToChain(msg, target.Chain)
	}
	if target.Sink != "" {
		// Forward to a sink (e.g., Kafka, file)
		return mb.forwardToSink(msg, target.Sink)
	}
	return fmt.Errorf("no valid target specified")
}

// forwardToChain forwards a message to another chain
func (mb *MessageBridge) forwardToChain(msg *abstraction.CanonicalMessage, targetChain string) error {
	mapper, exists := mb.mappers[targetChain]
	if !exists {
		return fmt.Errorf("no mapper found for target chain: %s", targetChain)
	}

	// Convert canonical message to target chain format
	raw, err := mapper.FromCanonical(msg)
	if err != nil {
		return fmt.Errorf("failed to convert to target chain format: %v", err)
	}

	log.Printf("Forwarded message to chain %s: type=%s, height=%v",
		targetChain, raw.MessageType, msg.Height)
	return nil
}

// forwardToSink forwards a message to a sink
func (mb *MessageBridge) forwardToSink(msg *abstraction.CanonicalMessage, sink string) error {
	// In a real implementation, this would send to Kafka, write to file, etc.
	log.Printf("Forwarded message to sink %s: chain=%s, type=%s, height=%v",
		sink, msg.ChainID, msg.Type, msg.Height)
	return nil
}

// GetSupportedChains returns the list of supported chains
func (mb *MessageBridge) GetSupportedChains() []string {
	var chains []string
	for name := range mb.mappers {
		chains = append(chains, name)
	}
	return chains
}

// GetChainInfo returns information about a specific chain
func (mb *MessageBridge) GetChainInfo(chainName string) (map[string]interface{}, error) {
	mapper, exists := mb.mappers[chainName]
	if !exists {
		return nil, fmt.Errorf("chain not found: %s", chainName)
	}

	info := map[string]interface{}{
		"chain_type":      mapper.GetChainType(),
		"supported_types": mapper.GetSupportedTypes(),
	}

	return info, nil
}

func main() {
	// Load configuration
	configFile := "configs/bridge.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create message bridge
	bridge := NewMessageBridge(config)

	// Print supported chains
	fmt.Println("Supported chains:")
	for _, chain := range bridge.GetSupportedChains() {
		info, err := bridge.GetChainInfo(chain)
		if err != nil {
			log.Printf("Failed to get chain info for %s: %v", chain, err)
			continue
		}
		fmt.Printf("  %s: %v\n", chain, info)
	}

	// Run demo with sample messages
	runDemo(bridge)
}

// loadConfig loads configuration from a file
func loadConfig(filename string) (BridgeConfig, error) {
	var config BridgeConfig

	// For demo purposes, create a default config
	config = BridgeConfig{
		Chains: []ChainConfig{
			{
				Name:     "cometbft",
				Enabled:  true,
				Endpoint: "testnet-cometbft",
				Ingress: IngressConfig{
					Type:    "collector",
					Decoder: "proto",
				},
				Egress: EgressConfig{
					Targets: []EgressTarget{
						{Type: "kafka", Topic: "cometbft.consensus"},
					},
				},
			},
			{
				Name:     "fabric",
				Enabled:  true,
				Endpoint: "testnet-fabric",
				Ingress: IngressConfig{
					Type:    "collector",
					Decoder: "proto",
				},
				Egress: EgressConfig{
					Targets: []EgressTarget{
						{Type: "kafka", Topic: "fabric.consensus"},
					},
				},
			},
			{
				Name:     "kaia",
				Enabled:  true,
				Endpoint: "testnet-kaia",
				Ingress: IngressConfig{
					Type:    "collector",
					Decoder: "rlp",
				},
				Egress: EgressConfig{
					Targets: []EgressTarget{
						{Type: "kafka", Topic: "kaia.consensus"},
					},
				},
			},
		},
		Router: RouterConfig{
			Rules: []RoutingRule{
				{
					Match: MatchCondition{
						Chain:       "cometbft",
						MessageType: "vote",
					},
					Forward: []ForwardTarget{
						{Chain: "fabric"},
						{Sink: "kafka://consensus.vote"},
					},
				},
			},
		},
	}

	return config, nil
}

// runDemo runs a demonstration of the message bridge
func runDemo(bridge *MessageBridge) {
	fmt.Println("\n=== Message Bridge Demo ===")

	// Create sample messages for each chain
	sampleMessages := []abstraction.RawConsensusMessage{
		{
			ChainType:   abstraction.ChainTypeCometBFT,
			ChainID:     "cometbft",
			MessageType: "Proposal",
			Payload:     []byte(`{"height":1000,"round":1,"type":"Proposal","block_hash":"0xabc123","proposer":"node1"}`),
			Encoding:    "json",
			Timestamp:   time.Now(),
		},
		{
			ChainType:   abstraction.ChainTypeHyperledger,
			ChainID:     "fabric",
			MessageType: "PROPOSAL",
			Payload:     []byte(`{"block_number":1000,"type":"PROPOSAL","block_hash":"0xdef456","proposer":"peer1","channel_id":"mychannel"}`),
			Encoding:    "json",
			Timestamp:   time.Now(),
		},
		{
			ChainType:   abstraction.ChainTypeKaia,
			ChainID:     "kaia",
			MessageType: "PROPOSAL",
			Payload:     []byte(`{"block_number":1000,"round_number":1,"type":"PROPOSAL","block_hash":"0x789abc","proposer":"validator1","gas_limit":8000000}`),
			Encoding:    "json",
			Timestamp:   time.Now(),
		},
	}

	// Process each sample message
	for i, rawMsg := range sampleMessages {
		fmt.Printf("\n--- Processing Message %d ---\n", i+1)
		fmt.Printf("Chain: %s, Type: %s\n", rawMsg.ChainID, rawMsg.MessageType)

		if err := bridge.ProcessMessage(rawMsg); err != nil {
			log.Printf("Failed to process message: %v", err)
		}
	}

	fmt.Println("\n=== Demo Complete ===")
}
