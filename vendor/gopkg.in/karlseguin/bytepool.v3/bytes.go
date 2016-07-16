package bytepool

import (
	"encoding/binary"
	"io"
)

type bytes interface {
	position(n uint) bytes
	writeByte(b byte) (bytes, error)
	write(b []byte) (bytes, int, error)
	readNFrom(n int64, r io.Reader) (bytes, int64, error)

	Len() int
	Bytes() []byte
	String() string
	ReadByte() (byte, error)
	Read(b []byte) (int, error)
	WriteTo(w io.Writer) (int64, error)
}

type Bytes struct {
	bytes
	pool    *Pool
	fixed   *fixed
	scratch []byte
	enc     binary.ByteOrder
}

func NewBytes(capacity int) *Bytes {
	return NewEndianBytes(capacity, binary.BigEndian)
}

func NewEndianBytes(capacity int, enc binary.ByteOrder) *Bytes {
	return newPooled(nil, capacity, enc)
}

func newPooled(pool *Pool, capacity int, enc binary.ByteOrder) *Bytes {
	b := &Bytes{
		enc:  enc,
		pool: pool,
		fixed: &fixed{
			capacity: capacity,
			bytes:    make([]byte, capacity),
		},
		scratch: make([]byte, 8),
	}
	if pool != nil {
		b.fixed.onExpand = pool.onExpand
	}
	b.bytes = b.fixed
	return b
}

// Set a custom OnExpand callback (overwriting the one already set).
// Only useful for when NewBytes is called directly as opposed
// to through the pool.
func (b *Bytes) SetOnExpand(callback func()) {
	b.fixed.onExpand = callback
}

// Write the bytes
func (b *Bytes) Write(data []byte) (n int, err error) {
	b.bytes, n, err = b.write(data)
	return n, err
}

// Write a byte
func (b *Bytes) WriteByte(d byte) (err error) {
	b.bytes, err = b.writeByte(d)
	return err
}

func (b *Bytes) WriteUint16(n uint16) {
	b.enc.PutUint16(b.scratch, n)
	b.bytes, _, _ = b.write(b.scratch[:2])
}

func (b *Bytes) WriteUint32(n uint32) {
	b.enc.PutUint32(b.scratch, n)
	b.bytes, _, _ = b.write(b.scratch[:4])
}

func (b *Bytes) WriteUint64(n uint64) {
	b.enc.PutUint64(b.scratch, n)
	b.bytes, _, _ = b.write(b.scratch[:8])
}

func (b *Bytes) ReadByte() (byte, error) {
	bt, err := b.bytes.ReadByte()
	return bt, err
}

func (b *Bytes) ReadUint16() (uint16, error) {
	n, _ := b.bytes.Read(b.scratch[:2])
	if n == 2 {
		return b.enc.Uint16(b.scratch), nil
	}
	return 0, io.EOF
}

func (b *Bytes) ReadUint32() (uint32, error) {
	n, _ := b.bytes.Read(b.scratch[:4])
	if n == 4 {
		return b.enc.Uint32(b.scratch), nil
	}
	return 0, io.EOF
}

func (b *Bytes) ReadUint64() (uint64, error) {
	n, _ := b.bytes.Read(b.scratch[:8])
	if n == 8 {
		return b.enc.Uint64(b.scratch), nil
	}
	return 0, io.EOF
}

// Write a string
func (b *Bytes) WriteString(str string) (int, error) {
	return b.Write([]byte(str))
}

// Read from the io.Reader
func (b *Bytes) ReadFrom(r io.Reader) (n int64, err error) {
	return b.ReadNFrom(0, r)
}

// Read N bytes from the io.Reader
func (b *Bytes) ReadNFrom(n int64, r io.Reader) (m int64, err error) {
	b.bytes, m, err = b.readNFrom(n, r)
	return m, err
}

func (b *Bytes) Position(n uint) {
	b.bytes = b.position(n)
}

// Reset the object without releasing it
func (b *Bytes) Reset() {
	b.fixed.reset()
	b.bytes = b.fixed
}

// Release the item back into the pool
func (b *Bytes) Release() {
	if b.pool != nil {
		b.Reset()
		b.pool.list <- b
	}
}

// Alias for Release
func (b *Bytes) Close() error {
	b.Release()
	return nil
}
