package hexpairwriter_test

import (
    "bytes"
    "log"
    "testing"

    "fmt"
    "github.com/wader/fq/internal/hexpairwriter"
)

func TestWrite(t *testing.T) {
    b := &bytes.Buffer{}
    h := hexpairwriter.New(b, 4, 0, hexpairwriter.Pair)
    _, _ = h.Write([]byte(""))
    _, _ = h.Write([]byte("ab"))
    _, _ = h.Write([]byte("c"))
    _, _ = h.Write([]byte("d"))

    log.Printf("b.Bytes(): '%s'\n", b.Bytes())
}
func TestPairFunction(t *testing.T) {
    // Test that the Pair function returns a two-character hexadecimal string.
    tests := []struct {
    input    byte
    expected string
    }{
    {0, "00"},
    {1, "01"},
    {15, "0f"},
    {255, "ff"},
    }
    for _, tc := range tests {
    result := hexpairwriter.Pair(tc.input)
    if result != tc.expected {
    t.Errorf("Pair(%d) = %q; want %q", tc.input, result, tc.expected)
    }
    }
}

func TestWriteCompleteLine(t *testing.T) {
    // Test writing exactly a complete line (width=4) in one call.
    buf := &bytes.Buffer{}
    w := hexpairwriter.New(buf, 4, 0, hexpairwriter.Pair)
    // Write 4 bytes.
    _, err := w.Write([]byte{0, 1, 2, 3})
    if err != nil {
    t.Fatalf("Write failed: %v", err)
    }
    // Expected output: for each byte a hex pair and a space (except for no trailing space on the last pair).
    expected := "00 01 02 03"
    if got := buf.String(); got != expected {
    t.Errorf("got %q; expected %q", got, expected)
    }
}

func TestWriteSplitAcrossCalls(t *testing.T) {
    // Test that writing data in two successive Write calls produces a continuous output.
    buf := &bytes.Buffer{}
    w := hexpairwriter.New(buf, 4, 0, hexpairwriter.Pair)
    // First call: write 2 bytes.
    _, err := w.Write([]byte{0, 1})
    if err != nil {
    t.Fatalf("First Write failed: %v", err)
    }
    // Second call: write the next 2 bytes.
    _, err = w.Write([]byte{2, 3})
    if err != nil {
    t.Fatalf("Second Write failed: %v", err)
    }
    // For the first call, the writer writes "00 01"
    // For the second call, since there is an existing non-empty line state, a prefix will be inserted.
    // Expected overall output: "00 01" followed directly by " 02 03" (note the leading space).
    expected := "00 01 02 03"
    if got := buf.String(); got != expected {
    t.Errorf("got %q; expected %q", got, expected)
    }
}

func TestWriteCustomFn(t *testing.T) {
    // Test that a custom formatting function is correctly applied.
    customFn := func(b byte) string {
    // Instead of hex, return the corresponding letter starting from 'A'
    return "[" + string('A'+(b%26)) + "]"
    }
    buf := &bytes.Buffer{}
    w := hexpairwriter.New(buf, 4, 0, customFn)
    // Write 4 bytes.
    _, err := w.Write([]byte{0, 1, 2, 3})
    if err != nil {
    t.Fatalf("Write failed: %v", err)
    }
    // Expected: "[A] [B] [C] [D]" (with no trailing space after the last pair).
    expected := "[A] [B] [C] [D]"
    if got := buf.String(); got != expected {
    t.Errorf("got %q; expected %q", got, expected)
    }
}

func TestWriteWithStartLineOffset(t *testing.T) {
    // Test that providing a startLineOffset writes filler content (spaces or newline) before writing data.
    buf := &bytes.Buffer{}
    // With startLineOffset=2 and width=4, the writer should first output filler for two offsets.
    w := hexpairwriter.New(buf, 4, 2, hexpairwriter.Pair)
    // Write 3 bytes.
    _, err := w.Write([]byte{0, 1, 2})
    if err != nil {
    t.Fatalf("Write failed: %v", err)
    }
    // Simulate filler: for offset=0 and 1 an output of "   " each is produced.
    filler := "   " + "   "
    // Then for byte 0 (at offset 2): "00 " is appended.
    // For byte 1 (at offset 3, width-1), the trailing space is replaced by a newline and flushes: "01\n"
    // For byte 2 (at new line offset 0): "02" is written (last byte, no trailing space).
    expected := filler + "00 01\n02"
    if got := buf.String(); got != expected {
    t.Errorf("got %q; expected %q", got, expected)
    }
}
// TestWriteMultiLine tests writing data that spans multiple complete lines in a single Write call.
func TestWriteMultiLine(t *testing.T) {
    buf := &bytes.Buffer{}
    w := hexpairwriter.New(buf, 4, 0, hexpairwriter.Pair)
    // Write 8 bytes to produce 2 lines:
    // The first 4 bytes form a complete line which flushes with a newline.
    // The next 4 bytes start a new line without a trailing newline.
    data := []byte{0, 1, 2, 3, 4, 5, 6, 7}
    n, err := w.Write(data)
    if err != nil {
        t.Fatalf("Write returned error: %v", err)
    }
    if n != len(data) {
        t.Errorf("Write returned %d; expected %d", n, len(data))
    }
    expected := "00 01 02 03\n04 05 06 07"
    if got := buf.String(); got != expected {
        t.Errorf("got %q; expected %q", got, expected)
    }
}

// errorWriter is a test helper that simulates a failing underlying io.Writer.
type errorWriter struct {
    fail bool
}

func (ew *errorWriter) Write(p []byte) (int, error) {
    if ew.fail {
        return 0, fmt.Errorf("simulated write error")
    }
    return len(p), nil
}

// TestWriteError verifies that errors from the underlying writer are properly propagated.
func TestWriteError(t *testing.T) {
    ew := &errorWriter{fail: true}
    w := hexpairwriter.New(ew, 4, 0, hexpairwriter.Pair)
    _, err := w.Write([]byte{0})
    if err == nil {
        t.Fatal("Expected error from underlying writer, but got nil")
    }
}
func TestEmptyWriteWithFiller(t *testing.T) {
    // Test that writing an empty slice still writes the required filler when startLineOffset > current offset.
    buf := &bytes.Buffer{}
    // Using width=4 and startLineOffset=3 means the writer will first output filler for offsets 0, 1, and 2.
    // For offsets where (offset % width != width-1), filler "   " (3 spaces) is used.
    // Hence, the expected output is 9 spaces.
    w := hexpairwriter.New(buf, 4, 3, hexpairwriter.Pair)
    n, err := w.Write([]byte{})
    if err != nil {
        t.Fatalf("Write failed: %v", err)
    }
    if n != 0 {
        t.Errorf("Write returned %d; expected 0", n)
    }
    expected := "         " // 9 spaces
    if got := buf.String(); got != expected {
        t.Errorf("got %q; expected %q", got, expected)
    }
}

func TestEmptyWriteAfterData(t *testing.T) {
    // Test that writing an empty slice after writing data does not produce additional output or alter state.
    buf := &bytes.Buffer{}
    w := hexpairwriter.New(buf, 4, 0, hexpairwriter.Pair)
    // Write 2 bytes (an incomplete line).
    _, err := w.Write([]byte{0, 1})
    if err != nil {
        t.Fatalf("First Write failed: %v", err)
    }
    // Capture the output after the first write.
    out1 := buf.String()
    if out1 == "" {
        t.Errorf("Expected some output after first write, got empty string")
    }
    // Write an empty slice – this should not change the output or state.
    n, err := w.Write([]byte{})
    if err != nil {
        t.Fatalf("Empty Write failed: %v", err)
    }
    if n != 0 {
        t.Errorf("Empty Write returned %d; expected 0", n)
    }
    // Now complete the line by writing 2 more bytes.
    _, err = w.Write([]byte{2, 3})
    if err != nil {
        t.Fatalf("Second Write failed: %v", err)
    }
    // The overall expected output should be a single complete line: "00 01 02 03"
    expected := "00 01 02 03"
    if got := buf.String(); got != expected {
        t.Errorf("got %q; expected %q", got, expected)
    }
}
