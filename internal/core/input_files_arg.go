// internal/core/input_files.go

package core

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/env/glob"
	"github.com/file-conversor/file_conversor/internal/validation"
)

type InputsArg struct {
	InputFiles []string // input file paths (use - for stdin), can be dirs or .pdf files
	Recurse    bool     // recurse into subdirectories when input is a directory
	BatchFile  string   // batch file containing list of input files (one per line)
}

func (this *InputsArg) Parse(formats Format) error {
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

func (this *InputsArg) Validate(acceptStdin bool, formats Format) error {
	return errors.Join(
		validation.CheckInputStdin(acceptStdin, this.InputFiles...),
		validation.InputFileExt(this.InputFiles, formats.In()...),
		validation.InputPathExists(this.InputFiles...),
		validation.InputNotEmpty(this.InputFiles...),
	)
}

func (this *InputsArg) OpenInputFiles() (inIos []io.ReadSeeker, cleanupFuncs []func() error, errGrp error) {
	var readStdin bool

	for _, input := range this.InputFiles {
		switch input {
		case "":
			errGrp = errors.Join(errGrp, fmt.Errorf("input file path cannot be empty"))
		case "-":
			if readStdin {
				continue // already reading from stdin, skip additional "-"
			}
			readStdin = true
			tmpFile, cleanup, err := env.CopyToTmpFileRaw("", os.Stdin)
			if err != nil {
				errGrp = errors.Join(errGrp, fmt.Errorf("copy stdin to tmp file: %w", err))
			} else {
				cleanupFuncs = append(cleanupFuncs, func() error { cleanup(); return nil }) // ensure tmp file is cleaned up
				inIos = append(inIos, tmpFile)
			}
		default:
			inFile, err := os.Open(input)
			if err != nil {
				errGrp = errors.Join(errGrp, fmt.Errorf("open input file '%s': %w", input, err))
			} else {
				cleanupFuncs = append(cleanupFuncs, inFile.Close) // ensure file is closed
				inIos = append(inIos, inFile)
			}
		}
		if errGrp != nil {
			for _, cleanup := range cleanupFuncs {
				cleanup() // clean up any files that were opened before returning error
			}
			cleanupFuncs = nil
			inIos = nil
			return
		}
	}
	return
}
