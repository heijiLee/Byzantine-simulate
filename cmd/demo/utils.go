package main

import (
	"encoding/json"
	"fmt"

	"codec/message/abstraction"
)

func printRawMessage(raw abstraction.RawConsensusMessage) {
	fmt.Println("   Raw message")
	prettyPrintJSON(raw)
}

func printCanonicalMessage(canonical *abstraction.CanonicalMessage) {
	fmt.Println("   Canonical message")
	prettyPrintJSON(canonical)
}

func prettyPrintJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "      ", "  ")
	if err != nil {
		fmt.Printf("      error marshaling JSON: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(jsonData))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
