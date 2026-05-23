// internal/pdf/merge.go

package pdf

import (
	"fmt"
	"io"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Merge merges multiple PDF files into a single PDF file.
func Merge(appendPdf bool, output string, inputs ...string) error {
	var inIos []io.ReadSeeker
	var outIo *os.File
	var readStdin bool

	if appendPdf && env.FileExists(output) {
		// copy existing output file to tmp file (if it exists) so that we can append it to output
		outFile, err := os.Open(output)
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

	for _, input := range inputs {
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

	switch output {
	case "":
		return fmt.Errorf("output cannot be empty")
	case "-":
		outIo = os.Stdout
	default:
		outFile, err := env.OpenOutputFile(output, false)
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
