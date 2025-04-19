package linewrap

import (
	"bytes"
	// "fmt"
	"io"
	"strings"
)

// Descriptive var for when not using an io.Writer
var NilWriter io.Writer = nil

type LineWrapper struct {
	maxWidth  int
	currWidth int
	tabWidth  int
	writer    io.Writer
}

func NewLineWrapper(maxWidth, tabWidth int, lwWriter io.Writer) *LineWrapper {
	return &LineWrapper{
		maxWidth:  maxWidth,
		tabWidth:  tabWidth,
		writer:    lwWriter,
		currWidth: 0,
	}
}

// This is primarily for when the terminal is resized and Charm sends the
// WindowSizeMsg; in that case we need to update our wrapper's maxWidth
func (lw *LineWrapper) SetMaxWidth(width int) {
	lw.maxWidth = width
}

func (lw *LineWrapper) Reset() {
	lw.currWidth = 0
}

func (lw *LineWrapper) SetCurrWidth(width int) {
	lw.currWidth = width
}

// Allow wrapping for input that comes in chunks, versus building the line and
// then splitting it and printing the whole line at once.
func (lw *LineWrapper) Write(data []byte) int {
	buffer := wrap_impl(data, lw)

	lw.writer.Write(buffer.Bytes())

	return len(buffer.Bytes())
}

// When not using an io.Writer but will allow the caller to handle when to display:
// essentially just inserting newlines at the aprpriate places
func (lw *LineWrapper) Wrap(data []byte) string {
	buffer := wrap_impl(data, lw)

	return buffer.String()
}

// wrap_impl performs the core word wrapping logic.

func wrap_impl(data []byte, lw *LineWrapper) bytes.Buffer {
	var buffer bytes.Buffer

	for i, b := range data {
		// Debug stuff I want to leave for future usage
		// if b < 32 || b > 126 {
		// 	fmt.Printf("Special character: %q (ASCII: %d)\n", b, b)
		// } else {
		// 	fmt.Printf("Character: %q (ASCII: %d)\n", b, b)
		// }

		switch b {
		case '\n':
			buffer.WriteByte('\n')
			lw.currWidth = 0
		case '\t':
			spaces := strings.Repeat(" ", lw.tabWidth)
			lw.currWidth += lw.tabWidth
			buffer.WriteString(spaces)
		case ' ':
			// Don't write a space if we're at the beginning of a line...
			if lw.currWidth != 0 {
				buffer.WriteByte(' ')
			}
			// ...unless the next character is a space (and is within bounds),
			// then we want to write it. This is for code indentation. However,
			// there is the possibility that we are reading a line that ends on
			// exactly maxWidth, followed by two spaces on the beginning of the
			// next line, both of which would be printed even though we
			// wouldn't want that. I'm not sure how to account for that.
			if lw.currWidth == 0 && i+1 < len(data) {
				if data[i+1] == ' ' {
					buffer.WriteByte(' ')
				}
			}

			// We still need to count the width; otherwise, all initial spaces
			// will likely be ignored
			lw.currWidth++

			if lw.currWidth >= lw.maxWidth {
				buffer.WriteByte('\n')
				lw.currWidth = 0
			}
		default:
			buffer.WriteByte(b)
			lw.currWidth++
		}
	}

	return buffer
}
