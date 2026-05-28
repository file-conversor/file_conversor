// internal/core/input_files.go

package core

import (
	"errors"
	"fmt"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/env/glob"
	"github.com/file-conversor/file_conversor/internal/validation"
)

type InputsArg struct {
	InputFiles  []string // input file paths (use - for stdin), can be dirs or .pdf files
	Recurse     bool     // recurse into subdirectories when input is a directory
	BatchFile   string   // batch file containing list of input files (one per line)
	AcceptStdin bool     // whether to accept stdin as input (set by command implementations)
}

func (i *InputsArg) Parse(formats FormatInterface) error {
	// read input file paths from batch file (one per line)
	if i.BatchFile != "" {
		err := env.ReadLines(i.BatchFile, func(line string) {
			i.InputFiles = append(i.InputFiles, line)
		})
		if err != nil {
			return fmt.Errorf("parse inputs - read batch file: %w", err)
		}
	}

	// if recurse flag is set, expand directories in input
	if i.Recurse {
		glob, err := glob.New(formats.In()...)
		if err != nil {
			return fmt.Errorf("parse inputs - create globber: %w", err)
		}
		inputFiles, err := glob.GlobFiles(i.InputFiles...)
		if err != nil {
			return fmt.Errorf("parse inputs - recurse dirs, or incorrect file extension: %w", err)
		}
		i.InputFiles = append(i.InputFiles, inputFiles...)
	}

	// default to stdin if no inputs provided
	if len(i.InputFiles) == 0 {
		i.InputFiles = append(i.InputFiles, "-")
	}
	return nil
}

func (i *InputsArg) Validate(formats FormatInterface) error {
	return errors.Join(
		validation.AllowOrDenyInputStdin(i.AcceptStdin, i.InputFiles...),
		validation.InputFileExt(i.InputFiles, formats.In()...),
		validation.InputPathExists(i.InputFiles...),
		validation.IsNotEmpty("input", i.InputFiles...),
	)
}
