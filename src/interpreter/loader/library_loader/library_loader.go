package library_loader

import (
	"plugin"
	"reflect"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types/library"
)

func LoadLibrary(path string) *library.Library {
	plugin_, err := plugin.Open(path)

	if err != nil {
		panic(err)
	}

	return &library.Library{
		Path: path,
		GetField: func(name string) (value.Value, bool) {
			symbol, err := plugin_.Lookup(name)

			if err != nil {
				return nil, false
			}

			symbolValue := reflect.ValueOf(symbol)

			if symbolValue.Kind() == reflect.Pointer {
				if symbolDereferencedValue, ok := symbolValue.Elem().Interface().(value.Value); ok {
					return symbolDereferencedValue, true
				}
			}

			errors.RaiseError(runtime_errors.LibrarySymbolNotValue(path, name))

			return nil, false
		},
	}
}
