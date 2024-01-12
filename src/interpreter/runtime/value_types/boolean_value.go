package value_types

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type BooleanValue bool

func (value_ BooleanValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.BooleanNotMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.BooleanNotMethod.Name,
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return !value_
				},

				built_in_declarations.BooleanNotMethod.Type,
			),

			built_in_declarations.BooleanAndMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.BooleanAndMethod.Name,
					reflect.TypeOf(value_),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value_ && arguments[0].(BooleanValue)
				},

				built_in_declarations.BooleanAndMethod.Type,
			),

			built_in_declarations.BooleanOrMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.BooleanOrMethod.Name,
					reflect.TypeOf(value_),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value_ || arguments[0].(BooleanValue)
				},

				built_in_declarations.BooleanOrMethod.Type,
			),
		},
	}
}
