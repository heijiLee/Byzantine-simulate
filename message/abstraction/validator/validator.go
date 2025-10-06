package validator

import (
	"fmt"
	"math/big"
	"time"

	"codec/message/abstraction"
)

// Validator validates canonical messages and enforces chain-specific constraints
type Validator struct {
	chainType abstraction.ChainType
	rules     ValidationRules
}

// ValidationRules defines validation rules for a specific chain
type ValidationRules struct {
	RequiredFields []string               `json:"required_fields"`
	FieldTypes     map[string]string      `json:"field_types"`
	Constraints    map[string]interface{} `json:"constraints"`
	CustomRules    []CustomValidationRule `json:"custom_rules"`
}

// CustomValidationRule defines a custom validation function
type CustomValidationRule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Function    func(*abstraction.CanonicalMessage) error
}

// NewValidator creates a new validator for the specified chain type
func NewValidator(chainType abstraction.ChainType) *Validator {
	return &Validator{
		chainType: chainType,
		rules:     getDefaultRules(chainType),
	}
}

// Validate validates a canonical message against chain-specific rules
func (v *Validator) Validate(msg *abstraction.CanonicalMessage) error {
	if msg == nil {
		return &abstraction.MessageValidationError{
			Field:   "message",
			Message: "message cannot be nil",
			Code:    "MISSING_FIELD",
		}
	}

	// Validate required fields
	if err := v.validateRequiredFields(msg); err != nil {
		return err
	}

	// Validate field types
	if err := v.validateFieldTypes(msg); err != nil {
		return err
	}

	// Validate constraints
	if err := v.validateConstraints(msg); err != nil {
		return err
	}

	// Run custom validation rules
	if err := v.runCustomRules(msg); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields checks that all required fields are present
func (v *Validator) validateRequiredFields(msg *abstraction.CanonicalMessage) error {
	for _, field := range v.rules.RequiredFields {
		if err := v.checkFieldPresent(msg, field); err != nil {
			return err
		}
	}
	return nil
}

// validateFieldTypes checks that fields have the correct types
func (v *Validator) validateFieldTypes(msg *abstraction.CanonicalMessage) error {
	for field, expectedType := range v.rules.FieldTypes {
		if err := v.checkFieldType(msg, field, expectedType); err != nil {
			return err
		}
	}
	return nil
}

// validateConstraints checks field-specific constraints
func (v *Validator) validateConstraints(msg *abstraction.CanonicalMessage) error {
	for field, constraint := range v.rules.Constraints {
		if err := v.checkConstraint(msg, field, constraint); err != nil {
			return err
		}
	}
	return nil
}

// runCustomRules executes custom validation rules
func (v *Validator) runCustomRules(msg *abstraction.CanonicalMessage) error {
	for _, rule := range v.rules.CustomRules {
		if err := rule.Function(msg); err != nil {
			return &abstraction.MessageValidationError{
				Field:   rule.Name,
				Message: fmt.Sprintf("custom validation failed: %v", err),
				Code:    "CUSTOM_VALIDATION_FAILED",
			}
		}
	}
	return nil
}

// checkFieldPresent verifies that a required field is present and not empty
func (v *Validator) checkFieldPresent(msg *abstraction.CanonicalMessage, field string) error {
	switch field {
	case "chain_id":
		if msg.ChainID == "" {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "chain_id is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "height":
		if msg.Height == nil {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "height is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "timestamp":
		if msg.Timestamp.IsZero() {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "timestamp is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "type":
		if msg.Type == "" {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "type is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "round":
		if msg.Round == nil {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "round is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "proposer":
		if msg.Proposer == "" {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "proposer is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "validator":
		if msg.Validator == "" {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "validator is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "signature":
		if msg.Signature == "" {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "signature is required",
				Code:    "MISSING_FIELD",
			}
		}
	case "block_hash":
		if msg.BlockHash == "" {
			return &abstraction.MessageValidationError{
				Field:   field,
				Message: "block_hash is required",
				Code:    "MISSING_FIELD",
			}
		}
	}
	return nil
}

// checkFieldType verifies that a field has the correct type
func (v *Validator) checkFieldType(msg *abstraction.CanonicalMessage, field, expectedType string) error {
	switch field {
	case "height":
		if msg.Height != nil && expectedType == "bigint" {
			return nil
		}
	case "round":
		if msg.Round != nil && expectedType == "bigint" {
			return nil
		}
	case "view":
		if msg.View != nil && expectedType == "bigint" {
			return nil
		}
	case "timestamp":
		if !msg.Timestamp.IsZero() && expectedType == "time" {
			return nil
		}
	case "chain_id":
		if msg.ChainID != "" && expectedType == "string" {
			return nil
		}
	case "type":
		if msg.Type != "" && expectedType == "string" {
			return nil
		}
	}

	return &abstraction.MessageValidationError{
		Field:   field,
		Message: fmt.Sprintf("field %s has incorrect type, expected %s", field, expectedType),
		Code:    "INVALID_FIELD_TYPE",
	}
}

// checkConstraint verifies that a field meets specific constraints
func (v *Validator) checkConstraint(msg *abstraction.CanonicalMessage, field string, constraint interface{}) error {
	switch field {
	case "height":
		if msg.Height != nil {
			if minHeight, ok := constraint.(map[string]interface{})["min"]; ok {
				if minVal, ok := minHeight.(float64); ok {
					if msg.Height.Cmp(big.NewInt(int64(minVal))) < 0 {
						return &abstraction.MessageValidationError{
							Field:   field,
							Message: fmt.Sprintf("height must be >= %v", minVal),
							Code:    "CONSTRAINT_VIOLATION",
						}
					}
				}
			}
		}
	case "round":
		if msg.Round != nil {
			if minRound, ok := constraint.(map[string]interface{})["min"]; ok {
				if minVal, ok := minRound.(float64); ok {
					if msg.Round.Cmp(big.NewInt(int64(minVal))) < 0 {
						return &abstraction.MessageValidationError{
							Field:   field,
							Message: fmt.Sprintf("round must be >= %v", minVal),
							Code:    "CONSTRAINT_VIOLATION",
						}
					}
				}
			}
		}
	case "timestamp":
		if !msg.Timestamp.IsZero() {
			if maxAge, ok := constraint.(map[string]interface{})["max_age_seconds"]; ok {
				if maxAgeVal, ok := maxAge.(float64); ok {
					age := time.Since(msg.Timestamp).Seconds()
					if age > maxAgeVal {
						return &abstraction.MessageValidationError{
							Field:   field,
							Message: fmt.Sprintf("message is too old: %v seconds", age),
							Code:    "CONSTRAINT_VIOLATION",
						}
					}
				}
			}
		}
	}
	return nil
}

// getDefaultRules returns default validation rules for each chain type
func getDefaultRules(chainType abstraction.ChainType) ValidationRules {
	switch chainType {
	case abstraction.ChainTypeCometBFT:
		return ValidationRules{
			RequiredFields: []string{"chain_id", "height", "round", "timestamp", "type"},
			FieldTypes: map[string]string{
				"chain_id":  "string",
				"height":    "bigint",
				"round":     "bigint",
				"timestamp": "time",
				"type":      "string",
			},
			Constraints: map[string]interface{}{
				"height": map[string]interface{}{
					"min": float64(0),
				},
				"round": map[string]interface{}{
					"min": float64(0),
				},
				"timestamp": map[string]interface{}{
					"max_age_seconds": float64(3600), // 1 hour
				},
			},
			CustomRules: []CustomValidationRule{
				{
					Name:        "cometbft_message_type",
					Description: "Validate CometBFT-specific message types",
					Function:    validateCometBFTMessageType,
				},
			},
		}
	case abstraction.ChainTypeHyperledger:
		return ValidationRules{
			RequiredFields: []string{"chain_id", "height", "timestamp", "type"},
			FieldTypes: map[string]string{
				"chain_id":  "string",
				"height":    "bigint",
				"timestamp": "time",
				"type":      "string",
			},
			Constraints: map[string]interface{}{
				"height": map[string]interface{}{
					"min": float64(0),
				},
				"timestamp": map[string]interface{}{
					"max_age_seconds": float64(7200), // 2 hours
				},
			},
			CustomRules: []CustomValidationRule{
				{
					Name:        "hyperledger_message_type",
					Description: "Validate Hyperledger-specific message types",
					Function:    validateHyperledgerMessageType,
				},
			},
		}
	case abstraction.ChainTypeKaia:
		return ValidationRules{
			RequiredFields: []string{"chain_id", "height", "round", "timestamp", "type"},
			FieldTypes: map[string]string{
				"chain_id":  "string",
				"height":    "bigint",
				"round":     "bigint",
				"timestamp": "time",
				"type":      "string",
			},
			Constraints: map[string]interface{}{
				"height": map[string]interface{}{
					"min": float64(0),
				},
				"round": map[string]interface{}{
					"min": float64(0),
				},
				"timestamp": map[string]interface{}{
					"max_age_seconds": float64(1800), // 30 minutes
				},
			},
			CustomRules: []CustomValidationRule{
				{
					Name:        "kaia_message_type",
					Description: "Validate Kaia-specific message types",
					Function:    validateKaiaMessageType,
				},
			},
		}
	default:
		return ValidationRules{
			RequiredFields: []string{"chain_id", "height", "timestamp", "type"},
			FieldTypes: map[string]string{
				"chain_id":  "string",
				"height":    "bigint",
				"timestamp": "time",
				"type":      "string",
			},
		}
	}
}

// Chain-specific validation functions
func validateCometBFTMessageType(msg *abstraction.CanonicalMessage) error {
	validTypes := map[abstraction.MsgType]bool{
		abstraction.MsgTypeProposal:  true,
		abstraction.MsgTypePrevote:   true,
		abstraction.MsgTypePrecommit: true,
		abstraction.MsgTypeBlock:     true,
	}

	if !validTypes[msg.Type] {
		return fmt.Errorf("unsupported CometBFT message type: %s", msg.Type)
	}
	return nil
}

func validateHyperledgerMessageType(msg *abstraction.CanonicalMessage) error {
	validTypes := map[abstraction.MsgType]bool{
		abstraction.MsgTypeProposal:   true,
		abstraction.MsgTypePrepare:    true,
		abstraction.MsgTypeCommit:     true,
		abstraction.MsgTypeViewChange: true,
		abstraction.MsgTypeNewView:    true,
	}

	if !validTypes[msg.Type] {
		return fmt.Errorf("unsupported Hyperledger message type: %s", msg.Type)
	}
	return nil
}

func validateKaiaMessageType(msg *abstraction.CanonicalMessage) error {
	validTypes := map[abstraction.MsgType]bool{
		abstraction.MsgTypeProposal: true,
		abstraction.MsgTypeVote:     true,
		abstraction.MsgTypeBlock:    true,
	}

	if !validTypes[msg.Type] {
		return fmt.Errorf("unsupported Kaia message type: %s", msg.Type)
	}
	return nil
}
