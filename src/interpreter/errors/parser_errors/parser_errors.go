package parser_errors

import (
	"fmt"

	"project_umbrella/interpreter/errors"
)

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

func NonInfixMethodCalledImproperly(leftHandSide string, fieldName string) *errors.Error {
	return &errors.Error{
		Section: "PARSER",
		Code:    8,
		Name: fmt.Sprintf(
			"`%s` is not an infix method and cannot be called so",
			fieldName,
		),

		Description: fmt.Sprintf(
			"Consider replacing that call with `%s.%s(...)`, substituting in the right-hand operand.",
			leftHandSide,
			fieldName,
		),
	}
}
