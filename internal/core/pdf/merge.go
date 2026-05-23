// internal/pdf/merge.go

package pdf

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/validation"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type MergeFormats struct{}

func (this MergeFormats) In() []string {
	return []string{".pdf"}
}

func (this MergeFormats) Out() []string {
	return []string{".pdf"}
}

type Merge struct {
	Append          bool // append to output file (if it exists)
	core.OutputFlag      // output file arguments (path, overwrite flag)
	core.InputsArg       // input file arguments (paths, recurse flag, batch file)
}

func (this *Merge) Parse() error {
	// if append is true, we need to allow overwriting the output file
	this.OutputFlag.Overwrite = this.OutputFlag.Overwrite || this.Append

	if err := errors.Join(
		this.OutputFlag.Parse(MergeFormats{}),
		this.InputsArg.Parse(MergeFormats{}),
	); err != nil {
		return fmt.Errorf("parse input files: %v", err)
	}
	return nil
}

func (this *Merge) Validate() error {
	return errors.Join(
		// validation for output file
		this.OutputFlag.Validate(true, MergeFormats{}),

		// validation for input files
		this.InputsArg.Validate(true, MergeFormats{}),

		// validation for input and output (together)
		validation.InputOutputNotEqual(this.OutputFile, this.InputFiles...),
	)
}

// Merge merges multiple PDF files into a single PDF file.
func (this *Merge) Run() error {
	var inIos []io.ReadSeeker
	var outIo *os.File
	var readStdin bool

	// parse user provided arguments (e.g. handle directories in input)
	if err := errors.Join(
		this.Parse(),
		this.Validate(),
	); err != nil {
		return fmt.Errorf("parse input: %v", err)
	}

	if this.Append && env.FileExists(this.OutputFile) {
		// copy existing output file to tmp file (if it exists) so that we can append it to output
		tmpFile, cleanup, err := env.CopyToTmpFile(this.OutputFile)
		if err != nil {
			return fmt.Errorf("copy output to tmp file: %v", err)
		}
		defer cleanup()                // ensure tmp file is cleaned up
		inIos = append(inIos, tmpFile) // add tmp file as input to be merged with the other input files
	}

	for _, input := range this.InputFiles {
		switch input {
		case "":
			return fmt.Errorf("input file path cannot be empty")
		case "-":
			if readStdin {
				continue // already reading from stdin, skip additional "-"
			}
			readStdin = true
			tmpFile, cleanup, err := env.CopyToTmpFileRaw("", os.Stdin)
			if err != nil {
				return fmt.Errorf("copy stdin to tmp file: %v", err)
			}
			defer cleanup() // ensure tmp file is cleaned up
			inIos = append(inIos, tmpFile)
		default:
			inFile, err := os.Open(input)
			if err != nil {
				return err
			}
			defer inFile.Close()
			inIos = append(inIos, inFile)
		}
	}

	switch this.OutputFile {
	case "":
		return fmt.Errorf("output cannot be empty")
	case "-":
		outIo = os.Stdout
	default:
		outFile, err := env.OpenOutputFile(this.OutputFile, false)
		if err != nil {
			return err
		}
		defer outFile.Close()
		outIo = outFile
	}

	// use MergeRaw to merge input => output without intermediate files on disk
	// note: api.Merge() is not used because it only accepts file paths, and we want to support stdin/stdout
	return api.MergeRaw(inIos, outIo, false, nil)
}
