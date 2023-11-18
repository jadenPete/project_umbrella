package value_types

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator/built_ins"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type BooleanValue bool

func (value_ BooleanValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_ins.NotMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_ins.NotMethod.Name),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return !value_
				},

				built_ins.NotMethod.Type,
			),

			built_ins.AndMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_ins.AndMethod.Name, reflect.TypeOf(value_)),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value_ && arguments[0].(BooleanValue)
				},

				built_ins.AndMethod.Type,
			),

			built_ins.OrMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(built_ins.OrMethod.Name, reflect.TypeOf(value_)),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value_ || arguments[0].(BooleanValue)
				},

				built_ins.OrMethod.Type,
			),
		},
	}
}
