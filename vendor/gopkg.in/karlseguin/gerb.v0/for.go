package gerb

import (
	"errors"
	"fmt"
	"github.com/karlseguin/gerb/core"
	"reflect"
)

func ForFactory(p *core.Parser) (core.Code, error) {
	if p.TagContains(';') {
		return ExplicitForFactory(p)
	}
	if p.SkipSpaces() == '{' {
		p.Next()
		return &ForCode{NormalContainer: new(core.NormalContainer)}, nil
	}
	return RangedForFactory(p)
}

func ExplicitForFactory(p *core.Parser) (core.Code, error) {
	code := &ForCode{NormalContainer: new(core.NormalContainer)}
	if p.SkipSpaces() != ';' {
		assignment, err := p.ReadAssignment()
		if err != nil {
			return nil, err
		}
		code.init = assignment
	}

	if p.SkipSpaces() != ';' {
		return nil, p.Error("Invalid for loop, expecting INIT; CONDITION; STEP (1)")
	}
	p.Next()

	verifiable, err := p.ReadConditionGroup(false)
	if err != nil {
		return nil, err
	}

	code.verifiable = verifiable

	if p.SkipSpaces() != ';' {
		return nil, p.Error("Invalid for loop, expecting INIT; CONDITION; STEP (1)")
	}
	p.Next()

	if p.SkipSpaces() != '{' {
		value, err := p.ReadAssignment()
		if err != nil {
			return nil, err
		}
		code.step = value
	}
	if p.SkipSpaces() != '{' {
		return nil, p.Error("Missing openening brace for for statement")
	}
	p.Next()
	return code, nil
}

func RangedForFactory(p *core.Parser) (core.Code, error) {
	code := &RangedForCode{NormalContainer: new(core.NormalContainer)}
	tokens, err := p.ReadTokenList()
	if err != nil {
		return nil, err
	}
	if len(tokens) != 2 {
		return nil, p.Error("Invalid for loop, ranged loop should have two variables")
	}
	code.tokens = tokens
	c := p.SkipSpaces()
	if c == ':' {
		c = p.Next()
	}
	if c != '=' {
		return nil, p.Error("Invalid for loop, ranged loop expecting assignment operator")
	}
	p.Next()

	if p.SkipSpaces() != 'r' || p.ConsumeIf([]byte("range")) == false {
		return nil, p.Error("invalid for loop, expected 'range' keyword")
	}

	value, err := p.ReadValue()
	if err != nil {
		return nil, err
	}
	code.value = value
	if p.SkipSpaces() != '{' {
		return nil, p.Error("Missing openening brace for if statement")
	}
	p.Next()

	return code, nil

}

type ForCode struct {
	*core.NormalContainer
	init       *core.Assignment
	verifiable core.Verifiable
	step       *core.Assignment
}

func (c *ForCode) Execute(context *core.Context) core.ExecutionState {
	if c.init != nil {
		c.init.Execute(context)
	}
	for {
		if c.verifiable != nil && c.verifiable.IsTrue(context) == false {
			break
		}
		state := c.NormalContainer.Execute(context)
		if state == core.BreakState {
			break
		}
		if c.step != nil {
			c.step.Execute(context)
		}
	}
	return core.NormalState
}

func (c *ForCode) IsCodeContainer() bool {
	return true
}

func (c *ForCode) IsContentContainer() bool {
	return true
}

func (c *ForCode) IsSibling() bool {
	return false
}

func (c *ForCode) AddCode(code core.Code) error {
	return errors.New(fmt.Sprintf("%v is not a valid tag as a descendant of a for loop", code))
}

type RangedForCode struct {
	*core.NormalContainer
	tokens []string
	value  core.Value
}

func (c *RangedForCode) Execute(context *core.Context) core.ExecutionState {
	value := reflect.ValueOf(c.value.Resolve(context))
	kind := value.Kind()

	if kind == reflect.Array || kind == reflect.Slice || kind == reflect.String || kind == reflect.Map {
		length := value.Len()
		if length == 0 {
			return core.NormalState
		}
		if kind == reflect.Map {
			c.executeMap(context, value, length)
		} else {
			c.executeArray(context, value, length)
		}
	}
	delete(context.Data, c.tokens[0])
	delete(context.Data, c.tokens[1])
	return core.NormalState
}

func (c *RangedForCode) IsCodeContainer() bool {
	return true
}

func (c *RangedForCode) IsContentContainer() bool {
	return true
}

func (c *RangedForCode) IsSibling() bool {
	return false
}

func (c *RangedForCode) AddCode(code core.Code) error {
	return errors.New(fmt.Sprintf("%v is not a valid tag as a descendant of a for loop", code))
}

func (c *RangedForCode) executeArray(context *core.Context, value reflect.Value, length int) {
	for i := 0; i < length; i++ {
		context.Data[c.tokens[0]] = i
		context.Data[c.tokens[1]] = value.Index(i).Interface()
		c.NormalContainer.Execute(context)
	}
}

func (c *RangedForCode) executeMap(context *core.Context, value reflect.Value, length int) {
	keys := value.MapKeys()
	for i := 0; i < length; i++ {
		key := keys[i]
		context.Data[c.tokens[0]] = key
		context.Data[c.tokens[1]] = value.MapIndex(key).Interface()
		c.NormalContainer.Execute(context)
	}
}
