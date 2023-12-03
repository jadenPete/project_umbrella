package main

import (
	"crypto/rand" // We use `crypto/rand` to ensure `RandomFloat` is different for every invocation
	"math"
	"math/big"
	"reflect"

	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
)

var MeaningOfLife value_types.IntegerValue = 42
var Square = function.NewBuiltInFunction(
	function.NewFixedFunctionArgumentValidator(
		"square_root",
		reflect.TypeOf(*new(value_types.IntegerValue)),
	),

	func(_ *runtime.Runtime, arguments ...value.Value) value.Value {
		argument := arguments[0].(value_types.IntegerValue)

		return argument * argument
	},

	parser_types.NormalFunction,
)

var InvalidSymbol = 0
var randomBigInteger = func() *big.Int {
	result, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))

	if err != nil {
		panic(err)
	}

	return result
}()

var RandomInteger = value_types.IntegerValue(randomBigInteger.Int64())
