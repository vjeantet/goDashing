package core

import (
	"fmt"
)

type Assignment struct {
	names      []string
	values     []Value
	definition bool
}

func (c *Assignment) Execute(context *Context) ExecutionState {
	index := 0
	hasNew := false
	remaining := len(c.names)
	for _, value := range c.values {
		values := value.ResolveAll(context)
		valueCount := len(values)
		if valueCount > remaining {
			Log.Error(fmt.Sprintf("%d more return value than there are variables", valueCount-remaining))
			valueCount = remaining
		}
		for i := 0; i < valueCount; i++ {
			name := c.names[index]
			index++
			remaining--
			if _, exists := context.Data[name]; !exists {
				hasNew = true
			}
			context.Data[name] = values[i]
		}
	}

	if remaining > 0 {
		Log.Error(fmt.Sprintf("Expected %d variable(s), got %d", len(c.names), len(c.names)-remaining))
	}

	if hasNew && !c.definition {
		Log.Error(fmt.Sprintf("Assigning to %v, which are undefined, using =", c.names))
	} else if !hasNew && c.definition {
		Log.Error(fmt.Sprintf("Assigning to %v, which are already defined, using :=", c.names))
	}
	return NormalState
}

func (c *Assignment) Rollback(context *Context) {
	if c.definition {
		for _, name := range c.names {
			delete(context.Data, name)
		}
	}

}

func (c *Assignment) IsCodeContainer() bool {
	return false
}

func (c *Assignment) IsContentContainer() bool {
	return false
}

func (c *Assignment) IsSibling() bool {
	return false
}

func (c *Assignment) AddExecutable(Executable) {
	panic("AddExecutable called on assignment tag")
}

func (c *Assignment) AddCode(Code) error {
	panic("AddCode called on EndScope tag")
}
