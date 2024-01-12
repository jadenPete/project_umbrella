package main

import (
	"math"
	"reflect"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
)

var SquareRoot = function.NewBuiltInFunction(
	function.NewIntersectionFunctionArgumentValidator(
		func(argumentTypes []reflect.Type) *errors.Error {
			return runtime_errors.IncorrectCallArgumentCount("1", false, len(argumentTypes))
		},

		function.NewFixedFunctionArgumentValidator(
			"square_root",
			reflect.TypeOf(*new(value_types.FloatValue)),
		),

		function.NewFixedFunctionArgumentValidator(
			"square_root",
			reflect.TypeOf(*new(value_types.IntegerValue)),
		),
	),

	func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
		switch argument := arguments[0].(type) {
		case value_types.FloatValue:
			return value_types.FloatValue(math.Sqrt(float64(argument)))

		case value_types.IntegerValue:
			return value_types.IntegerValue(math.Sqrt(float64(argument)))
		}

		return nil
	},

	parser_types.NormalFunction,
)
