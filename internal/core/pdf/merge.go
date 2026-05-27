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
		return fmt.Errorf("parse input files: %w", err)
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
	var outFile *os.File
	var cleanupFuncs []func() error

	// defer cleanup
	defer func() {
		for _, cleanup := range cleanupFuncs {
			cleanup() // clean up any files that were opened
		}
	}()

	// parse user provided arguments (e.g. handle directories in input)
	if err := this.Parse(); err != nil {
		return fmt.Errorf("parse input: %w", err)
	}

	// validate arguments (e.g. check files exist, etc)
	if err := this.Validate(); err != nil {
		return fmt.Errorf("validate input: %w", err)
	}

	// append existing output file (if exists)
	if this.Append && env.FileExists(this.OutputFile) {
		// copy existing output file to tmp file (if it exists) so that we can append it to output
		tmpFile, cleanup, err := env.CopyToTmpFile(this.OutputFile)
		if err != nil {
			return fmt.Errorf("copy output to tmp file: %w", err)
		}
		// ensure tmp file is cleaned up
		cleanupFuncs = append(cleanupFuncs, func() error { cleanup(); return nil })
		inIos = append(inIos, tmpFile) // add tmp file as input to be merged with the other input files
	}

	// open input files
	in, cleanup, err := this.OpenInputFiles()
	if err != nil {
		return fmt.Errorf("open input files: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, cleanup...)
	inIos = append(inIos, in...) // add input files to list of input io.ReadSeekers

	// open output file
	out, cleanup, err := this.OpenOutputFile()
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, cleanup...)
	outFile = out

	// use MergeRaw to merge input => output without intermediate files on disk
	// note: api.Merge() is not used because it only accepts file paths, and we want to support stdin/stdout
	return api.MergeRaw(inIos, outFile, false, nil)
}
