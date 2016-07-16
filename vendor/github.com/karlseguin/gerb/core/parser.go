package core

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

type Parser struct {
	end      int
	len      int
	position int
	data     []byte
}

func NewParser(data []byte) *Parser {
	p := &Parser{
		data: data,
		len:  len(data),
		end:  len(data) - 1,
	}
	return p
}

var (
	trueBytes    = []byte("true")
	falseBytes   = []byte("false")
	nilBytes     = []byte("nil")
	yieldBytes   = []byte("yield")
	closeTag     = []byte("%>")
	closeTrimTag = []byte("%%>")
)

func (p *Parser) ReadLiteral(trim bool) *Literal {
	start := p.position
	for {
		if trim {
			if c := p.data[p.position]; c == '\n' || c == '\n' {
				start++
				p.position++
			}
			trim = false
		}
		if p.SkipUntil('%') == false {
			return &Literal{clone(p.data[start:p.len])}
		}
		if p.Prev() == '<' {
			p.position++ //move past the %
			to := p.position - 2
			if p.data[p.position] == '%' { //trim head
				p.position++ //move past the 2nd %
				to--
				if to >= 0 {
					for c := p.data[to]; c == '\n' || c == '\r'; c = p.data[to] {
						to--
						if to < start {
							break
						}
					}
					to++
				}
			}
			if to < 0 {
				to = 0
			}
			return &Literal{clone(p.data[start:to])}
		}
	}
}

func (p *Parser) ReadValue() (Value, error) {
	first := p.SkipSpaces()
	negate := false
	invert := false
	if first == '-' {
		negate = true
		p.position++
		first = p.SkipSpaces()
	} else if first == '!' {
		invert = true
		p.position++
		first = p.SkipSpaces()
	}
	var value Value
	var err error
	var ok bool
	if first == 0 {
		return nil, p.Error("Expected value, got nothing")
	}
	if first >= '0' && first <= '9' {
		value, err = p.ReadNumber(negate, invert)
	} else if first == '\'' {
		value, err = p.ReadChar(negate, invert)
	} else if first == '"' {
		value, err = p.ReadString(negate, invert, '"', true)
	} else if first == '`' {
		value, err = p.ReadString(negate, invert, '`', false)
	} else {
		if value, ok = p.ReadBuiltin(invert); ok == false {
			value, err = p.ReadDynamic(negate, invert)
		}
	}
	if err != nil {
		return nil, err
	}
	c1 := p.SkipSpaces()
	c2 := p.data[p.position+1]
	if c1 == '%' && c2 == '>' {
		return value, nil
	}
	s1 := string(c1)
	s3 := s1 + string(c2)
	unaryFactory, ok := UnaryOperations[s3]
	if ok {
		p.position += 2
		return unaryFactory(value), nil
	}

	factory, ok := Operations[s3]
	if ok == false {
		if factory, ok = Operations[s1]; ok == false {
			return value, nil
		}
		p.position++
	} else {
		p.position += 2
	}
	right, err := p.ReadValue()
	if err != nil {
		return nil, err
	}
	return factory(value, right), nil
}

func (p *Parser) ReadNumber(negate, invert bool) (Value, error) {
	if invert {
		return nil, p.Error("Don't know what to do with a ! number")
	}
	integer := 0
	fraction := 0
	target := &integer
	partLength := 0
	isDecimal := false
	for ; p.position < p.end; p.position++ {
		c := p.data[p.position]
		if c == '.' {
			if isDecimal {
				break
			}
			target = &fraction
			partLength = 0
			isDecimal = true
			continue
		}
		if c < '0' || c > '9' {
			break
		}
		partLength++
		*target = *target*10 + int(c-'0')
	}

	if isDecimal {
		value := float64(integer) + float64(fraction)/math.Pow10(partLength)
		if negate {
			value *= -1
		}
		return &StaticValue{value}, nil
	}
	if negate {
		integer *= -1
	}
	return &StaticValue{integer}, nil
}

func (p *Parser) ReadChar(negate, invert bool) (Value, error) {
	if negate {
		return nil, p.Error("Don't know what to do with a negative character")
	}
	if invert {
		return nil, p.Error("Don't know what to do with a ! character")
	}
	c := p.Next()
	if c == '\\' {
		c = p.Next()
	}
	if p.Next() != '\'' {
		return nil, p.Error("Invalid character")
	}
	p.position++
	return &StaticValue{c}, nil
}

func (p *Parser) ReadString(negate, invert bool, end byte, allowEscape bool) (Value, error) {
	if negate {
		return nil, p.Error("Don't know what to do with a negative string")
	}
	if invert {
		return nil, p.Error("Don't know what to do with a ! string")
	}
	p.position++
	start := p.position
	escaped := 0

	for ; p.position < p.end; p.position++ {
		c := p.data[p.position]
		if c == '\\' && allowEscape {
			escaped++
			p.position++
			continue
		}
		if c == end {
			break
		}
	}

	var data []byte
	var err error
	if escaped > 0 {
		data, err = p.unescape(p.data[start:p.position], escaped)
		if err != nil {
			return nil, err
		}
	} else {
		data = p.data[start:p.position]
	}
	p.position++ //consume the "
	return &StaticValue{string(data)}, nil
}

func (p *Parser) ReadBuiltin(invert bool) (Value, bool) {
	if p.ConsumeIf(trueBytes) {
		if invert {
			return falseValue, true
		}
		return trueValue, true
	}
	if p.ConsumeIf(falseBytes) {
		if invert {
			return trueValue, true
		}
		return falseValue, true
	}
	if p.ConsumeIf(nilBytes) {
		return nilValue, true
	}
	at := p.position
	if p.ConsumeIf(yieldBytes) {
		if p.SkipSpaces() == '%' {
			return new(DefaultYieldValue), true
		} else {
			//roll this back, it must be a named yield
			p.position = at
		}
	}
	return nil, false
}

func (p *Parser) ReadDynamic(negate, invert bool) (Value, error) {
	start := p.position
	fields := make([]string, 0, 5)
	types := make([]DynamicFieldType, 0, 5)
	args := make([][]Value, 0, 5)
	for p.position < p.end {
		c := p.data[p.position]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (p.position != start && c >= '0' && c <= '9') || c == '_' {
			p.position++
			continue
		}
		field := string(bytes.ToLower(p.data[start:p.position]))
		var t DynamicFieldType
		var a []Value
		var err error
		isEnd := c != '.' && c != '(' && c != '['
		if c == '.' {
			t = FieldType
			a = nil
			p.position++
		} else if isEnd {
			t = FieldType
			a = nil
		} else if c == '[' {
			t = IndexedType
			p.position++
			a, err = p.ReadIndexing()
		} else if c == '(' {
			t = MethodType
			p.position++
			a, err = p.ReadArgs()
		}
		if err != nil {
			return nil, err
		}

		if isEnd && (p.data[start] == ' ' || start == p.position) {
			break
		}
		if len(field) != 0 || t == IndexedType {
			fields = append(fields, field)
			types = append(types, t)
			args = append(args, a)
		}
		if isEnd {
			break
		}
		start = p.position
	}
	id := ""
	if len(types) == 1 && types[0] == FieldType {
		id = fields[0]
	}
	return &DynamicValue{id, fields, types, args, negate, invert}, nil
}

func (p *Parser) ReadIndexing() ([]Value, error) {
	implicitStart := false
	if p.SkipSpaces() == ':' {
		implicitStart = true
		p.position++
	}
	first, err := p.ReadValue()
	if err != nil {
		return nil, err
	}
	if implicitStart {
		p.position++
		return []Value{&StaticValue{0}, first}, nil
	}

	c := p.SkipSpaces()
	if c == ']' {
		p.position++
		return []Value{first}, nil
	}
	if c != ':' {
		return nil, p.Error("Unrecognized array/map index")
	}

	p.position++
	if p.SkipSpaces() == ']' {
		p.position++
		return []Value{first, nilValue}, nil
	}
	second, err := p.ReadValue()
	if err != nil {
		return nil, err
	}

	if c = p.SkipSpaces(); c != ']' {
		return nil, p.Error("Expected closing array/map bracket")
	}
	p.position++
	return []Value{first, second}, nil
}

func (p *Parser) ReadArgs() ([]Value, error) {
	if p.data[p.position] == ')' {
		p.position++
		return nil, nil
	}

	values := make([]Value, 0, 3)
	for {
		value, err := p.ReadValue()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		c := p.SkipSpaces()
		if c == ')' {
			p.position++
			break
		}
		if c != ',' {
			return nil, p.Error("Invalid argument list given to function")
		}
		p.position++
	}
	return values, nil
}

func (p *Parser) ReadToken() (string, error) {
	if c := p.SkipSpaces(); c == 0 {
		return "", p.Error("Expect a valid code token")
	}
	start := p.position
	for ; p.position < p.end; p.position++ {
		c := p.data[p.position]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (p.position != start && (c < '0' || c > '9')) && c != '_' {
			break
		}
	}
	return string(p.data[start:p.position]), nil
}

func (p *Parser) ReadTokenList() ([]string, error) {
	names := make([]string, 0, 3)
	for {
		name, err := p.ReadToken()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		if c := p.SkipSpaces(); c != ',' {
			break
		}
		p.Next()
	}
	return names, nil
}

func (p *Parser) ReadAssignment() (*Assignment, error) {
	a := &Assignment{definition: false}
	names, err := p.ReadTokenList()
	if err != nil {
		return nil, err
	}
	a.names = names
	c := p.SkipSpaces()

	if c == ':' {
		a.definition = true
		c = p.Next()
	} else if len(names) == 1 {
		c1 := p.data[p.position+1]
		if (c == '+' || c == '-') && (c1 == '=' || c1 == c) {
			p.position += 2
			if err := p.operationalAssignment(c, c1, a); err != nil {
				return nil, err
			}
			return a, nil
		}
	}

	if c != '=' {
		return nil, p.Error("Invalid assignment, expecting '=' or ':=' ")
	}

	p.Next()
	values, err := p.ReadValueList()
	if err != nil {
		return nil, err
	}
	a.values = values
	return a, nil
}

var singleDynamicField = []DynamicFieldType{FieldType}

func (p *Parser) operationalAssignment(c1, c2 byte, a *Assignment) error {
	left := &DynamicValue{
		id:    a.names[0],
		names: a.names,
		types: singleDynamicField,
	}

	var right, value Value
	if c2 == '=' {
		var err error
		if right, err = p.ReadValue(); err != nil {
			return err
		}
	}

	if c1 == '+' {
		if c2 == '+' {
			value = IncrementOperation(left)
		} else {
			value = PlusEqualOperation(left, right)
		}
	} else {
		if c2 == '-' {
			value = DecrementOperation(left)
		} else {
			value = MinusEqualOperation(left, right)
		}
	}
	a.values = []Value{value}
	return nil
}

func (p *Parser) ReadValueList() ([]Value, error) {
	values := make([]Value, 0, 3)
	for {
		value, err := p.ReadValue()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		if c := p.SkipSpaces(); c != ',' {
			break
		}
		p.Next()
	}
	return values, nil
}

func (p *Parser) ReadConditionGroup(parentheses bool) (Verifiable, error) {
	group := &ConditionGroup{make([]Verifiable, 0, 2), make([]LogicalOperator, 0, 1)}
	for {
		if p.SkipSpaces() == '(' {
			p.position++
			verifiable, err := p.ReadConditionGroup(true)
			if err != nil {
				return nil, err
			}
			group.verifiables = append(group.verifiables, verifiable)
		} else {
			left, err := p.ReadValue()
			if err != nil {
				return nil, err
			}
			if left == nil {
				return nil, p.Error("Invalid of missing left value in condition")
			}

			var booleanCondition *BooleanCondition
			operator := p.ReadComparisonOperator()
			if operator == UnknownComparisonOperator {
				booleanCondition = &BooleanCondition{left}
			}

			if booleanCondition != nil {
				group.verifiables = append(group.verifiables, booleanCondition)
			} else {
				right, err := p.ReadValue()
				if err != nil {
					return nil, err
				}
				if right == nil {
					return nil, p.Error("Invalid of missing right value in condition")
				}
				group.verifiables = append(group.verifiables, &Condition{left, operator, right})
			}
		}

		c := p.SkipSpaces()
		if parentheses && c == ')' {
			p.position++
			c = p.SkipSpaces()
		}
		if c == '%' {
			break
		}

		logical := p.ReadLogicalOperator()
		if logical == UnknownLogicalOperator {
			break
		}
		group.joins = append(group.joins, logical)
	}
	return group, nil
}

func (p *Parser) ReadComparisonOperator() ComparisonOperator {
	c1 := p.SkipSpaces()
	c2 := p.data[p.position+1]

	if c2 == '=' {
		switch c1 {
		case '=':
			p.position += 2
			return Equals
		case '!':
			p.position += 2
			return NotEquals
		case '>':
			p.position += 2
			return GreaterThanOrEqual
		case '<':
			p.position += 2
			return LessThanOrEqual
		default:
			return UnknownComparisonOperator
		}
	}
	if c1 == '>' {
		p.position++
		return GreaterThan
	}
	if c1 == '<' {
		p.position++
		return LessThan
	}

	return UnknownComparisonOperator
}

func (p *Parser) ReadLogicalOperator() LogicalOperator {
	c1 := p.SkipSpaces()
	c2 := p.data[p.position+1]
	if c1 == '&' && c2 == '&' {
		p.position += 2
		return AND
	}
	if c1 == '|' && c2 == '|' {
		p.position += 2
		return OR
	}
	return UnknownLogicalOperator
}

func (p *Parser) ReadTagType() TagType {
	switch p.Consume() {
	case 0:
		return NoTag
	case '=':
		return OutputTag
	case '!':
		return UnsafeTag
	case '#':
		return CommentTag
	default:
		p.position--
		return CodeTag
	}
}

func (p *Parser) ReadComment() (trim bool, err error) {
	for {
		p.SkipUntil('%')
		if p.ConsumeIf(closeTrimTag) {
			return true, nil
		}
		if p.ConsumeIf(closeTag) {
			return false, nil
		}
		if p.Next() == 0 {
			return false, p.Error("Expected closing tag for comment")
		}
	}

	return false, nil
}

func (p *Parser) ReadCloseTag() error {
	if p.SkipSpaces() != '%' || p.Next() != '>' {
		return p.Error("Expected closing tag")
	}
	p.position++
	return nil
}

func (p *Parser) SkipUntil(b byte) bool {
	if at := bytes.IndexByte(p.data[p.position:], b); at != -1 {
		p.position = p.position + at
		return true
	}
	p.position = len(p.data)
	return false
}

func (p *Parser) SkipSpaces() byte {
	for ; p.position < p.end; p.position++ {
		c := p.data[p.position]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return c
		}
	}
	return 0
}

func (p *Parser) Consume() byte {
	if p.position > p.end {
		return 0
	}
	c := p.data[p.position]
	p.position++
	return c
}

func (p *Parser) Next() byte {
	p.position++
	if p.position > p.end {
		return 0
	}
	return p.data[p.position]
}

func (p *Parser) Peek() byte {
	position := p.position + 1
	if position > p.end {
		return 0
	}
	return p.data[position]
}

func (p *Parser) Prev() byte {
	return p.data[p.position-1]
}

func (p *Parser) Backwards(length int) {
	p.position -= length
}

func (p *Parser) ConsumeIf(bytes []byte) bool {
	length := len(bytes)
	position := p.position
	left := p.len - position
	if left < length {
		return false
	}

	for index, b := range bytes {
		if p.data[position+index] != b {
			return false
		}
	}
	p.position += length
	return true
}

func (p *Parser) TagContains(b byte) bool {
	p.SkipSpaces()
	position := p.position
	var terminator byte = 0
	for ; position < p.end; position++ {
		c := p.data[position]
		if terminator == 0 {
			if c == b {
				return true
			}
			if c == '"' {
				terminator = '"'
			} else if c == '\'' {
				terminator = '\''
			} else if (c == '>' && p.data[position-1] == '%') || c == '{' {
				break
			}
		} else if c == terminator && p.data[position-1] != '\\' {
			terminator = 0
		}
	}
	return false
}

func (p *Parser) Dump(prefix string) {
	fmt.Println(prefix, ": ", string(p.data[p.position:]))
}

func (p *Parser) Error(s string) error {
	end := p.position
	for ; end < p.end; end++ {
		if p.data[end] == '%' && p.data[end+1] == '>' {
			break
		}
	}
	end += 2 //consume the > + this is exclusive
	if end > p.len {
		end = p.len
	}
	start := p.position
	if start > p.end {
		start = p.end
	}
	for ; start > 0; start-- {
		if p.data[start] == '%' && p.data[start-1] == '<' {
			start--
			break
		}
	}
	return errors.New(fmt.Sprintf("%s: %v", s, string(p.data[start:end])))
}

func (p *Parser) unescape(data []byte, escaped int) ([]byte, error) {
	value := make([]byte, len(data)-escaped)
	at := 0
	for {
		index := bytes.IndexByte(data, '\\')
		if index == -1 {
			copy(value[at:], data)
			break
		}
		at += copy(value[at:], data[:index])
		switch data[index+1] {
		case 'n':
			value[at] = '\n'
		case 'r':
			value[at] = '\r'
		case 't':
			value[at] = '\t'
		case '"':
			value[at] = '"'
		case '\\':
			value[at] = '\\'
		default:
			return nil, p.Error(fmt.Sprintf("Unknown escape sequence \\%s", string(data[index+1])))
		}
		at++
		data = data[index+2:]
	}
	return value, nil
}

func clone(data []byte) []byte {
	c := make([]byte, len(data))
	copy(c, data)
	return c
}
