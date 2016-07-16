package gerb

import (
	"errors"
	"github.com/karlseguin/gerb/core"
)

func ContentFactory(p *core.Parser) (core.Code, error) {
	value, err := p.ReadValue()
	if err != nil {
		return nil, err
	}
	if p.SkipSpaces() != '{' {
		return nil, p.Error("Missing openening brace for content statement")
	}
	p.Next()
	return &ContentCode{new(core.NormalContainer), value}, nil
}

type ContentCode struct {
	*core.NormalContainer
	value core.Value
}

func (c *ContentCode) Execute(context *core.Context) core.ExecutionState {
	name, ok := c.value.Resolve(context).(string)
	if ok == false {
		core.Log.Error("Content tag expects a string variable")
		return core.NormalState
	}
	writer, ok := context.Contents[name]
	if ok == false {
		writer = core.BytePool.Checkout()
		context.Contents[name] = writer
	}
	prevWriter := context.Writer
	context.Writer = writer
	state := c.NormalContainer.Execute(context)
	context.Writer = prevWriter
	return state
}

func (c *ContentCode) IsCodeContainer() bool {
	return true
}

func (c *ContentCode) IsContentContainer() bool {
	return true
}

func (c *ContentCode) IsSibling() bool {
	return false
}

func (c *ContentCode) AddCode(core.Code) error {
	return errors.New("Failed to parse template, you might have an extra } (content tag related)")
}
