package parser_errors

import (
	"fmt"

	"github.com/alecthomas/participle/v2"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/parser/parser_types"
)

func ParserFailed(err participle.Error) *errors.Error {
	return &errors.Error{
		Section: "PARSER",
		Code:    1,
		Name:    fmt.Sprintf("The parser failed: %s", err.Message()),
	}
}

var BytecodeEncodingFailed = &errors.Error{
	Section: "PARSER",
	Code:    2,
	Name:    "Couldn't encode the resulting bytecode",
}

var InvalidRootExpression = &errors.Error{
	Section: "PARSER",
	Code:    3,
	Name:    "Expected an expression list, but got a different expression",
}

var NonexhaustiveConstantIDMap = &errors.Error{
	Section: "PARSER",
	Code:    4,
	Name:    "Nonexhaustive constant ID map",
}

var ValueReassigned = &errors.Error{
	Section:     "PARSER",
	Code:        5,
	Name:        "Reassigning to an already declared value is impossible",
	Description: "Consider assigning to a new value.",
}

func UnknownValue(valueName string) *errors.Error {
	return &errors.Error{
		Section: "PARSER",
		Code:    6,
		Name:    fmt.Sprintf("Unknown value: `%s`", valueName),
	}
}

func UnknownField(fieldName string) *errors.Error {
	return &errors.Error{
		Section: "PARSER",
		Code:    7,
		Name:    fmt.Sprintf("Unknown field: `%s`", fieldName),
	}
}

func MethodCalledImproperly(
	firstOperand string,
	fieldName string,
	fieldType parser_types.FieldType,
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

	switch fieldType {
	case parser_types.NormalField:
		expectedSyntax = fmt.Sprintf("%s.%s(...)", firstOperand, fieldName)

	case parser_types.InfixField:
		expectedSyntax = fmt.Sprintf("%s %s ...", firstOperand, fieldName)

	case parser_types.PrefixField:
		expectedSyntax = fmt.Sprintf("%s%s", fieldName, firstOperand)
	}

	return &errors.Error{
		Section: "PARSER",
		Code:    8,
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
