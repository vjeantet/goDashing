package core

import (
	"github.com/karlseguin/gerb/r"
	"reflect"
	"strings"
)

var Builtins = make(map[string]interface{})

func RegisterBuiltin(name string, f interface{}) {
	Builtins[strings.ToLower(name)] = reflect.ValueOf(f)
}

func init() {
	RegisterBuiltin("len", LenBuiltin)
	RegisterBuiltin("int", IntBuiltin)
	RegisterBuiltin("yield", YieldBuiltin)
}

func LenBuiltin(value interface{}) int {
	n, _ := r.ToLength(value)
	return n
}

func IntBuiltin(value interface{}) interface{} {
	switch t := value.(type) {
	case float32:
		return int(t)
	case float64:
		return int(t)
	}
	return value
}

func YieldBuiltin(name string, context *Context) interface{} {
	if writer, ok := context.Contents[name]; ok {
		return writer.String()
	}
	return nil
}
