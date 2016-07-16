package r

import (
	"reflect"
	"strings"
	"sync"
)

var (
	methodCache = make(map[reflect.Type]map[string]reflect.Value)
	methodLock  sync.RWMutex
)

func Method(value reflect.Value, name string) reflect.Value {
	t := value.Type()
	methodLock.RLock()
	data, exists := methodCache[t]
	methodLock.RUnlock()

	if exists == false {
		data = buildMethodData(t)
	}
	return data[name]
}

func buildMethodData(t reflect.Type) map[string]reflect.Value {
	data := make(map[string]reflect.Value)
	for i, l := 0, t.NumMethod(); i < l; i++ {
		m := t.Method(i)
		data[strings.ToLower(m.Name)] = m.Func
	}
	return data
}
