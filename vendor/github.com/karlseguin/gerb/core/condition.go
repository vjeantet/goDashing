package core

import (
	"reflect"
	"time"
)

var (
	TrueCondition  = &BooleanCondition{&StaticValue{true}}
	FalseCondition = &BooleanCondition{&StaticValue{false}}
)

type LogicalOperator int
type ComparisonOperator int
type Type int

const (
	OR LogicalOperator = iota
	AND
	UnknownLogicalOperator

	Equals ComparisonOperator = iota
	NotEquals
	LessThan
	GreaterThan
	LessThanOrEqual
	GreaterThanOrEqual
	UnknownComparisonOperator

	String Type = iota
	Nil
	Int
	Int64
	Uint
	Float64
	Complex128
	Bool
	Time
	Today
	Array
	Unknown
)

var KindToType = map[reflect.Kind]Type{
	reflect.String:     String,
	reflect.Int:        Int,
	reflect.Int64:      Int64,
	reflect.Uint:       Uint,
	reflect.Float64:    Float64,
	reflect.Complex128: Complex128,
	reflect.Bool:       Bool,
	reflect.Array:      Array,
	reflect.Slice:      Array,
	reflect.Map:        Array,
}

var TypeOperations = map[Type]map[ComparisonOperator]ConditionResolver{
	String: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(string) == right.(string) },
		LessThan: func(left, right interface{}) bool { return left.(string) < right.(string) },
	},
	Nil: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left == nil && right == nil },
		LessThan: func(left, right interface{}) bool { return false },
	},
	Int: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(int) == right.(int) },
		LessThan: func(left, right interface{}) bool { return left.(int) < right.(int) },
	},
	Int64: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(int64) == right.(int64) },
		LessThan: func(left, right interface{}) bool { return left.(int64) < right.(int64) },
	},
	Uint: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(uint) == right.(uint) },
		LessThan: func(left, right interface{}) bool { return left.(uint) < right.(uint) },
	},
	Float64: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(float64) == right.(float64) },
		LessThan: func(left, right interface{}) bool { return left.(float64) < right.(float64) },
	},
	Complex128: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(complex128) == right.(complex128) },
		LessThan: func(left, right interface{}) bool { return false },
	},
	Bool: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(bool) == right.(bool) },
		LessThan: func(left, right interface{}) bool { return false },
	},
	Time: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return left.(time.Time).Unix() == right.(time.Time).Unix() },
		LessThan: func(left, right interface{}) bool { return left.(time.Time).Unix() < right.(time.Time).Unix() },
	},
	Today: map[ComparisonOperator]ConditionResolver{
		Equals: func(left, right interface{}) bool {
			l, r := left.(time.Time), right.(time.Time)
			return l.YearDay() == r.YearDay() && l.Year() == r.Year()
		},
		LessThan: func(left, right interface{}) bool {
			l, r := left.(time.Time), right.(time.Time)
			if l.Year() > r.Year() {
				return false
			}
			if l.Year() < r.Year() {
				return true
			}
			return l.YearDay() < r.YearDay()
		},
	},
	Array: map[ComparisonOperator]ConditionResolver{
		Equals:   func(left, right interface{}) bool { return reflect.DeepEqual(left, right) },
		LessThan: func(left, right interface{}) bool { return reflect.ValueOf(left).Len() < reflect.ValueOf(right).Len() },
	},
}

// Resolves a condition
type ConditionResolver func(left, right interface{}) bool

var ConditionLookup = map[ComparisonOperator]ConditionResolver{
	Equals:             EqualsComparison,
	NotEquals:          NotEqualsComparison,
	LessThan:           LessThanComparison,
	GreaterThan:        GreaterThanComparison,
	LessThanOrEqual:    LessThanOrEqualComparison,
	GreaterThanOrEqual: GreaterThanOrEqualComparison,
}

type Verifiable interface {
	IsTrue(context *Context) bool
}

// represents a group of conditions
type ConditionGroup struct {
	verifiables []Verifiable
	joins       []LogicalOperator
}

func (g *ConditionGroup) IsTrue(context *Context) bool {
	l := len(g.verifiables) - 1
	if l == 0 {
		return g.verifiables[0].IsTrue(context)
	}

	for i := 0; i <= l; i++ {
		if g.verifiables[i].IsTrue(context) {
			if i == l || g.joins[i] == OR {
				return true
			}
		} else if i != l && g.joins[i] == AND {
			for ; i < l; i++ {
				if g.joins[i] == OR {
					break
				}
			}
		}
	}
	return false
}

type BooleanCondition struct {
	value Value
}

func (c *BooleanCondition) IsTrue(context *Context) bool {
	value := c.value.Resolve(context)
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// represents a conditions (such as x == y)
type Condition struct {
	left     Value
	operator ComparisonOperator
	right    Value
}

func (c *Condition) IsTrue(context *Context) bool {
	left := c.left.Resolve(context)
	right := c.right.Resolve(context)
	return ConditionLookup[c.operator](left, right)
}

func EqualsComparison(left, right interface{}) bool {
	var t Type
	if left, right, t = convertToSameType(left, right); t == Unknown {
		return false
	}
	return TypeOperations[t][Equals](left, right)
}

func NotEqualsComparison(left, right interface{}) bool {
	return !EqualsComparison(left, right)
}

func LessThanComparison(left, right interface{}) bool {
	var t Type
	if left, right, t = convertToSameType(left, right); t == Unknown {
		return false
	}
	return TypeOperations[t][LessThan](left, right)
}

func LessThanOrEqualComparison(left, right interface{}) bool {
	var t Type
	if left, right, t = convertToSameType(left, right); t == Unknown {
		return false
	}
	return TypeOperations[t][Equals](left, right) || TypeOperations[t][LessThan](left, right)
}

func GreaterThanComparison(left, right interface{}) bool {
	var t Type
	if left, right, t = convertToSameType(left, right); t == Unknown {
		return false
	}
	return !TypeOperations[t][Equals](left, right) && !TypeOperations[t][LessThan](left, right)
}

func GreaterThanOrEqualComparison(left, right interface{}) bool {
	var t Type
	if left, right, t = convertToSameType(left, right); t == Unknown {
		return false
	}
	return !TypeOperations[t][LessThan](left, right)
}

func convertToSameType(left, right interface{}) (interface{}, interface{}, Type) {
	//rely on the above code to handle this properly
	if left == nil || right == nil {
		return left, right, Nil
	}

	if l, ok := left.(string); ok {
		if r, ok := right.(string); ok {
			return l, r, String
		}
		return left, right, Unknown
	}

	leftValue, rightValue := reflect.ValueOf(left), reflect.ValueOf(right)
	leftKind, rightKind := leftValue.Kind(), rightValue.Kind()
	if leftKind == rightKind {
		if t, ok := KindToType[leftKind]; ok {
			return left, right, t
		}
	}
	if left, right, t := convertNumbersToSameType(leftValue, leftKind, rightValue, rightKind); t != Unknown {
		return left, right, t
	}
	return left, right, Unknown
}

func convertNumbersToSameType(leftValue reflect.Value, leftKind reflect.Kind, rightValue reflect.Value, rightKind reflect.Kind) (interface{}, interface{}, Type) {
	if isInt(leftKind) {
		if isInt(rightKind) {
			return leftValue.Int(), rightValue.Int(), Int64
		}
		return nil, nil, Unknown
	}

	if isFloat(leftKind) {
		if isFloat(rightKind) {
			return leftValue.Float(), rightValue.Float(), Float64
		}
		return nil, nil, Unknown
	}

	if isComplex(leftKind) && isComplex(rightKind) {
		return leftValue.Complex(), rightValue.Complex(), Complex128
	}
	return nil, nil, Unknown
}

func isInt(kind reflect.Kind) bool {
	return kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64
}

func isFloat(kind reflect.Kind) bool {
	return kind == reflect.Float64 || kind == reflect.Float32
}

func isComplex(kind reflect.Kind) bool {
	return kind == reflect.Complex128 || kind == reflect.Complex64
}
