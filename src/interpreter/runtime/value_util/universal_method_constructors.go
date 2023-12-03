package value_util

import (
	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
	"project_umbrella/interpreter/runtime/value_types/library"
)

func newEqualsMethod(value_ value.Value) *function.Function {
	return function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			built_in_declarations.UniversalEqualsMethod.Name,
			nil,
		),

		func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
			return builtInEquals(runtime_, value_, arguments[0])
		},

		built_in_declarations.UniversalEqualsMethod.Type,
	)
}

func newNotEqualsMethod(value_ value.Value) *function.Function {
	return function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			built_in_declarations.UniversalNotEqualsMethod.Name,
			nil,
		),

		func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
			return !builtInEquals(runtime_, value_, arguments[0])
		},

		built_in_declarations.UniversalNotEqualsMethod.Type,
	)
}

func newToStringMethod(value_ value.Value) *function.Function {
	return function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			built_in_declarations.UniversalToStringMethod.Name,
		),

		func(runtime_ *runtime.Runtime, _ ...value.Value) value.Value {
			var result string

			switch value_ := value_.(type) {
			case value_types.BooleanValue:
				result = booleanToString(value_)

			case value_types.FloatValue:
				result = floatToString(value_)

			case value_types.IntegerValue:
				result = integerToString(value_)

			case *function.Function:
				result = functionToString(value_)

			case *library.Library:
				result = "(library)"

			case value_types.StringValue:
				result = stringToString(value_)

			case value_types.TupleValue:
				result = tupleToString(runtime_, value_)

			case value_types.UnitValue:
				result = "(unit)"
			}

			return value_types.StringValue(result)
		},

		built_in_declarations.UniversalToStringMethod.Type,
	)
}
