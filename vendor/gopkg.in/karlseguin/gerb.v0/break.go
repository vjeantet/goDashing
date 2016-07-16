package gerb

import (
	"github.com/karlseguin/gerb/core"
)

var (
	breakCode    = &ControlFlowCode{core.BreakState}
	continueCode = &ControlFlowCode{core.ContinueState}
)

func BreakFactory(p *core.Parser) (core.Code, error) {
	return breakCode, nil
}

func ContinueFactory(p *core.Parser) (core.Code, error) {
	return continueCode, nil
}

type ControlFlowCode struct {
	state core.ExecutionState
}

func (c *ControlFlowCode) Execute(context *core.Context) core.ExecutionState {
	return c.state
}

func (c *ControlFlowCode) IsCodeContainer() bool {
	return false
}

func (c *ControlFlowCode) IsContentContainer() bool {
	return false
}

func (c *ControlFlowCode) IsSibling() bool {
	return false
}

func (c *ControlFlowCode) AddExecutable(core.Executable) {
	panic("AddExecutable called on control flow tag")
}

func (c *ControlFlowCode) AddCode(code core.Code) error {
	panic("AddCode called on control flow tag")
}
