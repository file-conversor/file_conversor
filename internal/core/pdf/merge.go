// internal/pdf/merge.go

package pdf

import (
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type Merge struct {
	core.ReduceCommand // ReduceCommand: input files => output file
}

func NewMerge(append bool, outputFileFlag flags.OutputFileFlag, inputsArg flags.InputFilesArg) *Merge {
	return &Merge{
		ReduceCommand: core.ReduceCommand{
			Append:         append,
			OutputFileFlag: outputFileFlag,
			InputFilesArg:  inputsArg,
		},
	}
}

func (m Merge) In() []string {
	return []string{".pdf"}
}

func (m Merge) Out() []string {
	return []string{".pdf"}
}

func (m *Merge) Parse() error {
	return m.ReduceCommand.Parse(m) // pass Merge as FormatInterface to ReduceCommand
}

func (m *Merge) Validate() error {
	return m.ReduceCommand.Validate(m) // pass Merge as FormatInterface to ReduceCommand
}

// Merge merges multiple PDF files into a single PDF file.
func (m *Merge) GetRunnable() <-chan *core.Runnable {
	return m.ReduceCommand.GetRunnable(func(inFiles []*os.File, outFile *os.File, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		// merge PDF files using pdfcpu api
		logger.Infof(
			"Merging files to '%s'\n",
			outFile.Name(),
		)

		inIos := env.ToIoReadSeekers(inFiles)
		if err := api.MergeRaw(inIos, outFile, false, nil); err != nil {
			return fmt.Errorf("pdf merge: %w", err)
		}
		return nil
	})
}
