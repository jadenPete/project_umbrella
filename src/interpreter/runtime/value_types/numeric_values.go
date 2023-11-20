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
	return newNumberDefinition(value_, "float")
}

type IntegerValue int64

func (value_ IntegerValue) Definition() *value.ValueDefinition {
	return newNumberDefinition(value_, "int")
}

func newMinusMethod[Value IntegerValue | FloatValue](value_ Value) *function.Function {
	return function.NewBuiltInFunction(
		function.NewIntersectionFunctionArgumentValidator(
			func(argumentTypes []reflect.Type) *errors.Error {
				if len(argumentTypes) == 1 {
					return runtime_errors.IncorrectBuiltInFunctionArgumentType(
						built_in_declarations.MinusMethod.Name,
						0,
					)
				}

				return runtime_errors.IncorrectCallArgumentCount(
					"0-1",
					true,
					len(argumentTypes),
				)
			},

			function.NewFixedFunctionArgumentValidator(built_in_declarations.MinusMethod.Name),
			function.NewFixedFunctionArgumentValidator(built_in_declarations.MinusMethod.Name, reflect.TypeOf(value_)),
		),

		func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
			if len(arguments) == 0 {
				return value.Value(-value_)
			}

			return value.Value(value_ - arguments[0].(Value))
		},

		built_in_declarations.MinusMethod.Type,
	)
}

func newNumberDefinition[Value IntegerValue | FloatValue](
	value_ Value,
	valueTypeName string,
) *value.ValueDefinition {
	valueType := reflect.TypeOf(value_)

	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.PlusMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.PlusMethod.Name, valueType),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value.Value(value_ + arguments[0].(Value))
				},

				built_in_declarations.PlusMethod.Type,
			),

			built_in_declarations.MinusMethod.Name: newMinusMethod(value_),
			built_in_declarations.TimesMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.TimesMethod.Name, valueType),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value.Value(value_ * arguments[0].(Value))
				},

				built_in_declarations.TimesMethod.Type,
			),

			built_in_declarations.OverMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.OverMethod.Name, valueType),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					rightHandSide := arguments[0].(Value)

					if rightHandSide == 0 {
						errors.RaiseError(
							runtime_errors.DivisionByZero(valueTypeName, built_in_declarations.OverMethod.Name),
						)
					}

					return value.Value(value_ / rightHandSide)
				},

				built_in_declarations.OverMethod.Type,
			),

			built_in_declarations.ModuloMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.ModuloMethod.Name, valueType),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					modulus := arguments[0].(Value)

					if modulus == 0 {
						errors.RaiseError(
							runtime_errors.DivisionByZero(valueTypeName, built_in_declarations.ModuloMethod.Name),
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

				built_in_declarations.ModuloMethod.Type,
			),

			built_in_declarations.LessThanMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.LessThanMethod.Name, valueType),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ < arguments[0].(Value))
				},

				built_in_declarations.LessThanMethod.Type,
			),

			built_in_declarations.LessThanOrEqualToMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.LessThanOrEqualToMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ <= arguments[0].(Value))
				},

				built_in_declarations.LessThanOrEqualToMethod.Type,
			),

			built_in_declarations.GreaterThanMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.GreaterThanMethod.Name, valueType),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ > arguments[0].(Value))
				},

				built_in_declarations.GreaterThanMethod.Type,
			),

			built_in_declarations.GreaterThanOrEqualToMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.GreaterThanOrEqualToMethod.Name,
					valueType,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return BooleanValue(value_ >= arguments[0].(Value))
				},

				built_in_declarations.GreaterThanOrEqualToMethod.Type,
			),
		},
	}
}
