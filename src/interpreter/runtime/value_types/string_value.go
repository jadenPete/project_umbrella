package value_types

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type StringValue string

func (value_ StringValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.StringLengthField.Name: IntegerValue(len([]rune(value_))),
			built_in_declarations.StringPlusMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.StringPlusMethod.Name,
					reflect.TypeOf(*new(StringValue)),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return value_ + arguments[0].(StringValue)
				},

				built_in_declarations.StringPlusMethod.Type,
			),
		},
	}
}
