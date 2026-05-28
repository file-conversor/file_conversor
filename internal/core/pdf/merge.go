// internal/pdf/merge.go

package pdf

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/validation"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type mergeCmd struct {
}

type Merge struct {
	Append              bool // append to output file (if it exists)
	core.OutputFileFlag      // output file arguments (path, overwrite flag)
	core.InputsArg           // input file arguments (paths, recurse flag, batch file)
}

func (this Merge) In() []string {
	return []string{".pdf"}
}

func (this Merge) Out() []string {
	return []string{".pdf"}
}

func NewMerge(append bool, outputFileFlag core.OutputFileFlag, inputsArg core.InputsArg) (*Merge, error) {
	merge := &Merge{
		Append:         append,
		OutputFileFlag: outputFileFlag,
		InputsArg:      inputsArg,
	}
	if err := merge.Parse(); err != nil {
		return nil, err
	}
	if err := merge.Validate(); err != nil {
		return nil, err
	}
	return merge, nil
}

func (this *Merge) Parse() error {
	// if append is true, we need to allow overwriting the output file
	this.OutputFileFlag.Overwrite = this.OutputFileFlag.Overwrite || this.Append

	if err := errors.Join(
		this.OutputFileFlag.Parse(this),
		this.InputsArg.Parse(this),
	); err != nil {
		return fmt.Errorf("parse input files: %w", err)
	}
	return nil
}

func (this *Merge) Validate() error {
	return errors.Join(
		// validation for output file
		this.OutputFileFlag.Validate(this),

		// validation for input files
		this.InputsArg.Validate(this),

		// validation for input and output (together)
		validation.InputOutputNotEqual(this.OutputFile, this.InputFiles...),
	)
}

// Merge merges multiple PDF files into a single PDF file.
func (this *Merge) GetRunnable() <-chan core.Runnable {
	runnable := core.NewRunnable()
	outChan := make(chan core.Runnable) // channel to send runnable to be executed

	go func() {
		defer func() {
			outChan <- *runnable // send runnable
			close(outChan)       // close channel when done
		}()

		var inIos []io.ReadSeeker
		var outFile *os.File

		// append existing output file (if exists)
		if this.Append && env.FileExists(this.OutputFile) {
			// copy existing output file to tmp file (if it exists) so that we can append it to output
			tmpFile, cleanup, err := env.CopyToTmpFile(this.OutputFile)
			if err != nil {
				runnable.SetError(fmt.Errorf("copy output to tmp file: %w", err))
				return
			}
			// ensure tmp file is cleaned up
			runnable.AppendCleanup(cleanup)
			inIos = append(inIos, tmpFile) // add tmp file as input to be merged with the other input files
		}

		// open input files
		cleanupFuncs, err := env.OpenInputFiles(func(file *os.File) error {
			inIos = append(inIos, file) // add input files to list of input io.ReadSeekers
			return nil
		}, this.InputFiles...)
		if err != nil {
			runnable.SetError(fmt.Errorf("open input files: %w", err))
			return
		}
		runnable.AppendCleanup(cleanupFuncs...)

		// open output file
		cleanupFuncs, err = env.OpenOutputFile(func(file *os.File) error {
			outFile = file // set output file handle for writing output
			return nil
		}, this.OutputFile)
		if err != nil {
			runnable.SetError(fmt.Errorf("open output file: %w", err))
			return
		}
		runnable.AppendCleanup(cleanupFuncs...)

		// use MergeRaw to merge input => output without intermediate files on disk
		// note: api.Merge() is not used because it only accepts file paths, and we want to support stdin/stdout
		runnable.SetRun(func(updateProgress interfaces.ProgressIncrement) error {
			defer updateProgress(100) // ensure progress is updated to 100% when done
			return api.MergeRaw(inIos, outFile, false, nil)
		})
	}()
	return outChan
}
