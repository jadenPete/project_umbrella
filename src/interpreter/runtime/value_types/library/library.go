package library

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
)

type Library struct {
	Path     string
	GetField func(string) (value.Value, bool)
}

func (library *Library) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{
			built_in_declarations.LibraryGetMethod.Name: function.NewBuiltInFunction(
				function.NewFixedFunctionArgumentValidator(
					built_in_declarations.LibraryGetMethod.Name,
					reflect.TypeOf(*new(value_types.StringValue)),
				),

				func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
					name := string(arguments[0].(value_types.StringValue))
					result, ok := library.GetField(name)

					if !ok {
						errors.RaiseError(runtime_errors.LibrarySymbolNotFound(library.Path, name))
					}

					return result
				},

				built_in_declarations.LibraryGetMethod.Type,
			),
		},
	}
}
