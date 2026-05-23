// internal/env/exec.go

package env

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommand runs a command and returns an error if it fails.
func RunCommand(cmd ...string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}

// RunCommands runs multiple commands sequentially and returns an error if any command fails.
func RunCommands(cmds ...[]string) error {
	for _, cmd := range cmds {
		if len(cmd) == 0 {
			return fmt.Errorf("empty command")
		}
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("cmd '%s': %w", strings.Join(cmd, " "), err)
		}
	}
	return nil
}
