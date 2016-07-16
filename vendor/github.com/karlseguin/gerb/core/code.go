package core

type Executable interface {
	Execute(context *Context) ExecutionState
}

type Code interface {
	Executable
	IsCodeContainer() bool
	IsContentContainer() bool
	IsSibling() bool
	AddExecutable(Executable)
	AddCode(Code) error
}

type Codes struct {
	Trim bool
	List []Code
}
