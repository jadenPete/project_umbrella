package runtime_errors

import (
	"fmt"
	"strings"

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

func ModuleNotFound(moduleName string) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    13,
		Name:    fmt.Sprintf("The module \"%s\" wasn't found", moduleName),
	}
}

func ModuleCycle(moduleLoaderStack []string) *errors.Error {
	var description strings.Builder

	description.WriteString(
		fmt.Sprintf(
			"\"%s\" couldn't be imported. See the following import stack.\n\n%s",
			moduleLoaderStack[len(moduleLoaderStack)-1],
			moduleLoaderStack[0],
		),
	)

	for _, moduleName := range moduleLoaderStack[1:] {
		description.WriteString(fmt.Sprintf("\n↳ %s", moduleName))
	}

	return &errors.Error{
		Section:     "RUNTIME",
		Code:        13,
		Name:        "Encountered an import cycle",
		Description: description.String(),
	}
}

func IndexOutOfBounds(typeName string, methodName string, i int, maximumIndex int) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    14,
		Name:    fmt.Sprintf("An out-of-bounds index was provided to %s#%s", typeName, methodName),
		Description: fmt.Sprintf(
			"Expected an index in the range [0, %d), but got %d.",
			maximumIndex+1,
			i,
		),
	}
}

func LibraryNotFound(libraryName string) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    15,
		Name:    fmt.Sprintf("The library \"%s\" wasn't found", libraryName),
	}
}

func librarySymbolError(
	libraryPath string,
	symbolName string,
	code int,
	description string,
) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    code,
		Name: fmt.Sprintf(
			"Couldn't fetch the symbol \"%s\" from the library at \"%s\"",
			symbolName,
			libraryPath,
		),

		Description: description,
	}
}

func LibrarySymbolNotValue(libraryPath string, symbolName string) *errors.Error {
	return librarySymbolError(
		libraryPath,
		symbolName,
		16,
		fmt.Sprintf("\"%s\" isn't a value.", symbolName),
	)
}

func LibrarySymbolNotFound(libraryPath string, symbolName string) *errors.Error {
	return librarySymbolError(
		libraryPath,
		symbolName,
		17,
		fmt.Sprintf("\"%s\" doesn't exist.", symbolName),
	)
}

func CodepointCalledOnNonCharacter(stringValue string) *errors.Error {
	return &errors.Error{
		Section: "RUNTIME",
		Code:    18,
		Name:    "`codepoint` was called on a non-character",
		Description: fmt.Sprintf(
			"`codepoint` was called on a string of length %d: \"%s\"",
			len([]rune(stringValue)),
			stringValue,
		),
	}
}
