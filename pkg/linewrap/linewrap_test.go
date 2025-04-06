package linewrap

import (
	"bytes"
	"testing"
)

func TestLineWrapper_Write(t *testing.T) {
	t.Run("simple text", func(t *testing.T) {
		var buffer bytes.Buffer
		lw := NewLineWrapper(20, 4, &buffer)
		expectedOutput := "This is a\nvery long\ntest string."
		n := lw.Write([]byte(expectedOutput))
		if n != len(expectedOutput) {
			t.Errorf("expected output: %q but length incorrect: %d (%d)", expectedOutput, n, len(expectedOutput))
		}
		got := buffer.String()
		if got != expectedOutput {
			t.Errorf("got: %q, want: %q", got, expectedOutput)
		}
	})

	t.Run("tabs and spaces", func(t *testing.T) {
		var buffer bytes.Buffer
		lw := NewLineWrapper(20, 4, &buffer)

		input := "This\tis a very long test string."
		expectedOutput := "This    is a very long \ntest string."

		n := lw.Write([]byte(input))
		if n != len(expectedOutput) {
			t.Errorf("expected output: %q but length incorrect: %d (%d)", expectedOutput, n, len(expectedOutput))
		}
		got := buffer.String()
		if got != expectedOutput {
			t.Errorf("got: %q, want: %q", got, expectedOutput)
		}
	})

	t.Run("long line breaks correctly", func(t *testing.T) {
		var buffer bytes.Buffer
		lw := NewLineWrapper(10, 4, &buffer)
		input := "This is a very long test string that should be broken into\nmultiple lines."
		expectedOutput := "This is a \nvery long \ntest string \nthat should \nbe broken \ninto\nmultiple lines."
		n := lw.Write([]byte(input))
		if n != len(expectedOutput) {
			t.Errorf("expected output: %q but got length incorrect: %d (%d)", expectedOutput, n, len(expectedOutput))
		}
		got := buffer.String()
		if got != expectedOutput {
			t.Errorf("got: %q, want: %q", got, expectedOutput)
		}
	})

	t.Run("multiple spaces and tabs on the same line", func(t *testing.T) {
		var buffer bytes.Buffer
		lw := NewLineWrapper(20, 4, &buffer)

		input := "This is a\tvery long   test string with multiple spaces\tand tabs."
		expectedOutput := "This is a    very long \n  test string with multiple \nspaces    and tabs."

		n := lw.Write([]byte(input))
		if n != len(expectedOutput) {
			t.Errorf("expected output: %q but got length incorrect: %d (%d)", expectedOutput, n, len(expectedOutput))
		}
		got := buffer.String()
		if got != expectedOutput {
			t.Errorf("got: %q, want: %q", got, expectedOutput)
		}
	})

	t.Run("line break at the end of a line with multiple spaces and tabs", func(t *testing.T) {
		var buffer bytes.Buffer
		lw := NewLineWrapper(20, 4, &buffer)
		input := "This is a very long test string with multiple\nspaces and\ttabs."
		expectedOutput := "This is a very long \ntest string with multiple\nspaces and    tabs."
		n := lw.Write([]byte(input))
		if n != len(expectedOutput) {
			t.Errorf("expected output: %q but got length incorrect: %d (%d)", expectedOutput, n, len(expectedOutput))
		}
		got := buffer.String()
		if got != expectedOutput {
			t.Errorf("got: %q, want: %q", got, expectedOutput)
		}
	})
}

func TestLineWrapper_WriteNil(t *testing.T) {
	var buffer bytes.Buffer
	lw := NewLineWrapper(20, 4, &buffer)
	var data []byte
	n := lw.Write(data)
	if n != 0 {
		t.Errorf("expected output: %v but got length incorrect: %d", 0, n)
	}
}

func TestLineWrapper_WriteEmpty(t *testing.T) {
	var buffer bytes.Buffer
	lw := NewLineWrapper(20, 4, &buffer)
	data := make([]byte, 0)
	n := lw.Write(data)
	if n != 0 {
		t.Errorf("expected output: %v but got length incorrect: %d", 0, n)
	}
}

func TestLineWrapper_WriteBuffer(t *testing.T) {
	var buffer bytes.Buffer
	lw := NewLineWrapper(20, 4, &buffer)
	data := []byte{}
	n := lw.Write(data)
	if n != len(data) {
		t.Errorf("expected output: %v but got length incorrect: %v", len(data), n)
	}
}
