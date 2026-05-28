// internal/validation/input_output.go

package validation

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
)

func CheckInputStdin(allow bool, paths ...string) error {
	if env.IsStdIO(paths...) {
		if !allow {
			return fmt.Errorf("stdin is not allowed")
		}
		if env.IsTTY(os.Stdin) {
			return fmt.Errorf("stdin is empty - pipe data to app, or specify input files as arguments")
		}
	}
	return nil
}

func CheckOutputStdout(allow bool, paths ...string) error {
	if env.IsStdIO(paths...) && !allow {
		return fmt.Errorf("stdout is not allowed")
	}
	return nil
}

// check input not empty
func InputNotEmpty(paths ...string) error {
	if len(paths) == 0 {
		return fmt.Errorf("input cannot be empty")
	}
	return nil
}

// check if input exists
func InputPathExists(inputs ...string) error {
	for _, input := range inputs {
		if input == "-" {
			continue // skip stdin since it's not an actual file
		}
		if !env.PathExists(input) {
			return fmt.Errorf("input path '%s' does not exist", input)
		}
	}
	return nil
}

// Checks if the output file is not the same as any of the input files
func InputOutputNotEqual(output string, inputs ...string) error {
	for _, input := range inputs {
		if input == "-" || output == "-" {
			continue // skip stdin/stdout since they are not actual files
		}
		isequal, err := env.IsEqualPath(input, output)
		if err != nil || isequal {
			return fmt.Errorf("input == output: %w", err)
		}
	}
	return nil
}

// Checks if all input files have an allowed extension
func InputFileExt(inputs []string, allowedExts ...string) error {
	var errGrp error
	for _, input := range inputs {
		errGrp = errors.Join(errGrp, OutputFileExt(input, allowedExts...))
	}
	return errGrp
}

// Checks if the output file can be overwritten
// (exists and not a directory, or doesn't exist)
func OutputFileOverwritable(output string, overwrite bool) error {
	if output == "-" {
		return nil // stdout can always be overwritten
	}
	if env.FileExists(output) && !overwrite {
		return fmt.Errorf("output file already exists: %s", output)
	}
	return nil
}

// Checks if the output file has an allowed extension
func OutputFileExt(output string, allowedExts ...string) error {
	if output == "-" {
		return nil // stdout can have any extension since it's not an actual file
	}
	if !env.IsFileExt(output, allowedExts...) {
		return fmt.Errorf("invalid extension for '%s', expected %v", output, allowedExts)
	}
	return nil
}
