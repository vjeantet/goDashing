package r

import (
	"fmt"
	"strconv"
)

// Convert arbitrary data to []byte
func ToBytes(data interface{}) []byte {
	if data == nil {
		return []byte{}
	}
	switch typed := data.(type) {
	case byte:
		return []byte{typed}
	case []byte:
		return typed
	case string:
		return []byte(typed)
	case bool:
		return []byte(strconv.FormatBool(typed))
	case float64:
		return []byte(strconv.FormatFloat(typed, 'g', -1, 64))
	case uint64:
		return []byte(strconv.FormatUint(typed, 10))
	case uint:
		return []byte(strconv.FormatUint(uint64(typed), 10))
	case int:
		return []byte(strconv.Itoa(typed))
	case fmt.Stringer:
		return []byte(typed.String())
	}
	return []byte(fmt.Sprintf("%v", data))
}

// Convert arbitrary data to string
func ToString(data interface{}) string {
	switch typed := data.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return string(ToBytes(data))
	}
}

// Convert arbitrary data to string
func ToInt(data interface{}) (int, bool) {
	switch typed := data.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case uint:
		return int(typed), true
	case byte:
		return int(typed), true
	default:
		return 0, false
	}
}

// Convert arbitrary data to string
func ToFloat(data interface{}) (float64, bool) {
	switch typed := data.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	default:
		if n, ok := ToInt(data); ok {
			return float64(n), true
		}
		return 0, false
	}
}
