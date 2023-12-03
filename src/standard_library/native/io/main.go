package main

import (
	"fmt"
	"strings"

	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
	"project_umbrella/interpreter/runtime/value_util"
)

func print(
	runtime_ *runtime.Runtime,
	suffix string,
	arguments ...value.Value,
) value_types.UnitValue {
	serialized := make([]string, 0, len(arguments))

	for _, argument := range arguments {
		serialized = append(serialized, string(value_util.CallToStringMethod(runtime_, argument)))
	}

	fmt.Print(strings.Join(serialized, " ") + suffix)

	return value_types.UnitValue{}
}

var Print = function.NewBuiltInFunction(
	function.NewVariadicFunctionArgumentValidator("print", nil),
	func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
		return print(runtime_, "", arguments...)
	},

	parser_types.NormalFunction,
)

var Println = function.NewBuiltInFunction(
	function.NewVariadicFunctionArgumentValidator("println", nil),
	func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
		return print(runtime_, "\n", arguments...)
	},

	parser_types.NormalFunction,
)
