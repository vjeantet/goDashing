package bytepool

import (
	stdbytes "bytes"
	"io"
)

type buffer struct {
	*stdbytes.Buffer
}

func (b *buffer) write(data []byte) (bytes, int, error) {
	n, err := b.Write(data)
	return b, n, err
}

func (b *buffer) writeByte(data byte) (bytes, error) {
	err := b.WriteByte(data)
	return b, err
}

func (b *buffer) position(n uint) bytes {
	s := b.Len()
	nn := int(n)
	t := nn - s
	if t == 0 {
		return b
	}
	if t < 0 {
		b.Truncate(nn)
		return b
	}
	b.Grow(t)
	bytes := b.Bytes()
	bytes = bytes[:nn]
	b.Write(bytes[s:nn])
	return b
}

func (b *buffer) readNFrom(n int64, r io.Reader) (bytes, int64, error) {
	if n == 0 {
		m, err := b.ReadFrom(r)
		return b, m, err
	}
	s := b.Len()
	t := int(n) + s
	b.Grow(t)
	bytes := b.Bytes()
	bytes = bytes[:t]
	m, err := io.ReadFull(r, bytes[s:t])
	b.Write(bytes[s:t])
	return b, int64(m), err
}
