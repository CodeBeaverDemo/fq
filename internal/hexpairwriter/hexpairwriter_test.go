package hexpairwriter_test

import (
    "bytes"
    "log"
    "testing"
    "errors"
"strings"

    "github.com/wader/fq/internal/hexpairwriter"
)
// errorWriter is a fake writer that always returns a write error.
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
    return 0, errors.New("simulated write error")
}

func TestWrite(t *testing.T) {
    b := &bytes.Buffer{}
    h := hexpairwriter.New(b, 4, 0, hexpairwriter.Pair)
    _, _ = h.Write([]byte(""))
    _, _ = h.Write([]byte("ab"))
    _, _ = h.Write([]byte("c"))
    _, _ = h.Write([]byte("d"))

    log.Printf("b.Bytes(): '%s'\n", b.Bytes())
}

func TestStartLineOffset(t *testing.T) {
    // Test writer with startLineOffset > 0 to ensure that the initial offset lines are written.
    b := &bytes.Buffer{}
    // Use width=4 and a startLineOffset=5 so that the pre-written blank lines are output.
    h := hexpairwriter.New(b, 4, 5, hexpairwriter.Pair)
    // Writing a single byte triggers the branch that writes the data after pre-filling.
    if _, err := h.Write([]byte("A")); err != nil {
        t.Fatalf("Write error: %v", err)
    }
    out := b.String()
    if out == "" {
        t.Errorf("Expected non-empty output, got empty")
    }
    t.Logf("TestStartLineOffset output:\n%s", out)
}
func TestWriteMultipleCalls(t *testing.T) {
    // Test writing data in multiple calls so the internal buffer properly flushes complete lines.
    b := &bytes.Buffer{}
    // width=3 so that every 3 bytes a newline is inserted.
    h := hexpairwriter.New(b, 3, 0, hexpairwriter.Pair)
    // Write first chunk (partial line)
    chunk1 := []byte("ab")
    if _, err := h.Write(chunk1); err != nil {
        t.Fatalf("Write error on chunk1: %v", err)
    }
    // Write second chunk (causing one full line and one partial line)
    chunk2 := []byte("cdef")
    if _, err := h.Write(chunk2); err != nil {
        t.Fatalf("Write error on chunk2: %v", err)
    }
    out := b.String()
    if out == "" {
        t.Errorf("Expected non-empty output, got empty")
    }
    t.Logf("TestWriteMultipleCalls output:\n%s", out)
}
func TestWithCustomFn(t *testing.T) {
    // Test using a custom formatting function, here wrapping the byte in brackets.
    customFn := func(b byte) string {
        return "[" + string(b) + "]"
    }
    b := &bytes.Buffer{}
    // Use width=5 so that we don’t always complete a full line.
    h := hexpairwriter.New(b, 5, 0, customFn)
    input := []byte("hello")
    if _, err := h.Write(input); err != nil {
        t.Fatalf("Write error: %v", err)
    }
    out := b.String()
    if out == "" {
        t.Errorf("Expected non-empty output, got empty")
    }
    t.Logf("TestWithCustomFn output:\n%s", out)
}
// TestEmptyInput verifies that writing an empty slice produces no output.
func TestEmptyInput(t *testing.T) {
    b := &bytes.Buffer{}
    h := hexpairwriter.New(b, 4, 0, hexpairwriter.Pair)
    n, err := h.Write([]byte(""))
    if err != nil {
        t.Fatalf("Write error: %v", err)
    }
    if n != 0 {
        t.Errorf("Expected 0 written bytes, got %d", n)
    }
    if b.Len() != 0 {
        t.Errorf("Expected empty output, got %q", b.String())
    }
}

// TestWriteError verifies that Write properly propagates errors from the underlying writer.
func TestWriteError(t *testing.T) {
    // errorWriter is a fake writer that always returns an error.

    w := &errorWriter{}
    h := hexpairwriter.New(w, 4, 0, hexpairwriter.Pair)
    _, err := h.Write([]byte("abc"))
    if err == nil {
        t.Error("Expected error, got nil")
    }
}

// TestCompleteLineFlush verifies that the writer flushes complete lines as expected.
func TestCompleteLineFlush(t *testing.T) {
    b := &bytes.Buffer{}
    h := hexpairwriter.New(b, 2, 0, hexpairwriter.Pair)
    // Write "abcd" which should produce one complete line and one partial line.
    _, err := h.Write([]byte("abcd"))
    if err != nil {
        t.Fatalf("Write error: %v", err)
    }
    expected := hexpairwriter.Pair('a') + " " + hexpairwriter.Pair('b') + "\n" +
        hexpairwriter.Pair('c') + " " + hexpairwriter.Pair('d')
    got := b.String()
    if got != expected {
        t.Errorf("Expected output:\n%q\ngot:\n%q", expected, got)
    }
}

// TestCustomFnEmpty verifies behavior when the custom formatting function returns an empty string.
func TestCustomFnEmpty(t *testing.T) {
    emptyFn := func(b byte) string { return "" }
    b := &bytes.Buffer{}
    h := hexpairwriter.New(b, 3, 0, emptyFn)
    _, err := h.Write([]byte("abc"))
    if err != nil {
        t.Fatalf("Write error: %v", err)
    }
    got := b.String()
    if len(got) == 0 {
        t.Errorf("Expected non-empty output when custom function returns empty string, got empty")
    }
}
func TestWriteOneByteAtATime(t *testing.T) {
    // Test writing one byte at a time to ensure flushes occur correctly for each call when
    // the line width is reached. This helps validate the per-byte flushing logic.
    var b bytes.Buffer
    h := hexpairwriter.New(&b, 4, 0, hexpairwriter.Pair)
    for i := 0; i < 16; i++ {
        if _, err := h.Write([]byte{byte(i)}); err != nil {
            t.Fatalf("Write error at iteration %d: %v", i, err)
        }
    }
    out := b.String()
    expected := hexpairwriter.Pair(0) + " " + hexpairwriter.Pair(1) + " " + hexpairwriter.Pair(2) + " " + hexpairwriter.Pair(3) + "\n" +
                hexpairwriter.Pair(4) + " " + hexpairwriter.Pair(5) + " " + hexpairwriter.Pair(6) + " " + hexpairwriter.Pair(7) + "\n" +
                hexpairwriter.Pair(8) + " " + hexpairwriter.Pair(9) + " " + hexpairwriter.Pair(10) + " " + hexpairwriter.Pair(11) + "\n" +
                hexpairwriter.Pair(12) + " " + hexpairwriter.Pair(13) + " " + hexpairwriter.Pair(14) + " " + hexpairwriter.Pair(15)
    if out != expected {
        t.Errorf("Expected output:\n%q\ngot:\n%q", expected, out)
    }
}

func TestWidthOne(t *testing.T) {
    // Test using width=1 to ensure that every written byte (when formatted) appears on its own line.
    var b bytes.Buffer
    h := hexpairwriter.New(&b, 1, 0, hexpairwriter.Pair)
    _, err := h.Write([]byte("abcd"))
    if err != nil {
        t.Fatalf("Write error: %v", err)
    }
    // With a width of one, every written byte should be flushed as its own line.
    expected := hexpairwriter.Pair('a') + "\n" +
                hexpairwriter.Pair('b') + "\n" +
                hexpairwriter.Pair('c') + "\n" +
                hexpairwriter.Pair('d')
    if b.String() != expected {
            t.Errorf("Expected %q, got %q", expected, b.String())
    }
}
// TestLongInputFlushes writes a long input slice at once to verify that complete lines flush correctly
func TestLongInputFlushes(t *testing.T) {
    const width = 4
    const numBytes = 20
    // Create an input slice with 20 sequential bytes (0 to 19)
    input := make([]byte, numBytes)
    for i := 0; i < numBytes; i++ {
        input[i] = byte(i)
    }

    b := &bytes.Buffer{}
    h := hexpairwriter.New(b, width, 0, hexpairwriter.Pair)
    if _, err := h.Write(input); err != nil {
        t.Fatalf("Write error: %v", err)
    }

    out := b.String()
    // Every time we complete a full line (except possibly the last partial one) the writer flushes with a newline.
    // For width=4 and numBytes=20 bytes, we expect flushes after 4,8,12,16 bytes => 4 newlines total.
    newlineCount := strings.Count(out, "\n")
    const expectedNewlines = 4
    if newlineCount != expectedNewlines {
        t.Errorf("Expected %d newlines, got %d", expectedNewlines, newlineCount)
    }
    // Verify that the output is non-empty and the last character is not a trailing space.
    if out == "" {
        t.Errorf("Expected non-empty output, got empty")
    }
    if out[len(out)-1] == ' ' {
        t.Errorf("Output should not end with a space, got %q", out)
    }
}

// TestConstantFn verifies that a custom formatting function returning a constant string works as expected.
func TestConstantFn(t *testing.T) {
    // customFn always returns "const" irrespective of the input byte.
    customFn := func(b byte) string {
        return "const"
    }
    b := &bytes.Buffer{}
    // For width=3 writing three characters should produce a single flushed line.
    h := hexpairwriter.New(b, 3, 0, customFn)
    input := []byte("ABC")
    if _, err := h.Write(input); err != nil {
        t.Fatalf("Write error: %v", err)
    }
    out := b.String()
    // The expected output is the concatenation of "const" for each input, separated by a space.
    // Since the final flush is issued on the last byte without the extra trailing space,
    // the expected output is:
    // "const const const"
    expected := "const const const"
    if out != expected {
        t.Errorf("Expected output:\n%q\ngot:\n%q", expected, out)
    }
}