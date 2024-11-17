package linewrap

import (
	"bytes"
	"fmt"
	"strings"
)

type LineWrapper struct {
	maxWidth  int
	currWidth int
	tabWidth  int
}

func NewLineWrapper(maxWidth int, tabWidth int) *LineWrapper {
	return &LineWrapper{
		maxWidth: maxWidth,
		tabWidth: tabWidth,
	}
}

func (lw *LineWrapper) Write(data []byte) (n int, err error) {
	var buffer bytes.Buffer

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
			buffer.WriteByte(' ')
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

	fmt.Print(buffer.String())
	return len(data), nil
}
