// internal/validation/input.go

package validation

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
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

// Checks if all input files have an allowed extension
func InputFileExt(inputs []string, allowedExts ...string) error {
	var errGrp error
	for _, input := range inputs {
		errGrp = errors.Join(errGrp, OutputFileExt(input, allowedExts...))
	}
	return errGrp
}
