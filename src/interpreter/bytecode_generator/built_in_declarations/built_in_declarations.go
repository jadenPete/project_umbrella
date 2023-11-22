package built_in_declarations

import "project_umbrella/interpreter/parser/parser_types"

type BuiltInField struct {
	Name string
	Type *parser_types.FunctionType
}

var (
	// Implemented on every type
	ToStringMethod = &BuiltInField{
		Name: "__to_str__",
		Type: parser_types.NormalFunction,
	}

	EqualsMethod = &BuiltInField{
		Name: "==",
		Type: parser_types.InfixFunction,
	}

	NotEqualsMethod = &BuiltInField{
		Name: "!=",
		Type: parser_types.InfixFunction,
	}

	// Implemented on int and float
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

	// Implemented on bool
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

	// Implemented on structs
	StructConstructorMethod = &BuiltInField{
		Name: "__constructor__",
		Type: parser_types.NormalFunction,
	}
)

type BuiltInValueID int

const (
	FalseValueID BuiltInValueID = -iota - 1
	TrueValueID
	UnitValueID

	IfElseFunctionID
	ImportFunctionID
	PrintFunctionID
	PrintlnFunctionID
	TupleFunctionID
	StructFunctionID
)
