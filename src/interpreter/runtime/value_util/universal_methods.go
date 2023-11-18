package value_util

import (
	"fmt"
	"strings"

	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
)

func booleanToString(value_ value_types.BooleanValue) string {
	return fmt.Sprintf("%t", value_)
}

func builtInEquals(
	runtime_ *runtime.Runtime,
	value1 value.Value,
	value2 value.Value,
) value_types.BooleanValue {
	tuple1, ok1 := value1.(value_types.TupleValue)
	tuple2, ok2 := value2.(value_types.TupleValue)

	if ok1 && ok2 {
		if len(tuple1.Elements) != len(tuple2.Elements) {
			return false
		}

		for i, element := range tuple1.Elements {
			if !CallEqualsMethod(runtime_, element, tuple2.Elements[i]) {
				return false
			}
		}

		return true
	}

	return value1 == value2
}

func floatToString(value_ value_types.FloatValue) string {
	return fmt.Sprintf("%g", value_)
}

func integerToString(value_ value_types.IntegerValue) string {
	return fmt.Sprintf("%d", value_)
}

func functionToString(value_ *function.Function) string {
	return value_.Name
}

func stringToString(value_ value_types.StringValue) string {
	return value_.Content
}

func tupleToString(runtime_ *runtime.Runtime, value_ value_types.TupleValue) string {
	var insideParentheses string

	switch len(value_.Elements) {
	case 0:
		insideParentheses = ","

	case 1:
		insideParentheses = fmt.Sprintf(
			"%s,",
			CallToStringMethod(runtime_, value_.Elements[0]).Content,
		)

	default:
		elementsAsStrings := make([]string, 0, len(value_.Elements))

		for _, element := range value_.Elements {
			elementsAsStrings = append(
				elementsAsStrings,
				CallToStringMethod(runtime_, element).Content,
			)
		}

		insideParentheses = strings.Join(elementsAsStrings, ", ")
	}

	return fmt.Sprintf("(%s)", insideParentheses)
}

func unitToString(value_ value_types.UnitValue) string {
	return "(unit)"
}
