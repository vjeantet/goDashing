package gerb

import (
	"bytes"
	"github.com/karlseguin/gerb/core"
	"github.com/karlseguin/gerb/r"
)

type OutputTag struct {
	value      core.Value
	autoEscape bool
}

func (o *OutputTag) Execute(context *core.Context) core.ExecutionState {
	value := r.ToBytes(o.value.Resolve(context))
	if o.autoEscape {
		value = escape(value)
	}
	context.Writer.Write(r.ToBytes(value))
	return core.NormalState
}

func createOutputTag(p *core.Parser, isUnsafe bool) (core.Executable, error) {
	value, err := p.ReadValue()
	if err != nil {
		return nil, err
	}
	if err = p.ReadCloseTag(); err != nil {
		return nil, err
	}
	return &OutputTag{value, !isUnsafe}, nil
}

const escapedChars = "<>"

func escape(b []byte) []byte {
	var buf bytes.Buffer
	i := bytes.IndexAny(b, escapedChars)
	for i != -1 {
		buf.Write(b[:i])
		var esc []byte
		switch b[i] {
		case '<':
			esc = []byte("&lt;")
		case '>':
			esc = []byte("&gt;")
		}
		b = b[i+1:]
		buf.Write(esc)
		i = bytes.IndexAny(b, escapedChars)
	}
	buf.Write(b)
	return buf.Bytes()
}
