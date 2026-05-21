// internal/env/tty.go

package env

import (
	"os"

	"github.com/mattn/go-isatty"
)

func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
