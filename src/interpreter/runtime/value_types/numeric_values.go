package value_types

import (
	"math"
	"reflect"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type FloatValue float64

func (value_ FloatValue) Definition() *value.ValueDefinition {
	result := newNumberDefinition(value_, "float")
	result.Fields[built_in_declarations.FloatCeilingMethod.Name] = function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			built_in_declarations.FloatCeilingMethod.Name,
		),

		func(_ *runtime.Runtime, _ ...value.Value) value.Value {
			return FloatValue(math.Ceil(float64(value_)))
		},

		built_in_declarations.FloatCeilingMethod.Type,
	)

	result.Fields[built_in_declarations.FloatFloorMethod.Name] = function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			built_in_declarations.FloatFloorMethod.Name,
		),

		func(_ *runtime.Runtime, _ ...value.Value) value.Value {
			return FloatValue(math.Floor(float64(value_)))
		},

		built_in_declarations.FloatFloorMethod.Type,
	)

	result.Fields[built_in_declarations.FloatToIntegerMethod.Name] = function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			built_in_declarations.FloatToIntegerMethod.Name,
		),

		func(_ *runtime.Runtime, _ ...value.Value) value.Value {
			return IntegerValue(value_)
		},

		built_in_declarations.FloatToIntegerMethod.Type,
	)

	return result
}

type IntegerValue int64

func (value_ IntegerValue) Definition() *value.ValueDefinition {
	result := newNumberDefinition(value_, "int")
	result.Fields[built_in_declarations.IntegerToCharacterMethod.Name] =
		function.NewBuiltInFunction(
			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.IntegerToCharacterMethod.Name,
			),

			func(_ *runtime.Runtime, _ ...value.Value) value.Value {
				return StringValue([]rune{rune(value_)})
			},

			built_in_declarations.IntegerToCharacterMethod.Type,
		)

	result.Fields[built_in_declarations.IntegerToFloatMethod.Name] = function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(built_in_declarations.IntegerToFloatMethod.Name),
		func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
			return FloatValue(value_)
		},

		built_in_declarations.IntegerToFloatMethod.Type,
	)

	return result
}

func newMinusMethod[Value IntegerValue | FloatValue](value_ Value) *function.Function {
	return function.NewBuiltInFunction(
		function.NewIntersectionFunctionArgumentValidator(
			func(argumentTypes []reflect.Type) *errors.Error {
				if len(argumentTypes) == 1 {
					return runtime_errors.IncorrectBuiltInFunctionArgumentType(
						built_in_declarations.NumericMinusMethod.Name,
						0,
					)
				}

				return runtime_errors.IncorrectCallArgumentCount(
					"0-1",
					true,
					len(argumentTypes),
				)
			},

			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.NumericMinusMethod.Name,
			),

			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.NumericMinusMethod.Name,
				reflect.TypeOf(value_),
			),
		),

		func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
			if len(arguments) == 0 {
				return value.Value(-value_)
			}

			return value.Value(value_ - arguments[0].(Value))
		},

		built_in_declarations.NumericMinusMethod.Type,
	)
}

func newNumberDefinition[Value IntegerValue | FloatValue](
	value_ Value,
	valueTypeName string,
) *value.ValueDefinition {
	valueType := reflect.TypeOf(value_)

	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.NumericPlusMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericPlusMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value.Value(value_ + arguments[0].(Value))
				},

				built_in_declarations.NumericPlusMethod.Type,
			),

			built_in_declarations.NumericMinusMethod.Name: newMinusMethod(value_),
			built_in_declarations.NumericTimesMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericTimesMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value.Value(value_ * arguments[0].(Value))
				},

				built_in_declarations.NumericTimesMethod.Type,
			),

			built_in_declarations.NumericOverMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericOverMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					rightHandSide := arguments[0].(Value)

					if rightHandSide == 0 {
						errors.RaiseError(
							runtime_errors.DivisionByZero(
								valueTypeName,
								built_in_declarations.NumericOverMethod.Name,
							),
						)
					}

					return value.Value(value_ / rightHandSide)
				},

				built_in_declarations.NumericOverMethod.Type,
			),

			built_in_declarations.NumericModuloMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericModuloMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					modulus := arguments[0].(Value)

					if modulus == 0 {
						errors.RaiseError(
							runtime_errors.DivisionByZero(
								valueTypeName,
								built_in_declarations.NumericModuloMethod.Name,
							),
						)
					}

					switch value_ := any(value_).(type) {
					case IntegerValue:
						return value_ % IntegerValue(modulus)

					case FloatValue:
						return FloatValue(math.Mod(float64(value_), float64(modulus)))

					default:
						return nil
					}
				},

				built_in_declarations.NumericModuloMethod.Type,
			),

			built_in_declarations.NumericLessThanMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericLessThanMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ < arguments[0].(Value))
				},

				built_in_declarations.NumericLessThanMethod.Type,
			),

			built_in_declarations.NumericLessThanOrEqualToMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericLessThanOrEqualToMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ <= arguments[0].(Value))
				},

				built_in_declarations.NumericLessThanOrEqualToMethod.Type,
			),

			built_in_declarations.NumericGreaterThanMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericGreaterThanMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ > arguments[0].(Value))
				},

				built_in_declarations.NumericGreaterThanMethod.Type,
			),

			built_in_declarations.NumericGreaterThanOrEqualToMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.NumericGreaterThanOrEqualToMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ >= arguments[0].(Value))
				},

				built_in_declarations.NumericGreaterThanOrEqualToMethod.Type,
			),
		},
	}
}
