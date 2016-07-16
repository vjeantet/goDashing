package core

type ExecutionState int
type TagType int

const (
	NormalState ExecutionState = iota
	BreakState
	ContinueState

	OutputTag TagType = iota
	UnsafeTag
	CodeTag
	CommentTag
	NoTag
)
