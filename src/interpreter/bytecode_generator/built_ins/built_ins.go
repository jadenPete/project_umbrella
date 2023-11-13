package built_ins

import "project_umbrella/interpreter/parser/parser_types"

type BuiltInField struct {
	Name string
	Type parser_types.FieldType
}

var (
	// Implemented on every type
	ToStringMethod = &BuiltInField{
		Name: "__to_str__",
		Type: parser_types.NormalField,
	}

	EqualsMethod = &BuiltInField{
		Name: "==",
		Type: parser_types.InfixField,
	}

	NotEqualsMethod = &BuiltInField{
		Name: "!=",
		Type: parser_types.InfixField,
	}

	// Implemented on int and float
	PlusMethod = &BuiltInField{ // Implemented on str, int, and float
		Name: "+",
		Type: parser_types.InfixField,
	}

	MinusMethod = &BuiltInField{
		Name: "-",
		Type: parser_types.InfixPrefixField,
	}

	TimesMethod = &BuiltInField{
		Name: "*",
		Type: parser_types.InfixField,
	}

	OverMethod = &BuiltInField{
		Name: "/",
		Type: parser_types.InfixField,
	}

	ModuloMethod = &BuiltInField{
		Name: "%",
		Type: parser_types.InfixField,
	}

	LessThanMethod = &BuiltInField{
		Name: "<",
		Type: parser_types.InfixField,
	}

	LessThanOrEqualToMethod = &BuiltInField{
		Name: "<=",
		Type: parser_types.InfixField,
	}

	GreaterThanMethod = &BuiltInField{
		Name: ">",
		Type: parser_types.InfixField,
	}

	GreaterThanOrEqualToMethod = &BuiltInField{
		Name: ">=",
		Type: parser_types.InfixField,
	}

	// Implemented on bool
	NotMethod = &BuiltInField{
		Name: "!",
		Type: parser_types.PrefixField,
	}

	AndMethod = &BuiltInField{
		Name: "&&",
		Type: parser_types.InfixField,
	}

	OrMethod = &BuiltInField{
		Name: "||",
		Type: parser_types.InfixField,
	}
)

type BuiltInValueID int

const (
	PrintFunctionID BuiltInValueID = -iota - 1
	PrintlnFunctionID

	UnitValueID
	FalseValueID
	TrueValueID

	IfElseFunctionID
	TupleFunctionID
	StructFunctionID
)
