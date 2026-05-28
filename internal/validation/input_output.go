// internal/validation/input_output.go

package validation

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/utils"
)

func AllowOrDenyInputStdin(allow bool, paths ...string) error {
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

func AllowOrDenyOutputStdout(allow bool, paths ...string) error {
	if env.IsStdIO(paths...) && !allow {
		return fmt.Errorf("stdout is not allowed")
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

// check if input exists
func InputPathExists(inputs ...string) error {
	for _, input := range inputs {
		switch input {
		case "-":
			continue // skip stdin since it's not an actual file
		default:
			if !env.PathExists(input) {
				return fmt.Errorf("input path '%s' does not exist", input)
			}
		}
	}
	return nil
}

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

// Checks if all input files have an allowed extension
func InputFileExt(inputs []string, allowedExts ...string) error {
	var errGrp error
	for _, input := range inputs {
		errGrp = errors.Join(errGrp, OutputFileExt(input, allowedExts...))
	}
	return errGrp
}

func OutputValidSuffix(suffix string) error {
	switch suffix {
	case "":
		return nil // empty suffix is allowed
	default:
		if utils.ContainsAnySubstrings(suffix,
			string(os.PathSeparator),
			string(os.PathListSeparator),
			".",
		) {
			return fmt.Errorf("invalid output suffix '%s'", suffix)
		}
		return nil
	}
}

// checks if output dir exists
func OutputDirExists(outputDir string) error {
	if !env.DirExists(outputDir) {
		return fmt.Errorf("output dir '%s' does not exist", outputDir)
	}
	return nil
}

// Checks if the output file can be overwritten
// (exists and not a directory, or doesn't exist)
func OutputFileOverwritable(output string, overwrite bool) error {
	switch output {
	case "-":
		return nil // stdout can always be overwritten
	default:
		if env.FileExists(output) && !overwrite {
			return fmt.Errorf("output file already exists: %s", output)
		}
		return nil
	}
}

// Checks if the output file has an allowed extension
func OutputFileExt(output string, allowedExts ...string) error {
	switch output {
	case "-":
		return nil // stdout can always be overwritten
	default:
		if !env.IsFileExt(output, allowedExts...) {
			return fmt.Errorf("invalid extension for '%s', expected %v", output, allowedExts)
		}
		return nil
	}
}
