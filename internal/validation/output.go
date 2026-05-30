// internal/validation/input_output.go

package validation

import (
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/utils"
)

func AllowOrDenyOutputStdout(allow bool, paths ...string) error {
	if env.IsStdIO(paths...) && !allow {
		return fmt.Errorf("stdout is not allowed")
	}
	return nil
}

func OutputValidSuffix(suffix string) error {
	switch suffix {
	case "":
		return nil // empty suffix is allowed
	default:
		if utils.ContainsAnySubstrings(suffix,
			string(os.PathSeparator),
			string(os.PathListSeparator),
		) {
			return fmt.Errorf("invalid output suffix '%s'", suffix)
		}
		return nil
	}
}

// ensure output dir exists
func OutputDirEnsure(outputDir string) error {
	switch outputDir {
	case "-":
		// skip stdout since it's not an actual directory
	default:
		if err := env.MkdirAll(outputDir); err != nil {
			return fmt.Errorf("create output directory '%s': %w", outputDir, err)
		}
	}
	return nil
}

// checks if output dir exists
func OutputDirExists(outputDir string) error {
	switch outputDir {
	case "-":
		return nil // skip stdout since it's not an actual directory
	default:
		if !env.DirExists(outputDir) {
			return fmt.Errorf("output dir '%s' does not exist", outputDir)
		}
		return nil
	}
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
