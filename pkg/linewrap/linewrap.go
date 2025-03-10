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
	var wordBuffer bytes.Buffer

	for i := 0; i < len(data); i++ {
		b := data[i]

		switch b {
		case '\n':
			// Write any pending word
			if wordBuffer.Len() > 0 {
				buffer.Write(wordBuffer.Bytes())
				wordBuffer.Reset()
			}
			buffer.WriteByte('\n')
			lw.currWidth = 0

		case '\t':
			// Write any pending word
			if wordBuffer.Len() > 0 {
				buffer.Write(wordBuffer.Bytes())
				wordBuffer.Reset()
			}
			spaces := strings.Repeat(" ", lw.tabWidth)
			buffer.WriteString(spaces)
			lw.currWidth += lw.tabWidth

		case ' ':
			// Write any pending word
			if wordBuffer.Len() > 0 {
				// Check if word would exceed line length
				if lw.currWidth+wordBuffer.Len() > lw.maxWidth {
					buffer.WriteByte('\n')
					lw.currWidth = 0
				}
				buffer.Write(wordBuffer.Bytes())
				lw.currWidth += wordBuffer.Len()
				wordBuffer.Reset()
			}

			// Handle the space
			if lw.currWidth > 0 {
				buffer.WriteByte(' ')
				lw.currWidth++
			}

			// If we're at max width, start new line
			if lw.currWidth >= lw.maxWidth {
				buffer.WriteByte('\n')
				lw.currWidth = 0
			}

		default:
			// Accumulate characters into word buffer
			wordBuffer.WriteByte(b)

			// If this word alone exceeds max width, write it with a hyphen
			if wordBuffer.Len() >= lw.maxWidth {
				if lw.currWidth > 0 {
					buffer.WriteByte('\n')
				}
				buffer.Write(wordBuffer.Bytes()[:lw.maxWidth-1])
				buffer.WriteByte('-')
				buffer.WriteByte('\n')
				wordBuffer.Reset()
				wordBuffer.Write(data[i-wordBuffer.Len()+lw.maxWidth-1 : i+1])
				lw.currWidth = 0
			}
		}
	}

	// Write any remaining word
	if wordBuffer.Len() > 0 {
		if lw.currWidth+wordBuffer.Len() > lw.maxWidth {
			buffer.WriteByte('\n')
			lw.currWidth = 0
		}
		buffer.Write(wordBuffer.Bytes())
	}

	lw.writer.Write(buffer.Bytes())
	return len(data), nil
}

// isPunctuation returns true if the byte represents a punctuation character
func isPunctuation(b byte) bool {
	return strings.ContainsRune(",.!?;:)", rune(b))
}
