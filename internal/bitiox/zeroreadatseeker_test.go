package bitiox

import (
    "io"
    "testing"

    "github.com/wader/fq/pkg/bitio"
)

// TestZeroReadAtSeeker_SeekBits tests the SeekBits method including valid operations and error conditions.
func TestZeroReadAtSeeker_SeekBits(t *testing.T) {
    const totalBits = 64
    zr := NewZeroAtSeeker(totalBits)

    // Valid Seek using io.SeekStart.
    pos, err := zr.SeekBits(10, io.SeekStart)
    if err != nil {
    t.Fatalf("expected no error, got %v", err)
    }
    if pos != 10 {
    t.Fatalf("expected pos 10, got %d", pos)
    }

    // Valid Seek using io.SeekCurrent.
    pos, err = zr.SeekBits(5, io.SeekCurrent)
    if err != nil {
    t.Fatalf("expected no error, got %v", err)
    }
    if pos != 15 {
    t.Fatalf("expected pos 15, got %d", pos)
    }

    // Valid Seek using io.SeekEnd.
    pos, err = zr.SeekBits(-10, io.SeekEnd)
    if err != nil {
    t.Fatalf("expected no error, got %v", err)
    }
    if pos != totalBits-10 {
    t.Fatalf("expected pos %d, got %d", totalBits-10, pos)
    }

    // Test error: negative position with io.SeekStart.
    _, err = zr.SeekBits(-1, io.SeekStart)
    if err != bitio.ErrOffset {
    t.Errorf("expected ErrOffset for negative seek, got %v", err)
    }

    // Test error: position greater than total bits with io.SeekStart.
    _, err = zr.SeekBits(totalBits+1, io.SeekStart)
    if err != bitio.ErrOffset {
    t.Errorf("expected ErrOffset for excessive seek, got %v", err)
    }

    // Test panic for unknown whence.
    func() {
    defer func() {
    if r := recover(); r == nil {
    t.Errorf("expected panic for unknown whence, but no panic occurred")
    }
    }()
    // This should panic.
    zr.SeekBits(10, 999)
    }()
}

// TestZeroReadAtSeeker_ReadBitsAt tests the ReadBitsAt method including edge cases and partial reads.
func TestZeroReadAtSeeker_ReadBitsAt(t *testing.T) {
    const totalBits = 64
    zr := NewZeroAtSeeker(totalBits)

    // Test error: negative bitOff.
    buf := make([]byte, 10)
    _, err := zr.ReadBitsAt(buf, 8, -1)
    if err != bitio.ErrOffset {
    t.Errorf("expected ErrOffset for negative bitOff, got %v", err)
    }

    // Test EOF when bitOff equals total bits.
    _, err = zr.ReadBitsAt(buf, 8, totalBits)
    if err != io.EOF {
    t.Errorf("expected EOF when bitOff equals totalBits, got %v", err)
    }

    // Test reading within available bits: reading 20 bits from bitOff=0.
    n, err := zr.ReadBitsAt(buf, 20, 0)
    if err != nil {
    t.Fatalf("expected no error, got %v", err)
    }
    if n != 20 {
    t.Errorf("expected to read 20 bits, got %d", n)
    }
    // Verify that the corresponding bytes are zero.
    rBytes := bitio.BitsByteCount(20)
    for i := int64(0); i < rBytes; i++ {
    if buf[i] != 0 {
    t.Errorf("expected byte at index %d to be 0, got %d", i, buf[i])
    }
    }

    // Test reading beyond available bits: request 20 bits starting from bitOff=50, available = 14 bits.
    n, err = zr.ReadBitsAt(buf, 20, 50)
    if err != nil {
    t.Fatalf("expected no error, got %v", err)
    }
    if n != 14 {
    t.Errorf("expected to read 14 bits, got %d", n)
    }
}

// TestZeroReadAtSeeker_CloneReadAtSeeker tests cloning of the reader and confirms independent operation.
func TestZeroReadAtSeeker_CloneReadAtSeeker(t *testing.T) {
    const totalBits = 64
    zr := NewZeroAtSeeker(totalBits)

    // Advance the original seeker.
    _, err := zr.SeekBits(10, io.SeekStart)
    if err != nil {
    t.Fatalf("unexpected error: %v", err)
    }

    // Clone the seeker.
    clone, err := zr.CloneReadAtSeeker()
    if err != nil {
    t.Fatalf("unexpected error in clone: %v", err)
    }

    // The clone should be independent and start at position 0.
    posClone, err := clone.SeekBits(0, io.SeekCurrent)
    if err != nil {
    t.Fatalf("unexpected error on clone seek: %v", err)
    }
    if posClone != 0 {
    t.Errorf("expected cloned seeker pos to be 0, got %d", posClone)
    }

    // Changing clone's position should not affect original.
    _, err = clone.SeekBits(20, io.SeekStart)
    if err != nil {
    t.Fatalf("unexpected error on clone seek: %v", err)
    }
    posOrig, err := zr.SeekBits(0, io.SeekCurrent)
    if err != nil {
    t.Fatalf("unexpected error on original seek: %v", err)
    }
    if posOrig != 10 {
    t.Errorf("expected original seeker pos to remain 10, got %d", posOrig)
    }
}
// TestZeroReadAtSeeker_ReadBitsAt_ZeroBits tests that reading zero bits returns 0 and no error.
func TestZeroReadAtSeeker_ReadBitsAt_ZeroBits(t *testing.T) {
    const totalBits = 32
    zr := NewZeroAtSeeker(totalBits)

    buf := make([]byte, 5)
    n, err := zr.ReadBitsAt(buf, 0, 0)
    if err != nil {
        t.Fatalf("expected no error for zero bits read, got %v", err)
    }
    if n != 0 {
        t.Errorf("expected to read 0 bits, got %d", n)
    }
}

// TestZeroReadAtSeeker_ReadBitsAt_PartialByte tests reading a number of bits that does not perfectly align to a full byte.
func TestZeroReadAtSeeker_ReadBitsAt_PartialByte(t *testing.T) {
    const totalBits = 16
    zr := NewZeroAtSeeker(totalBits)

    buf := make([]byte, 2)
    // Request 5 bits starting at offset 0.
    n, err := zr.ReadBitsAt(buf, 5, 0)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if n != 5 {
        t.Errorf("expected to read 5 bits, got %d", n)
    }
    // The first byte should have been set to 0.
    if buf[0] != 0 {
        t.Errorf("expected the first byte to be 0, got %d", buf[0])
    }
}

// TestZeroReadAtSeeker_SeekBits_NegativeCurrent tests that seeking backwards past the beginning returns an error.
func TestZeroReadAtSeeker_SeekBits_NegativeCurrent(t *testing.T) {
    const totalBits = 64
    zr := NewZeroAtSeeker(totalBits)

    // Move to 30 bits.
    _, err := zr.SeekBits(30, io.SeekStart)
    if err != nil {
        t.Fatalf("expected no error seeking to 30, got %v", err)
    }

    // Attempt to move backwards beyond the beginning.
    _, err = zr.SeekBits(-40, io.SeekCurrent)
    if err != bitio.ErrOffset {
        t.Errorf("expected ErrOffset for seeking before zero, got %v", err)
    }
}
// TestZeroReadAtSeeker_ReadBitsAt_BufferTooSmall tests that providing a buffer which is too small causes a panic.
func TestZeroReadAtSeeker_ReadBitsAt_BufferTooSmall(t *testing.T) {
    zr := NewZeroAtSeeker(64)
    buf := make([]byte, 1) // insufficient buffer: reading 16 bits requires at least 2 bytes.
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("expected panic due to insufficient buffer size, but no panic occurred")
        }
    }()
    _, _ = zr.ReadBitsAt(buf, 16, 0) // this should trigger a panic due to index out of range.
}

// TestZeroReadAtSeeker_FullRead tests that reading the full available bits works correctly.
func TestZeroReadAtSeeker_FullRead(t *testing.T) {
    const totalBits = 50
    zr := NewZeroAtSeeker(totalBits)
    bufSize := bitio.BitsByteCount(totalBits)
    buf := make([]byte, bufSize)
    n, err := zr.ReadBitsAt(buf, totalBits, 0)
    if err != nil {
        t.Fatalf("expected no error on full read, got %v", err)
    }
    if n != totalBits {
        t.Errorf("expected to read %d bits, got %d", totalBits, n)
    }
    for i, b := range buf {
        if b != 0 {
            t.Errorf("expected buffer byte %d to be 0, got %d", i, b)
        }
    }
}

// TestZeroReadAtSeeker_SeekBits_End tests seeking to the end position using io.SeekEnd with offset 0.
func TestZeroReadAtSeeker_SeekBits_End(t *testing.T) {
    const totalBits = 64
    zr := NewZeroAtSeeker(totalBits)
    pos, err := zr.SeekBits(0, io.SeekEnd)
    if err != nil {
        t.Fatalf("expected no error seeking to end, got %v", err)
    }
    if pos != totalBits {
        t.Errorf("expected position to be %d, got %d", totalBits, pos)
    }
}
// TestZeroReadAtSeeker_ReadBitsAt_BufferIntegrity tests that ReadBitsAt writes only the required bytes in the buffer
// while leaving additional bytes unchanged.
func TestZeroReadAtSeeker_ReadBitsAt_BufferIntegrity(t *testing.T) {
    const totalBits = 64
    zr := NewZeroAtSeeker(totalBits)
    // Prepare a buffer longer than needed and initialize extra bytes to a non-zero value.
    buf := []byte{0xFF, 0xFF, 0xFF, 0xFF}
    // Request to read 12 bits, which requires 2 bytes (since (12+7)/8 == 2).
    n, err := zr.ReadBitsAt(buf, 12, 0)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if n != 12 {
        t.Errorf("expected to read 12 bits, got %d", n)
    }
    // The first 2 bytes should be set to 0.
    if buf[0] != 0 || buf[1] != 0 {
        t.Errorf("expected first two bytes to be 0, got %v", buf[:2])
    }
    // The remaining bytes should remain unchanged (0xFF).
    if buf[2] != 0xFF || buf[3] != 0xFF {
        t.Errorf("expected remaining bytes to be unchanged, got %v", buf[2:4])
    }
}

// TestZeroReadAtSeeker_ZeroTotalBits tests that an instance with zero total bits behaves as expected.
func TestZeroReadAtSeeker_ZeroTotalBits(t *testing.T) {
    zr := NewZeroAtSeeker(0)

    // Seeking to 0 should be allowed.
    pos, err := zr.SeekBits(0, io.SeekStart)
    if err != nil {
        t.Fatalf("expected no error seeking to 0, got %v", err)
    }
    if pos != 0 {
        t.Errorf("expected position to be 0, got %d", pos)
    }

    // Attempting to seek to any positive value should return an ErrOffset.
    _, err = zr.SeekBits(1, io.SeekStart)
    if err != bitio.ErrOffset {
        t.Errorf("expected ErrOffset when seeking beyond 0, got %v", err)
    }

    // Reading at offset 0 should return EOF as there are no bits.
    buf := make([]byte, 1)
    _, err = zr.ReadBitsAt(buf, 1, 0)
    if err != io.EOF {
        t.Errorf("expected EOF when reading from an instance with zero total bits, got %v", err)
    }
}