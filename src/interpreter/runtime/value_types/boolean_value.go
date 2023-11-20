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
			built_in_declarations.NotMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.NotMethod.Name),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return !value_
				},

				built_in_declarations.NotMethod.Type,
			),

			built_in_declarations.AndMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.AndMethod.Name, reflect.TypeOf(value_)),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value_ && arguments[0].(BooleanValue)
				},

				built_in_declarations.AndMethod.Type,
			),

			built_in_declarations.OrMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_in_declarations.OrMethod.Name, reflect.TypeOf(value_)),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value_ || arguments[0].(BooleanValue)
				},

				built_in_declarations.OrMethod.Type,
			),
		},
	}
}
