package core

import (
	"fmt"
	"github.com/karlseguin/gerb/r"
	"reflect"
)

var (
	trueValue  = &StaticValue{true}
	falseValue = &StaticValue{false}
	nilValue   = &StaticValue{nil}
)

type Value interface {
	Resolve(context *Context) interface{}
	ResolveAll(context *Context) []interface{}
	Id() string
}

type Coercable interface {
	ResolveCoerce(context *Context, to reflect.Type) reflect.Value
}

type StaticValue struct {
	value interface{}
}

func (v *StaticValue) Resolve(context *Context) interface{} {
	return v.value
}

func (v *StaticValue) ResolveCoerce(context *Context, to reflect.Type) reflect.Value {
	value := reflect.ValueOf(v.value)
	if reflect.TypeOf(v.value).ConvertibleTo(to) {
		return value.Convert(to)
	}
	return value
}

func (v *StaticValue) ResolveAll(context *Context) []interface{} {
	return []interface{}{v.value}
}

func (v *StaticValue) Id() string {
	return ""
}

type DynamicFieldType int

const (
	FieldType DynamicFieldType = iota
	MethodType
	IndexedType
)

type DynamicValue struct {
	id     string
	names  []string
	types  []DynamicFieldType
	args   [][]Value
	negate bool
	invert bool
}

func (v *DynamicValue) Resolve(context *Context) interface{} {
	value, _ := v.resolve(context, false)
	if v.negate {
		return applyNegate(value)
	}
	if v.invert {
		return applyInvert(value)
	}
	return value
}

func (v *DynamicValue) ResolveCorece(context *Context, to reflect.Type) interface{} {
	return v.Resolve(context)
}

func (v *DynamicValue) ResolveAll(context *Context) []interface{} {
	value, isArray := v.resolve(context, true)
	var values []interface{}
	if isArray {
		values = value.([]interface{})
	} else {
		values = []interface{}{value}
	}

	if len(values) > 1 && (v.negate || v.invert) {
		Log.Error("cannot apply ! or - to a multi-value return")
	} else if v.negate {
		values[0] = applyNegate(values[0])
	} else if v.invert {
		values[0] = applyInvert(values[0])
	}
	return values
}

func (v *DynamicValue) resolve(context *Context, all bool) (interface{}, bool) {
	var d interface{} = context.Data
	ok := true
	isRoot := true
	isAlias := false

	for i, l := 0, len(v.names); i < l; i++ {
		name := v.names[i]
		t := v.types[i]
		if t == FieldType {
			if d, ok = r.ResolveField(d, name); ok == false {
				if isRoot {
					if alias, ok := FunctionAliases[name]; ok {
						d = alias
						isAlias = true
						isRoot = false
						continue
					}
				} else if pkg, ok := OtherAliases[v.names[i-1]]; ok {
					if alias, ok := pkg[name]; ok {
						return alias, false
					}
				}
				return v.loggedNil(i), false
			}
		} else if t == IndexedType {
			if len(name) > 0 {
				if d, ok = r.ResolveField(d, name); ok == false {
					return v.loggedNil(i), false
				}
			}
			if d = unindex(d, v.args[i], context); d == nil {
				return v.loggedNil(i), false
			}
		} else if t == MethodType {
			if d = run(d, name, v.args[i], isRoot, isAlias, context); d == nil {
				return v.loggedNilMethod(i), false
			}
			if returns, ok := d.([]reflect.Value); ok {
				if all && i == l-1 {
					values := make([]interface{}, len(returns))
					for index, r := range returns {
						values[index] = r.Interface()
					}
					return values, true
				}
				d = returns[0].Interface()
			}
		}
		isAlias = false
		isRoot = false
	}
	return r.ResolveFinal(d), false
}

func (v *DynamicValue) Id() string {
	return v.id
}

func (v *DynamicValue) loggedNil(index int) interface{} {
	if index == 0 {
		Log.Error(fmt.Sprintf("%s is undefined", v.names[index]))
	} else {
		Log.Error(fmt.Sprintf("%s.%s is undefined", v.names[index-1], v.names[index]))
	}
	return nil
}

func (v *DynamicValue) loggedNilMethod(index int) interface{} {
	if index == 0 {
		Log.Error(fmt.Sprintf("%s is undefined", v.names[index]))
	} else {
		Log.Error(fmt.Sprintf("%s.%s is undefined or had undefined parameters", v.names[index-1], v.names[index]))
	}
	return nil
}

func unindex(container interface{}, params []Value, context *Context) interface{} {
	valueLength := len(params)
	if valueLength == 0 {
		return nil
	}

	value := reflect.ValueOf(container)
	kind := value.Kind()
	if kind == reflect.Array || kind == reflect.Slice || kind == reflect.String {
		length := value.Len()
		first, ok := r.ToInt(params[0].Resolve(context))
		if ok == false {
			return nil
		}
		if first < 0 {
			first = 0
		} else if first > length-1 {
			first = length
		}
		if valueLength == 2 {
			second, ok := r.ToInt(params[1].Resolve(context))
			if ok == false {
				second = length
			} else if second > length {
				second = length
			}
			return value.Slice(first, second).Interface()
		} else {
			return value.Index(first).Interface()
		}

	} else if kind == reflect.Map {
		indexValue := reflect.ValueOf(params[0].Resolve(context))
		return value.MapIndex(indexValue).Interface()
	}
	return nil
}

func run(container interface{}, name string, params []Value, isRoot, isAlias bool, context *Context) interface{} {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(r)
		}
	}()
	if isRoot {
		return runBuiltIn(name, params, context)
	}
	if isAlias {
		return runAlias(container.(map[string]interface{}), name, params, context)
	}

	c := reflect.ValueOf(container)
	m := r.Method(c, name)
	if m.IsValid() == false {
		return nil
	}
	v := make([]reflect.Value, len(params)+1)
	v[0] = c
	for index, param := range params {
		paramValue := reflect.ValueOf(param.Resolve(context))
		if paramValue.IsValid() == false {
			return nil
		}
		v[index+1] = paramValue
	}
	if returns := m.Call(v); len(returns) > 0 {
		return returns
	}
	return nil
}

func runBuiltIn(name string, params []Value, context *Context) interface{} {
	return runFromLookup(Builtins, name, params, context)
}

func runAlias(pkg map[string]interface{}, name string, params []Value, context *Context) interface{} {
	return runFromLookup(pkg, name, params, context)
}

func runFromLookup(lookup map[string]interface{}, name string, params []Value, context *Context) interface{} {
	m, ok := lookup[name]
	if ok == false {
		return nil
	}

	switch typed := m.(type) {
	case reflect.Value:
		t := typed.Type()
		c := t.NumIn()
		v := make([]reflect.Value, c)
		for index, param := range params {
			var value reflect.Value
			if cp, ok := param.(Coercable); ok {
				value = cp.ResolveCoerce(context, t.In(index))
			} else {
				value = reflect.ValueOf(param.Resolve(context))
			}
			v[index] = value
		}
		if c > len(params) {
			v[len(params)] = reflect.ValueOf(context)
		}
		if returns := typed.Call(v); len(returns) > 0 {
			return returns
		}
	case reflect.Type:
		if len(params) != 1 {
			Log.Error(fmt.Sprintf("Conversion to %s should have 1 parameter", name))
			return nil
		}
		v := reflect.ValueOf(params[0].Resolve(context))
		if v.Type().ConvertibleTo(typed) == false {
			Log.Error(fmt.Sprintf("Cannot convert %s to %s", v.Type().Name(), typed.Name()))
			return nil
		}
		return v.Convert(typed).Interface()
	}
	return nil
}

func applyNegate(v interface{}) interface{} {
	value := reflect.ValueOf(v)
	kind := value.Kind()
	if isInt(kind) {
		return -value.Int()
	}
	if isFloat(kind) {
		return -value.Float()
	}
	Log.Error(fmt.Sprintf("trying to negate a non-numeric value: %v", v))
	return v
}

func applyInvert(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return !b
	}
	Log.Error(fmt.Sprintf("trying to invert a non-boolean value: %v", v))
	return v
}

type DefaultYieldValue struct{}

func (v *DefaultYieldValue) Resolve(context *Context) interface{} {
	return YieldBuiltin("$", context)
}

func (v *DefaultYieldValue) ResolveAll(context *Context) []interface{} {
	return []interface{}{v.Resolve(context)}
}

func (v *DefaultYieldValue) Id() string {
	return ""
}
