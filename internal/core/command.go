// internal/core/command.go

package core

type CommandInterface interface {
	// In returns the accepted input file formats (extensions) for this command.
	In() []string
	// Out returns the output file formats (extensions) produced by this command.
	Out() []string
	// parse and validate the command's flags and arguments
	Parse() error
	Validate() error
	// GetRunnable returns a channel of Runnables that can be executed concurrently to perform the command's work.
	GetRunnable() <-chan *Runnable
}
