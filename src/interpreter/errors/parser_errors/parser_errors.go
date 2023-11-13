package parser_errors

import (
	"fmt"

	"github.com/alecthomas/participle/v2"

	"project_umbrella/interpreter/errors"
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
