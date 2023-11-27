package built_in_declarations

import "project_umbrella/interpreter/parser/parser_types"

type BuiltInField struct {
	Name string
	Type *parser_types.FunctionType
}

// Implemented on every type
var (
	EqualsMethod = &BuiltInField{
		Name: "==",
		Type: parser_types.InfixFunction,
	}

	NotEqualsMethod = &BuiltInField{
		Name: "!=",
		Type: parser_types.InfixFunction,
	}

	ToStringMethod = &BuiltInField{
		Name: "__to_str__",
		Type: parser_types.NormalFunction,
	}
)

// Implemented on int and float
var (
	PlusMethod = &BuiltInField{ // Implemented on str, int, and float
		Name: "+",
		Type: parser_types.InfixFunction,
	}

	MinusMethod = &BuiltInField{
		Name: "-",
		Type: parser_types.InfixPrefixFunction,
	}

	TimesMethod = &BuiltInField{
		Name: "*",
		Type: parser_types.InfixFunction,
	}

	OverMethod = &BuiltInField{
		Name: "/",
		Type: parser_types.InfixFunction,
	}

	ModuloMethod = &BuiltInField{
		Name: "%",
		Type: parser_types.InfixFunction,
	}

	LessThanMethod = &BuiltInField{
		Name: "<",
		Type: parser_types.InfixFunction,
	}

	LessThanOrEqualToMethod = &BuiltInField{
		Name: "<=",
		Type: parser_types.InfixFunction,
	}

	GreaterThanMethod = &BuiltInField{
		Name: ">",
		Type: parser_types.InfixFunction,
	}

	GreaterThanOrEqualToMethod = &BuiltInField{
		Name: ">=",
		Type: parser_types.InfixFunction,
	}
)

// Implemented on bool
var (
	NotMethod = &BuiltInField{
		Name: "!",
		Type: parser_types.PrefixFunction,
	}

	AndMethod = &BuiltInField{
		Name: "&&",
		Type: parser_types.InfixFunction,
	}

	OrMethod = &BuiltInField{
		Name: "||",
		Type: parser_types.InfixFunction,
	}
)

// Implemented on tuple
var (
	GetMethod = &BuiltInField{
		Name: "get",
		Type: parser_types.NormalFunction,
	}

	LengthField = &BuiltInField{
		Name: "length",
		Type: nil,
	}
)

// Implemented on structs
var StructConstructorMethod = &BuiltInField{
	Name: "__constructor__",
	Type: parser_types.NormalFunction,
}

type BuiltInValueID int

const (
	FalseValueID BuiltInValueID = -iota - 1
	TrueValueID
	UnitValueID

	IfElseFunctionID
	ImportFunctionID
	ModuleFunctionID
	PrintFunctionID
	PrintlnFunctionID
	TupleFunctionID
	StructFunctionID
)
