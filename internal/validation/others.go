// internal/validation/others.go

package validation

import (
	"fmt"

	"github.com/file-conversor/file_conversor/internal/env"
)

// Checks if the output file is not the same as any of the input files
func InputOutputNotEqual(output string, inputs ...string) error {
	if output == "-" {
		return nil // skip stdout since it's not an actual file
	}
	for _, input := range inputs {
		switch {
		case input == "-":
			continue // skip stdin since its not an actual file
		default:
			isequal, err := env.IsEqualPath(input, output)
			if err != nil || isequal {
				return fmt.Errorf("input == output: %w", err)
			}
		}
	}
	return nil
}

// check if var is not empty
func IsNotEmpty(msg string, strs ...string) error {
	if len(strs) == 0 {
		return fmt.Errorf("%s cannot be empty", msg)
	}
	for _, s := range strs {
		if s == "" {
			return fmt.Errorf("%s cannot be empty", msg)
		}
	}
	return nil
}
