// internal/core/output_file.go

package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/validation"
)

type OutputFlag struct {
	Overwrite  bool   // overwrite output file if it exists (default: false)
	OutputFile string // output file path (use - for stdout)
}

func (this *OutputFlag) Parse(formats Format) error {
	return nil
}

func (this *OutputFlag) Validate(acceptStdout bool, formats Format) error {
	return errors.Join(
		validation.CheckOutputStdout(acceptStdout, this.OutputFile),
		validation.OutputFileExt(this.OutputFile, formats.Out()...),
		validation.OutputFileOverwritable(this.OutputFile, this.Overwrite),
	)
}

func (this *OutputFlag) OpenOutputFile() (outFile *os.File, cleanupFuncs []func() error, errGrp error) {
	var err error
	switch this.OutputFile {
	case "":
		errGrp = errors.Join(errGrp, fmt.Errorf("output cannot be empty"))
	case "-":
		outFile = os.Stdout
	default:
		outFile, err = env.OpenOutputFile(this.OutputFile, false)
		if err != nil {
			errGrp = errors.Join(errGrp, fmt.Errorf("open output file '%s': %w", this.OutputFile, err))
		} else {
			cleanupFuncs = append(cleanupFuncs, func() error { outFile.Close(); return nil }) // ensure file is closed
		}
	}
	if errGrp != nil {
		for _, cleanup := range cleanupFuncs {
			cleanup() // clean up any files that were opened before returning error
		}
		cleanupFuncs = nil
		outFile = nil
	}
	return
}
