// internal/core/formats.go

package core

type Format interface {
	In() []string
	Out() []string
}
