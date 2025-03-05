package linewrap

import (
	"bytes"
	// "fmt"
	"io"
	"strings"
)

type LineWrapper struct {
	maxWidth  int
	currWidth int
	tabWidth  int
	writer    io.Writer
}

func NewLineWrapper(maxWidth, tabWidth int, lwWriter io.Writer) *LineWrapper {
	return &LineWrapper{
		maxWidth: maxWidth,
		tabWidth: tabWidth,
		writer:   lwWriter,
	}
}

// Allow wrapping for input that comes in chunks, versus building the line and
// then splitting it and printing the whole line at once.
func (lw *LineWrapper) Write(data []byte) (n int, err error) {
	var buffer bytes.Buffer

	// Special case for single space - preserve it regardless of position
	if len(data) == 1 && data[0] == ' ' {
		buffer.WriteByte(' ')
		lw.currWidth++
		if lw.currWidth >= lw.maxWidth {
			buffer.WriteByte('\n')
			lw.currWidth = 0
		}
		lw.writer.Write(buffer.Bytes())
		return 1, nil
	}

	for _, b := range data {
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
			// MODIFIED: Always write spaces in streaming mode
			buffer.WriteByte(' ')

			// We still need to count the width
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

	lw.writer.Write(buffer.Bytes())

	return len(data), nil
}
