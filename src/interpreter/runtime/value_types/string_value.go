package value_types

import (
	"reflect"
	"strings"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type StringValue string

func (value_ StringValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.StringCodepointMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.StringCodepointMethod.Name,
				),

				func(_ *runtime.Runtime, _ ...value.Value) value.Value {
					if len([]rune(value_)) != 1 {
						errors.RaiseError(
							runtime_errors.CodepointCalledOnNonCharacter(string(value_)),
						)
					}

					return IntegerValue([]rune(value_)[0])
				},

				built_in_declarations.StringCodepointMethod.Type,
			),

			built_in_declarations.StringGetMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.TupleGetMethod.Name,
					reflect.TypeOf(*new(IntegerValue)),
				),

				func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
					i := int(arguments[0].(IntegerValue))

					if i < 0 || i >= len(value_) {
						errors.RaiseError(
							runtime_errors.IndexOutOfBounds(
								"str",
								built_in_declarations.StringGetMethod.Name,
								i,
								len(value_),
							),
						)
					}

					return StringValue([]rune(value_)[i])
				},

				built_in_declarations.StringGetMethod.Type,
			),

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

			built_in_declarations.StringSplit.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.StringPlusMethod.Name,
					reflect.TypeOf(*new(StringValue)),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					result := &TupleValue{
						Elements: []value.Value{},
					}

					for _, component := range strings.Split(
						string(value_),
						string(arguments[0].(StringValue)),
					) {
						result.Elements = append(result.Elements, StringValue(component))
					}

					return result
				},

				built_in_declarations.StringSplit.Type,
			),
		},
	}
}
