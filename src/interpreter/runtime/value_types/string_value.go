package value_types

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator/built_ins"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type StringValue struct {
	Content string
}

func (value_ StringValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_ins.PlusMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator("+", reflect.TypeOf(*new(StringValue))),
				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return StringValue{value_.Content + arguments[0].(StringValue).Content}
				},

				built_ins.PlusMethod.Type,
			),
		},
	}
}
