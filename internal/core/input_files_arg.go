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

func (this *InputsArg) Parse(formats FormatInterface) error {
	// read input file paths from batch file (one per line)
	if this.BatchFile != "" {
		err := env.ReadLines(this.BatchFile, func(line string) {
			this.InputFiles = append(this.InputFiles, line)
		})
		if err != nil {
			return fmt.Errorf("parse inputs - read batch file: %w", err)
		}
	}

	// if recurse flag is set, expand directories in input
	if this.Recurse {
		glob, err := glob.New(formats.In()...)
		if err != nil {
			return fmt.Errorf("parse inputs - create globber: %w", err)
		}
		inputFiles, err := glob.GlobFiles(this.InputFiles...)
		if err != nil {
			return fmt.Errorf("parse inputs - recurse dirs, or incorrect file extension: %w", err)
		}
		this.InputFiles = append(this.InputFiles, inputFiles...)
	}

	// default to stdin if no inputs provided
	if len(this.InputFiles) == 0 {
		this.InputFiles = append(this.InputFiles, "-")
	}
	return nil
}

func (this *InputsArg) Validate(formats FormatInterface) error {
	return errors.Join(
		validation.CheckInputStdin(this.AcceptStdin, this.InputFiles...),
		validation.InputFileExt(this.InputFiles, formats.In()...),
		validation.InputPathExists(this.InputFiles...),
		validation.InputNotEmpty(this.InputFiles...),
	)
}
