// Package common provides shared utilities for command execution.
package common

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ReadYAMLFromInput reads YAML content from file, stdin, or pipe
// filename: file path, or "-" for stdin
// Returns the raw bytes and any error
func ReadYAMLFromInput(filename string) ([]byte, error) {
	// Case 1: Explicit file or stdin
	if filename != "" {
		if filename == "-" {
			return os.ReadFile("/dev/stdin")
		}
		return os.ReadFile(filename)
	}

	// Case 2: Check for pipe input
	if HasPipeInput() {
		return os.ReadFile("/dev/stdin")
	}

	return nil, fmt.Errorf("no input source provided")
}

// HasPipeInput checks if there's data piped to stdin
func HasPipeInput() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// ParseYAML parses YAML bytes into the target struct
func ParseYAML(data []byte, target interface{}) error {
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("error parsing YAML: %w", err)
	}
	return nil
}

// ReadAndParseYAML combines reading and parsing YAML from input source
// Returns true if successful, false if no input was provided
func ReadAndParseYAML(filename string, target interface{}) (bool, error) {
	data, err := ReadYAMLFromInput(filename)
	if err != nil {
		if err.Error() == "no input source provided" {
			return false, nil
		}
		return false, err
	}

	if err := ParseYAML(data, target); err != nil {
		return false, err
	}

	return true, nil
}
