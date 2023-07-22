package runtime_errors

import (
	"fmt"

	"project_umbrella/interpreter/errors"
)

func IncorrectCallArgumentCount(expectedArgumentCount int, actualArgumentCount int) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    1,
		Name: fmt.Sprintf(
			"A function accepting %d arguments was called with %d arguments",
			expectedArgumentCount,
			actualArgumentCount,
		),
	}
}

func IncorrectBuiltInInfixMethodArgumentType(typeName string, methodName string) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    2,
		Name:    "Incorrectly typed argument to a built-in infix method",
		Description: fmt.Sprintf(
			"Expected the right-hand side of %[1]s#%[2]s to be of type %[1]s.",
			typeName,
			methodName,
		),
	}
}

var ToStrReturnedNonString = &errors.Error{
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

func DivisionByZero(typeName string) *errors.Error {
	return &errors.Error{
		Section:     "RUNTIME",
		Code:        7,
		Name:        "Cannot divide by zero",
		Description: fmt.Sprintf("Expected the right-hand side of %s#/ to be nonzero.", typeName),
	}
}
