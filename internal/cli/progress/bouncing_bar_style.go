// internal/cli/progress/filler.go

package progress

import (
	"fmt"
	"io"
	"strings"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// bouncingBarStyle implements mpb.BarFillerBuilder
type bouncingBarStyle struct {
	lbound     string
	filler     string
	padding    string
	rbound     string
	blockWidth int // width of the bouncing block
}

func BouncingBarStyle() *bouncingBarStyle {
	return &bouncingBarStyle{
		lbound:     "[",
		filler:     "█",
		padding:    "░",
		rbound:     "]",
		blockWidth: 5,
	}
}

func (b *bouncingBarStyle) Lbound(char string) *bouncingBarStyle {
	b.lbound = char
	return b
}

func (b *bouncingBarStyle) Filler(char string) *bouncingBarStyle {
	b.filler = char
	return b
}

func (b *bouncingBarStyle) Padding(char string) *bouncingBarStyle {
	b.padding = char
	return b
}

func (b *bouncingBarStyle) Rbound(char string) *bouncingBarStyle {
	b.rbound = char
	return b
}

func (b *bouncingBarStyle) BlockWidth(width int) *bouncingBarStyle {
	b.blockWidth = width
	return b
}

func (b bouncingBarStyle) Build() mpb.BarFiller {
	// Returns the stateful filler for a single bar instance
	barFiller := &bouncingFiller{pos: 0, dir: 1, style: b}
	return barFiller
}

// bouncingFiller implements mpb.BarFiller and maintains state for the bouncing animation
//
// pos is the current position of the block,
//
// dir is the current direction (1 for right, -1 for left)
type bouncingFiller struct {
	pos   int
	dir   int
	style bouncingBarStyle
}

func (f *bouncingFiller) Fill(w io.Writer, stat decor.Statistics) error {
	var (
		LBOUND_CHAR  = f.style.lbound
		FILLER_CHAR  = f.style.filler
		PADDING_CHAR = f.style.padding
		RBOUND_CHAR  = f.style.rbound
		BLOCK_WIDTH  = f.style.blockWidth
	)
	// requested width is the total width available for the bar (including bounds and decorators)
	reqWidth := stat.RequestedWidth
	if reqWidth == 0 {
		reqWidth = stat.AvailableWidth
	}

	// exact number of characters available in the terminal
	// after all prepend/append decorators are drawn!
	if reqWidth < 2 {
		return nil // not enough space to draw even the bounds, skip drawing
	}

	innerWidth := reqWidth - 2 // -2 for the '[' and ']' brackets

	// Handle terminal states gracefully
	if stat.Completed {
		fmt.Fprintf(w,
			"%s%s%s",
			LBOUND_CHAR, strings.Repeat(FILLER_CHAR, innerWidth), RBOUND_CHAR,
		)
		return nil
	}
	if stat.Aborted {
		fmt.Fprintf(w,
			"%s%s%s",
			LBOUND_CHAR, strings.Repeat(PADDING_CHAR, innerWidth), RBOUND_CHAR,
		)
		return nil
	}

	// Ensure block width does not exceed available space
	if BLOCK_WIDTH > innerWidth {
		BLOCK_WIDTH = innerWidth
	}

	// Calculate spacing
	leftSpace := strings.Repeat(PADDING_CHAR, f.pos)
	block := strings.Repeat(FILLER_CHAR, BLOCK_WIDTH)
	rightSpace := strings.Repeat(PADDING_CHAR, max(0, innerWidth-BLOCK_WIDTH-f.pos))

	// Draw the frame
	fmt.Fprintf(w,
		"%s%s%s%s%s",
		LBOUND_CHAR, leftSpace, block, rightSpace, RBOUND_CHAR,
	)

	// Update block position based on current direction
	f.pos += f.dir
	if f.pos <= 0 {
		f.pos = 0
		f.dir = 1
	} else if f.pos+BLOCK_WIDTH >= innerWidth {
		f.pos = max(0, innerWidth-BLOCK_WIDTH)
		f.dir = -1
	}
	return nil
}
