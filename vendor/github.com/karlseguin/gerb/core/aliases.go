package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func init() {
	RegisterAliases("strings",
		"ToUpper", strings.ToUpper,
		"ToLower", strings.ToLower,
		"Contains", strings.Contains,
		"ContainsAny", strings.ContainsAny,
		"ContainsRune", strings.ContainsRune,
		"Count", strings.Count,
		"Fields", strings.Fields,
		"HasPrefix", strings.HasPrefix,
		"HasSuffix", strings.HasSuffix,
		"Index", strings.Index,
		"IndexAny", strings.IndexAny,
		"IndexByte", strings.IndexByte,
		"IndexRune", strings.IndexRune,
		"Join", strings.Join,
		"LastIndex", strings.LastIndex,
		"LastIndexAny", strings.LastIndexAny,
		"Repeat", strings.Repeat,
		"Replace", strings.Replace,
		"Split", strings.Split,
		"SplitAfter", strings.SplitAfter,
		"SplitAfterN", strings.SplitAfterN,
		"SplitN", strings.SplitN,
		"Title", strings.Title,
		"ToLower", strings.ToLower,
		"ToLowerSpecial", strings.ToLowerSpecial,
		"ToTitle", strings.ToTitle,
		"ToTitleSpecial", strings.ToTitleSpecial,
		"ToUpper", strings.ToUpper,
		"ToUpperSpecial", strings.ToUpperSpecial,
		"Trim", strings.Trim,
		"TrimLeft", strings.TrimLeft,
		"TrimPrefix", strings.TrimPrefix,
		"TrimRight", strings.TrimRight,
		"TrimSpace", strings.TrimSpace,
		"TrimSuffix", strings.TrimSuffix,
	)

	RegisterAlias("fmt",
		"Sprintf", fmt.Sprintf)

	RegisterAlias("strconv",
		"Atoi", strconv.Atoi)

	RegisterAliases("time",
		"Now", time.Now,
		"Unix", time.Unix,
		"Nanosecond", time.Nanosecond,
		"Microsecond", time.Microsecond,
		"Millisecond", time.Millisecond,
		"Second", time.Second,
		"Minute", time.Minute,
		"Hour", time.Hour,
		"Duration", (*time.Duration)(nil),
	)
}

var FunctionAliases = make(map[string]map[string]interface{})
var OtherAliases = make(map[string]map[string]interface{})

func RegisterAlias(packageName, memberName string, f interface{}) {
	packageName = strings.ToLower(packageName)
	memberName = strings.ToLower(memberName)
	value := reflect.ValueOf(f)
	if value.Kind() == reflect.Func {
		registerFunctionAlias(packageName, memberName, value)
	} else if value.Kind() == reflect.Ptr {
		registerFunctionAlias(packageName, memberName, reflect.TypeOf(f).Elem())
	} else {
		registerOtherAlias(packageName, memberName, f)
	}
}

func RegisterAliases(packageName string, data ...interface{}) {
	for i := 0; i < len(data); i += 2 {
		RegisterAlias(packageName, data[i].(string), data[i+1])
	}
}

func registerFunctionAlias(packageName, memberName string, value interface{}) {
	pkg, ok := FunctionAliases[packageName]
	if ok == false {
		pkg = make(map[string]interface{})
		FunctionAliases[packageName] = pkg
	}
	pkg[memberName] = value
}

func registerOtherAlias(packageName, memberName string, value interface{}) {
	pkg, ok := OtherAliases[packageName]
	if ok == false {
		pkg = make(map[string]interface{})
		OtherAliases[packageName] = pkg
	}
	pkg[memberName] = value
}
