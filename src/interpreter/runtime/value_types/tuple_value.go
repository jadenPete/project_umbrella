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
							runtime_errors.IndexOutOfBounds(
								"tuple",
								built_in_declarations.TupleGetMethod.Name,
								i,
								len(value_.Elements)-1,
							),
						)
					}

					return value_.Elements[i]
				},

				built_in_declarations.TupleGetMethod.Type,
			),

			built_in_declarations.TupleLengthField.Name: IntegerValue(len(value_.Elements)),
			built_in_declarations.TuplePlusMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.TuplePlusMethod.Name,
					reflect.TypeOf(&TupleValue{}),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return &TupleValue{
						Elements: append(value_.Elements, arguments[0].(*TupleValue).Elements...),
					}
				},

				built_in_declarations.TuplePlusMethod.Type,
			),

			built_in_declarations.TupleSliceMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.TuplePlusMethod.Name,
					reflect.TypeOf(*new(IntegerValue)),
					reflect.TypeOf(*new(IntegerValue)),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					start := int(arguments[0].(IntegerValue))
					end := int(arguments[1].(IntegerValue))

					var resultElements []value.Value

					if start < 0 ||
						start > len(value_.Elements) ||
						end < 0 ||
						end > len(value_.Elements) ||
						end < start {
						resultElements = []value.Value{}
					} else {
						resultElements = value_.Elements[start:end]
					}

					return &TupleValue{
						Elements: resultElements,
					}
				},

				built_in_declarations.TupleSliceMethod.Type,
			),
		},
	}
}
