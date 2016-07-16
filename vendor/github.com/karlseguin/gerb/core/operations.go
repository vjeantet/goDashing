package core

import (
	"fmt"
	"github.com/karlseguin/gerb/r"
	"time"
)

var OneStaticValue = &StaticValue{1}

type OperationFactory func(a, b Value) Value
type UnaryOperationFactory func(a Value) Value

var UnaryOperations = map[string]UnaryOperationFactory{
	"++": IncrementOperation,
	"--": DecrementOperation,
}

var Operations = map[string]OperationFactory{
	"+=": PlusEqualOperation,
	"-=": MinusEqualOperation,
	"+":  AddOperation,
	"-":  SubOperation,
	"*":  MultiplyOperation,
	"/":  DivideOperation,
	"%":  ModuloOperation,
}

func IncrementOperation(a Value) Value {
	return &PlusEqualValue{a, OneStaticValue, "++", false}
}

func DecrementOperation(a Value) Value {
	return &PlusEqualValue{a, OneStaticValue, "--", true}
}

func PlusEqualOperation(a, b Value) Value {
	return &PlusEqualValue{a, b, "+=", false}
}

func MinusEqualOperation(a, b Value) Value {
	return &PlusEqualValue{a, b, "-=", true}
}

func AddOperation(a, b Value) Value {
	return &AdditiveValue{a, b, false}
}

func SubOperation(a, b Value) Value {
	return &AdditiveValue{a, b, true}
}

func MultiplyOperation(a, b Value) Value {
	return &MultiplicativeValue{a, b, false}
}

func DivideOperation(a, b Value) Value {
	return &MultiplicativeValue{a, b, true}
}

func ModuloOperation(a, b Value) Value {
	return &ModulatedValue{a, b}
}

type AdditiveValue struct {
	a      Value
	b      Value
	negate bool
}

func (v *AdditiveValue) Resolve(context *Context) interface{} {
	a := v.a.Resolve(context)
	b := v.b.Resolve(context)
	if na, ok := r.ToInt(a); ok {
		if nb, ok := r.ToInt(b); ok {
			if v.negate {
				nb = -nb
			}
			return na + nb
		}

	} else if fa, ok := r.ToFloat(a); ok {
		if fb, ok := r.ToFloat(b); ok {
			if v.negate {
				fb = -fb
			}
			return fa + fb
		}
	} else if ta, ok := a.(time.Duration); ok {
		if tb, ok := b.(time.Duration); ok {
			if v.negate {
				return ta - tb
			}
			return ta + tb
		}
	}
	if v.negate {
		return loggedOperationNil(a, b, "-", 0)
	}
	return loggedOperationNil(a, b, "+", 0)
}

func (v *AdditiveValue) ResolveAll(context *Context) []interface{} {
	return []interface{}{v.Resolve(context)}
}

func (v *AdditiveValue) Id() string {
	return ""
}

type MultiplicativeValue struct {
	a      Value
	b      Value
	divide bool
}

func (v *MultiplicativeValue) Resolve(context *Context) interface{} {
	a := v.a.Resolve(context)
	b := v.b.Resolve(context)
	if na, ok := r.ToInt(a); ok {
		if nb, ok := r.ToInt(b); ok {
			if v.divide {
				return na / nb
			}
			return na * nb
		}
	} else if fa, ok := r.ToFloat(a); ok {
		if fb, ok := r.ToFloat(b); ok {
			if v.divide {
				return fa / fb
			}
			return fa * fb
		}
	} else if ta, ok := a.(time.Duration); ok {
		if tb, ok := b.(time.Duration); ok {
			if v.divide {
				return ta / tb
			}
			return ta * tb
		}
	}
	if v.divide {
		return loggedOperationNil(a, b, "/", 0)
	}
	return loggedOperationNil(a, b, "*", 0)
}

func (v *MultiplicativeValue) ResolveAll(context *Context) []interface{} {
	return []interface{}{v.Resolve(context)}
}

func (v *MultiplicativeValue) Id() string {
	return ""
}

type ModulatedValue struct {
	a Value
	b Value
}

func (v *ModulatedValue) Resolve(context *Context) interface{} {
	a := v.a.Resolve(context)
	b := v.b.Resolve(context)
	if na, ok := r.ToInt(a); ok {
		if nb, ok := r.ToInt(b); ok {
			return na % nb
		}
	}
	return loggedOperationNil(a, b, "%", 0)
}

func (v *ModulatedValue) ResolveAll(context *Context) []interface{} {
	return []interface{}{v.Resolve(context)}
}

func (v *ModulatedValue) Id() string {
	return ""
}

func loggedOperationNil(a, b interface{}, sign string, r interface{}) interface{} {
	Log.Error(fmt.Sprintf("%v %s %v failed, invalid types", a, sign, b))
	return r
}

type PlusEqualValue struct {
	a         Value
	b         Value
	operation string
	negate    bool
}

func (v *PlusEqualValue) Resolve(context *Context) interface{} {
	id := v.a.Id()
	if len(id) == 0 {
		Log.Error(fmt.Sprintf("Invalid operation %s on a non-dynamic value ", v.operation))
	}
	counter, ok := context.Counters[id]
	if ok == false {
		counter, ok = r.ToInt(v.a.Resolve(context))
		if ok == false {
			Log.Error(fmt.Sprintf("Called %s on a non-integer %s", v.operation, id))
		}
	}
	b, ok := r.ToInt(v.b.Resolve(context))
	if ok == false {
		Log.Error(fmt.Sprintf("Trying to call %s on %s but the right value was not an integer", v.operation, id))
	}
	if v.negate {
		counter -= b
	} else {
		counter += b
	}
	context.Counters[id] = counter
	return counter
}

func (v *PlusEqualValue) ResolveAll(context *Context) []interface{} {
	return []interface{}{v.Resolve(context)}
}

func (v *PlusEqualValue) Id() string {
	return ""
}
