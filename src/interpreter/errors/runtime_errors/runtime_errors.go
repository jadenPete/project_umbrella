package runtime_errors

import (
	"fmt"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/parser/parser_types"
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

func MethodCalledImproperly(
	firstOperand string,
	fieldName string,
	functionType *parser_types.FunctionType,
	selectType parser_types.SelectType,
) *errors.Error {
	var selectTypeName string

	switch selectType {
	case parser_types.InfixSelect:
		selectTypeName = "infix"

	case parser_types.PrefixSelect:
		selectTypeName = "prefix"

	default:
		panic("Expected `selectType` to be either `InfixField` or `PrefixField`.")
	}

	var expectedSyntax string

	if functionType.IsInfix {
		expectedSyntax = fmt.Sprintf("%s %s ...", firstOperand, fieldName)
	} else if functionType.IsPrefix {
		expectedSyntax = fmt.Sprintf("%s%s", fieldName, firstOperand)
	} else {
		expectedSyntax = fmt.Sprintf("%s.%s(...)", firstOperand, fieldName)
	}

	return &errors.Error{
		Section: "RUNTIME",
		Code:    10,
		Name: fmt.Sprintf(
			"`%s` is not an %s method and cannot be called so",
			fieldName,
			selectTypeName,
		),

		Description: fmt.Sprintf(
			"Consider replacing that call with `%s`, substituting in the right-hand operand.",
			expectedSyntax,
		),
	}
}

var NonStringFieldName = &errors.Error{
	Section: "RUNTIME",
	Code:    11,
	Name:    "A constant value identifying a field name is not a string",
}

func UniversalMethodReturnedIncorrectValue(
	methodName string,
	expectedTypeName string,
) *errors.Error {
	return &errors.Error{
		Section:     "RUNTIME",
		Code:        12,
		Name:        "A universal method returned a value of an incorrect type",
		Description: fmt.Sprintf("%s should've returned a %s", methodName, expectedTypeName),
	}
}
