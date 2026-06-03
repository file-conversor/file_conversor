// internal/pdf/split.go

package pdf

import (
	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/engine"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type Split struct {
	core.MapCommand // MapCommand: batch input file => output directory
}

func NewSplit(outputDirFlag flags.OutputDirFlag, inputsArg flags.InputFilesArg) *Split {
	return &Split{
		MapCommand: core.MapCommand{
			OutputDirFlag: outputDirFlag,
			InputFilesArg: inputsArg,
		},
	}
}

func (d *Split) In() []string {
	return []string{".pdf"}
}

func (d *Split) Out() []string {
	return []string{".pdf"}
}

func (d *Split) Parse() error {
	return d.MapCommand.Parse(d)
}

func (d *Split) Validate() error {
	return d.MapCommand.Validate(d)
}

func (d *Split) GetRunnable() <-chan *core.Runnable {
	return d.MapCommand.GetRunnableDir(func(inFile string, outDir string, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		pdfcpuEngine := engine.NewPdfCpuEngine(
			engine.PdfCpuPasswordEmpty,
			engine.PdfCpuEncryptionNone,
			engine.PdfCpuPermissionsNone,
			false,
		)
		logger.Infof("Splitting '%s' => '%s' ...\n", inFile, outDir)
		return pdfcpuEngine.Split(inFile, outDir)
	})
}
