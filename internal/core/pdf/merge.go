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

type Merge struct {
	Append              bool // append to output file (if it exists)
	core.OutputFileFlag      // output file arguments (path, overwrite flag)
	core.InputsArg           // input file arguments (paths, recurse flag, batch file)
}

func (m Merge) In() []string {
	return []string{".pdf"}
}

func (m Merge) Out() []string {
	return []string{".pdf"}
}

func NewMerge(append bool, outputFileFlag core.OutputFileFlag, inputsArg core.InputsArg) (*Merge, error) {
	command := &Merge{
		Append:         append,
		OutputFileFlag: outputFileFlag,
		InputsArg:      inputsArg,
	}
	if err := command.Parse(); err != nil {
		return nil, err
	}
	if err := command.Validate(); err != nil {
		return nil, err
	}
	return command, nil
}

func (m *Merge) Parse() error {
	// if append is true, we need to allow overwriting the output file
	m.OutputFileFlag.Overwrite = m.OutputFileFlag.Overwrite || m.Append

	if err := errors.Join(
		m.OutputFileFlag.Parse(m),
		m.InputsArg.Parse(m),
	); err != nil {
		return fmt.Errorf("parse input files: %w", err)
	}
	return nil
}

func (m *Merge) Validate() error {
	return errors.Join(
		// validation for output file
		m.OutputFileFlag.Validate(m),

		// validation for input files
		m.InputsArg.Validate(m),

		// validation for input and output (together)
		validation.InputOutputNotEqual(m.OutputFile, m.InputFiles...),
	)
}

// Merge merges multiple PDF files into a single PDF file.
func (m *Merge) GetRunnable() <-chan core.Runnable {
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
		if m.Append && env.FileExists(m.OutputFile) {
			// copy existing output file to tmp file (if it exists) so that we can append it to output
			tmpFile, cleanup, err := env.CopyToTmpFile(m.OutputFile)
			if err != nil {
				runnable.SetError(fmt.Errorf("copy output to tmp file: %w", err))
				return
			}
			// ensure tmp file is cleaned up
			runnable.AppendCleanup(cleanup)
			inIos = append(inIos, tmpFile) // add tmp file as input to be merged with the other input files
		}

		// open input files
		err := env.OpenInputFiles(func(file *os.File, cleanup func() error) error {
			inIos = append(inIos, file) // add input files to list of input io.ReadSeekers
			runnable.AppendCleanup(cleanup)
			return nil
		}, m.InputFiles...)
		if err != nil {
			runnable.SetError(fmt.Errorf("open input files: %w", err))
			return
		}

		// open output file
		err = env.OpenOutputFile(func(file *os.File, cleanup func() error) error {
			outFile = file // set output file handle for writing output
			runnable.AppendCleanup(cleanup)
			return nil
		}, m.OutputFile)
		if err != nil {
			runnable.SetError(fmt.Errorf("open output file: %w", err))
			return
		}

		// use MergeRaw to merge input => output without intermediate files on disk
		// note: api.Merge() is not used because it only accepts file paths, and we want to support stdin/stdout
		runnable.SetRun(func(updateProgress interfaces.ProgressIncrement) error {
			if err := api.MergeRaw(inIos, outFile, false, nil); err != nil {
				runnable.AppendCleanup(func() error { return os.Remove(outFile.Name()) })
				return fmt.Errorf("pdf merge: %w", err)
			}
			return nil
		})
	}()
	return outChan
}
