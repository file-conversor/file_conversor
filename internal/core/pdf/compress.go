// internal/pdf/compress.go

package pdf

import (
	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/engine"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type Compress struct {
	core.MapCommand // MapCommand: batch input file => output directory
}

func NewCompress(outputDirFlag flags.OutputDirFlag, inputsArg flags.InputFilesArg) *Compress {
	return &Compress{
		MapCommand: core.MapCommand{
			OutputDirFlag: outputDirFlag,
			InputFilesArg: inputsArg,
		},
	}
}

func (d *Compress) In() []string {
	return []string{".pdf"}
}

func (d *Compress) Out() []string {
	return []string{".pdf"}
}

func (d *Compress) Parse() error {
	return d.MapCommand.Parse(d)
}

func (d *Compress) Validate() error {
	return d.MapCommand.Validate(d)
}

func (d *Compress) GetRunnable() <-chan *core.Runnable {
	return d.MapCommand.GetRunnable(func(inFile string, outFile string, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		pdfcpuEngine := engine.NewPdfCpuEngine(
			engine.PdfCpuPasswordEmpty,
			engine.PdfCpuEncryptionNone,
			engine.PdfCpuPermissionsNone,
			false,
		)
		logger.Infof("Compressing '%s' => '%s' ...\n", inFile, outFile)
		return pdfcpuEngine.Compress(inFile, outFile)
	})
}
