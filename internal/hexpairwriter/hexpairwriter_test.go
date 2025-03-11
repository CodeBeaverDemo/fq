package hexpairwriter_test

import (
    "bytes"
    "log"
    "testing"

    "errors"
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
// TestBasic tests a simple Write call with width=2 and startLineOffset=0.
func TestBasic(t *testing.T) {
    var buf bytes.Buffer
    // width=2, startLineOffset=0 so that no initial padding is added.
    h := hexpairwriter.New(&buf, 2, 0, hexpairwriter.Pair)
    _, err := h.Write([]byte("ab"))
    if err != nil {
    t.Fatalf("Write returned error: %v", err)
    }
    // 'a' (ASCII 97) should be "61" and 'b' (ASCII 98) "62", with a space separator.
    // The expected output: "61 62"  (no trailing space)
    expected := "61 62"
    got := buf.String()
    if got != expected {
    t.Errorf("Expected %q, got %q", expected, got)
    }
}

// TestWriteBoundary tests writing a number of bytes that spans a newline boundary.
// For width=2 and input "abc", after writing two bytes a line boundary should be triggered.
func TestWriteBoundary(t *testing.T) {
    var buf bytes.Buffer
    h := hexpairwriter.New(&buf, 2, 0, hexpairwriter.Pair)
    // Write "abc" in one call.
    _, err := h.Write([]byte("abc"))
    if err != nil {
    t.Fatalf("Write returned error: %v", err)
    }
    // Let's simulate what happens:
    //  - For the first byte 'a' (offset 0), writer appends "61 " (no newline triggered).
    //  - For the second byte 'b' (offset 1, which is width-1), the trailing space is replaced by a newline.
    //    So that part becomes "62\n" and is written; then the third byte 'c' is written on a fresh buffer.
    //  - For the third byte 'c' (offset 2, new chunk) we have "63" (with no trailing space).
    expected := "61 62\n63"
    got := buf.String()
    if got != expected {
    t.Errorf("Expected %q, got %q", expected, got)
    }
}

// TestMultipleCalls tests that subsequent calls to Write embed the proper header (a space or newline)
// when h.offset > h.startLineOffset.
func TestMultipleCalls(t *testing.T) {
    var buf bytes.Buffer
    // Use width=3 and startLineOffset=2 to generate header output on the first few writes.
    h := hexpairwriter.New(&buf, 3, 2, hexpairwriter.Pair)

    // First call: write one byte. Since h.offset is initially 0 and must catch up to startLineOffset,
    // the writer will output initial spacing.
    _, err := h.Write([]byte("A"))
    if err != nil {
    t.Fatalf("First Write returned error: %v", err)
    }

    // Second call: write two more bytes.
    _, err = h.Write([]byte("BC"))
    if err != nil {
    t.Fatalf("Second Write returned error: %v", err)
    }

    // Third call: write three bytes to exercise a full row.
    _, err = h.Write([]byte("DEF"))
    if err != nil {
    t.Fatalf("Third Write returned error: %v", err)
    }

    out := buf.String()
    // Since the output formatting involves leading blanks for startLineOffset and newline insertion on row boundaries,
    // we do a crude check that the output includes newline characters and hex pairs.
    if len(out) == 0 || !containsNewline(out) {
    t.Errorf("Output %q does not seem to have the expected newline formatting", out)
    }
}

// containsNewline is a helper to check if a string contains a newline.
func containsNewline(s string) bool {
    for i := 0; i < len(s); i++ {
    if s[i] == '\n' {
    return true
    }
    }
    return false
}

// TestErrorPropagation tests that if the underlying writer returns an error, Write propagates it.
func TestErrorPropagation(t *testing.T) {
    // Define a simple writer that always returns an error.
    errWriter := &errorWriter{}
    h := hexpairwriter.New(errWriter, 4, 0, hexpairwriter.Pair)
    _, err := h.Write([]byte("test"))
    if err == nil {
    t.Error("Expected an error, got nil")
    }
}
// TestCustomFn tests using a custom formatting function to convert bytes.
func TestCustomFn(t *testing.T) {
    var buf bytes.Buffer
    customFn := func(b byte) string {
        return "CUSTOM:" + string(b)
    }
    // Use width=3 and startLineOffset=0 for simplicity.
    h := hexpairwriter.New(&buf, 3, 0, customFn)
    _, err := h.Write([]byte("ABC"))
    if err != nil {
        t.Fatalf("Write returned error: %v", err)
    }
    // Expected output: "CUSTOM:A CUSTOM:B CUSTOM:C"
    expected := "CUSTOM:A CUSTOM:B CUSTOM:C"
    if got := buf.String(); got != expected {
        t.Errorf("Expected %q, got %q", expected, got)
    }
}
// TestPartialWriteAcrossBoundaries tests writing bytes across line boundaries with multiple Write calls.
func TestPartialWriteAcrossBoundaries(t *testing.T) {
    var buf bytes.Buffer
    // Use width=4 and startLineOffset=0.
    h := hexpairwriter.New(&buf, 4, 0, hexpairwriter.Pair)
    // Write first part: 3 bytes.
    _, err := h.Write([]byte("abc"))
    if err != nil {
        t.Fatalf("First Write returned error: %v", err)
    }
    // Write second part: 3 more bytes.
    _, err = h.Write([]byte("def"))
    if err != nil {
        t.Fatalf("Second Write returned error: %v", err)
    }
    // Expected: first line completes with 4 bytes ("a", "b", "c", "d") and adds a newline,
    // while the second line contains the remaining two bytes ("e", "f") with no trailing space.
    // The default Pair function converts: a(97)="61", b(98)="62", c(99)="63", d(100)="64", e(101)="65", f(102)="66".
    expected := "61 62 63 64\n65 66"
    if got := buf.String(); got != expected {
        t.Errorf("Expected %q, got %q", expected, got)
    }
}

// errorWriter is an io.Writer that always returns an error.
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
    return 0, errors.New("injected write error")
}
