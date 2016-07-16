package core

type Literal struct {
	data []byte
}

func (l *Literal) Execute(context *Context) ExecutionState {
	context.Writer.Write(l.data)
	return NormalState
}
