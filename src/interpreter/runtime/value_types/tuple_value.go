package value_types

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type TupleValue struct {
	Elements []value.Value
}

func (value_ *TupleValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.TupleGetMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.TupleGetMethod.Name,
					reflect.TypeOf(*new(IntegerValue)),
				),

				func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
					i := int(arguments[0].(IntegerValue))

					if i < 0 || i >= len(value_.Elements) {
						errors.RaiseError(
							runtime_errors.TupleGetIndexOutOfBounds(i, len(value_.Elements)),
						)
					}

					return value_.Elements[i]
				},

				built_in_declarations.TupleGetMethod.Type,
			),

			built_in_declarations.TupleLengthField.Name: IntegerValue(len(value_.Elements)),
		},
	}
}
