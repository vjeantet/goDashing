package bytepool

import (
	stdbytes "bytes"
	"io"
)

type fixed struct {
	r        int
	length   int
	capacity int
	bytes    []byte
	onExpand func()
}

func (f *fixed) Len() int {
	return f.length
}

func (f *fixed) Bytes() []byte {
	return f.bytes[:f.length]
}

func (f *fixed) String() string {
	return string(f.Bytes())
}

func (f *fixed) write(data []byte) (bytes, int, error) {
	if l := len(data); f.hasSpace(l) == false {
		buf := f.toBuffer()
		n, err := buf.Write(data)
		return buf, n, err
	}
	n := copy(f.bytes[f.length:], data)
	f.length += n
	return f, n, nil
}

func (f *fixed) writeByte(data byte) (bytes, error) {
	if f.length == f.capacity {
		buf := f.toBuffer()
		err := buf.WriteByte(data)
		return buf, err
	}
	f.bytes[f.length] = data
	f.length++
	return f, nil
}

func (f *fixed) position(n uint) bytes {
	nn := int(n)
	if nn >= f.capacity {
		return f.toBuffer().position(n)
	}
	f.length = nn
	return f
}

func (f *fixed) readNFrom(expected int64, reader io.Reader) (bytes, int64, error) {
	ex := int(expected)
	if f.hasSpace(ex) == false {
		return f.toBuffer().readNFrom(expected, reader)
	}
	end := f.capacity
	if ex != 0 {
		end = ex + f.length
	}

	read := 0
	for {
		if f.full() {
			buf := f.toBuffer()
			_, n, err := buf.readNFrom(expected, reader)
			return buf, int64(read) + n, err
		}
		r, err := reader.Read(f.bytes[f.length:end])
		read += r
		f.length += r
		if err == io.EOF || (expected != 0 && read == ex) {
			return f, int64(read), err
		}
		if err != nil {
			return f, int64(read), err
		}
	}
}

func (f *fixed) WriteTo(w io.Writer) (n int64, err error) {
	r := f.r
	l := f.length
	for r < l {
		n, err := w.Write(f.bytes[r:l])
		if err != nil {
			break
		}
		r += n
	}
	f.reset()
	f.r += int(n)
	return n, err
}

func (f *fixed) Read(data []byte) (int, error) {
	if f.r == f.length {
		return 0, io.EOF
	}
	n := copy(data, f.bytes[f.r:f.length])
	f.r += n
	if f.r == f.length {
		return n, io.EOF
	}
	return n, nil
}

func (f *fixed) ReadByte() (byte, error) {
	if f.r == f.length {
		return 0, io.EOF
	}
	b := f.bytes[f.r]
	f.r += 1
	return b, nil
}

func (f *fixed) toBuffer() *buffer {
	if f.onExpand != nil {
		f.onExpand()
	}
	buf := &buffer{stdbytes.NewBuffer(f.bytes[f.r:])}
	buf.Truncate(f.length - f.r)
	return buf
}

func (f *fixed) hasSpace(toAdd int) bool {
	return f.length+toAdd <= f.capacity
}

func (f *fixed) full() bool {
	return f.length == f.capacity
}

func (f *fixed) reset() {
	f.r = 0
	f.length = 0
}
