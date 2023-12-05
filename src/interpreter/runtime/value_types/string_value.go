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
	result := orderedCollectionDefinition(
		value_,
		"string",
		len([]rune(value_)),
		func(i int) value.Value {
			return StringValue([]rune(value_)[i])
		},

		func(other StringValue) StringValue {
			return value_ + other
		},

		func() StringValue {
			return ""
		},

		func(start int, end int) StringValue {
			return StringValue([]rune(value_)[start:end])
		},
	)

	result.Fields[built_in_declarations.StringCodepointMethod.Name] = function.NewBuiltInFunction(
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
	)

	result.Fields[built_in_declarations.StringSplit.Name] = function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			built_in_declarations.StringSplit.Name,
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
	)

	return result
}
