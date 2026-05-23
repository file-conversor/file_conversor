// internal/pdf/merge.go

package pdf

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/env/glob"
	"github.com/file-conversor/file_conversor/internal/validation"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

var MergeInFormats = []string{".pdf"}
var MergeOutFormats = []string{".pdf"}

type Merge struct {
	Overwrite bool     // overwrite output file if it exists (default: false)
	Append    bool     // append to output file (if it exists)
	Recurse   bool     // recurse into subdirectories when input is a directory
	Output    string   // output file path (use - for stdout)
	Inputs    []string // input file paths (use - for stdin), can be dirs or .pdf files
}

func (m *Merge) parse() error {
	if m.Recurse { // if recurse flag is set, expand directories in input
		glob, err := glob.New(MergeInFormats...)
		if err != nil {
			return fmt.Errorf("pdf merge - failed to create globber: %v", err)
		}
		m.Inputs, err = glob.GlobFiles(m.Inputs...)
		if err != nil {
			return fmt.Errorf("pdf merge - failed to recurse dirs, or incorrect file extension: %v", err)
		}
	}
	return nil
}

func (m *Merge) validate() error {
	return errors.Join(
		// validation for output file
		validation.CheckOutputStdout(true, m.Output),
		validation.OutputFileExt(m.Output, MergeOutFormats...),
		validation.OutputFileOverwritable(m.Output, m.Overwrite || m.Append),

		// validation for input files
		validation.CheckInputStdin(true, m.Inputs...),
		validation.InputFileExt(m.Inputs, MergeInFormats...),
		validation.InputPathExists(m.Inputs...),
		validation.InputNotEmpty(m.Inputs...),

		// validation for input and output (together)
		validation.InputOutputNotEqual(m.Output, m.Inputs...),
	)
}

// Merge merges multiple PDF files into a single PDF file.
func (m *Merge) Run() error {
	var inIos []io.ReadSeeker
	var outIo *os.File
	var readStdin bool

	// parse user provided arguments (e.g. handle directories in input)
	if err := errors.Join(
		m.parse(),
		m.validate(),
	); err != nil {
		return fmt.Errorf("pdf merge - parse input: %v", err)
	}

	if m.Append && env.FileExists(m.Output) {
		// copy existing output file to tmp file (if it exists) so that we can append it to output
		outFile, err := os.Open(m.Output)
		if err != nil {
			return fmt.Errorf("pdf merge (append mode) - open output file: %v", err)
		}

		tmpFile, cleanup, err := env.CreateTempFile("pdf_merge_append_*.pdf", outFile)
		defer cleanup() // ensure tmp file is cleaned up
		outFile.Close()
		if err != nil {
			return fmt.Errorf("pdf merge (append mode) - copy output to tmp file: %v", err)
		}
		inIos = append(inIos, tmpFile) // add tmp file as input to be merged with the other input files
	}

	for _, input := range m.Inputs {
		switch input {
		case "":
			return fmt.Errorf("input file path cannot be empty")
		case "-":
			if readStdin {
				continue // already reading from stdin, skip additional "-"
			}
			readStdin = true
			tmpFile, cleanup, err := env.CreateTempFile("pdf_merge_stdin_*.pdf", os.Stdin)
			if err != nil {
				return fmt.Errorf("pdf merge - copy stdin to tmp file: %v", err)
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

	switch m.Output {
	case "":
		return fmt.Errorf("output cannot be empty")
	case "-":
		outIo = os.Stdout
	default:
		outFile, err := env.OpenOutputFile(m.Output, false)
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
