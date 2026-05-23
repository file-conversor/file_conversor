// internal/env/tty.go

package env

import (
	"os"

	"github.com/mattn/go-isatty"
)

func IsTTY(file *os.File) bool {
	return isatty.IsTerminal(file.Fd())
}
