package runtime_errors

import (
	"fmt"

	"project_umbrella/interpreter/errors"
)

func IncorrectCallArgumentCount(arity string, arityPlural bool, argumentCount int) *errors.Error {
	var argumentWord string

	if arityPlural {
		argumentWord = "arguments"
	} else {
		argumentWord = "argument"
	}

	return &errors.Error{
		Section: "RUNTIME",
		Code:    1,
		Name: fmt.Sprintf(
			"A function accepting %s %s was called with %d arguments",
			arity,
			argumentWord,
			argumentCount,
		),
	}
}

func IncorrectBuiltInFunctionArgumentType(functionName string, i int) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    2,
		Name:    "A built-in function was called with an argument of incorrect type",
		Description: fmt.Sprintf(
			"%s expected argument #%d to be of a different type.",
			functionName,
			i+1,
		),
	}
}

var ToStringMethodReturnedNonString = &errors.Error{
	Section: "RUNTIME",
	Code:    3,
	Name:    "A __to_str__ method returned a non-string",
}

func UnrecognizedFieldID(value string, fieldID int) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    4,
		Name:    "Unrecognized field ID",
		Description: fmt.Sprintf(
			"%d is not a recognized field ID for the value `%s`.",
			fieldID,
			value,
		),
	}
}

var ValueCycle = &errors.Error{
	Section: "RUNTIME",
	Code:    5,
	Name:    "Encountered a cycle between values",
}

var EmptyFunctionBlockGraph = &errors.Error{
	Section: "RUNTIME",
	Code:    6,
	Name:    "Encountered an empty function block graph",
}

func DivisionByZero(typeName string, methodName string) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    7,
		Name:    "Cannot divide by zero",
		Description: fmt.Sprintf(
			"Expected the right-hand side of %s#%s to be nonzero.",
			typeName,
			methodName,
		),
	}
}

var NonFunctionCalled = &errors.Error{
	Section: "RUNTIME",
	Code:    8,
	Name:    "A non-function was called",
}

func UnknownField(fieldName string) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    9,
		Name:    fmt.Sprintf("Unknown field: `%s`", fieldName),
	}
}
