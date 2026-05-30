// internal/interfaces/interfaces.go

package interfaces

type ProgressIncrement func(int64)

type FormatInterface interface {
	In() []string
	Out() []string
}
