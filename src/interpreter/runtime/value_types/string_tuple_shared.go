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

func orderedCollectionDefinition[Value value.Value](
	value_ Value,
	valueTypeName string,
	valueLength int,
	valueElement func(int) value.Value,
	appendValue func(Value) Value,
	emptyValue func() Value,
	sliceValue func(int, int) Value,
) *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.OrderedGetMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.OrderedGetMethod.Name,
					reflect.TypeOf(*new(IntegerValue)),
				),

				func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
					i := int(arguments[0].(IntegerValue))

					if i < 0 || i >= valueLength {
						errors.RaiseError(
							runtime_errors.IndexOutOfBounds(
								valueTypeName,
								built_in_declarations.OrderedGetMethod.Name,
								i,
								valueLength-1,
							),
						)
					}

					return valueElement(i)
				},

				built_in_declarations.OrderedGetMethod.Type,
			),

			built_in_declarations.OrderedLengthField.Name: IntegerValue(valueLength),
			built_in_declarations.OrderedPlusMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.OrderedPlusMethod.Name,
					reflect.TypeOf(*new(Value)),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					return appendValue(arguments[0].(Value))
				},

				built_in_declarations.OrderedPlusMethod.Type,
			),

			built_in_declarations.OrderedSliceMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.OrderedSliceMethod.Name,
					reflect.TypeOf(*new(IntegerValue)),
					reflect.TypeOf(*new(IntegerValue)),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					start := int(arguments[0].(IntegerValue))
					end := int(arguments[1].(IntegerValue))

					if start < 0 {
						start = 0
					} else if start > valueLength {
						start = valueLength
					}

					if end < 0 {
						end = 0
					} else if end > valueLength {
						end = valueLength
					}

					end = max(end, start)

					if start < 0 {
						start = 0
					}

					return sliceValue(start, end)
				},

				built_in_declarations.OrderedSliceMethod.Type,
			),
		},
	}
}
