package built_in_declarations

import "project_umbrella/interpreter/parser/parser_types"

type BuiltInField struct {
	Name string
	Type *parser_types.FunctionType
}

// Implemented on every type
var (
	UniversalEqualsMethod = &BuiltInField{
		Name: "==",
		Type: parser_types.InfixFunction,
	}

	UniversalNotEqualsMethod = &BuiltInField{
		Name: "!=",
		Type: parser_types.InfixFunction,
	}

	UniversalToStringMethod = &BuiltInField{
		Name: "__to_str__",
		Type: parser_types.NormalFunction,
	}
)

// Implemented on str
var (
	StringLengthField = &BuiltInField{
		Name: "length",
		Type: nil,
	}

	StringPlusMethod = &BuiltInField{
		Name: "+",
		Type: parser_types.InfixFunction,
	}
)

// Implemented on int and float
var (
	NumericPlusMethod = &BuiltInField{
		Name: "+",
		Type: parser_types.InfixFunction,
	}

	NumericMinusMethod = &BuiltInField{
		Name: "-",
		Type: parser_types.InfixPrefixFunction,
	}

	NumericTimesMethod = &BuiltInField{
		Name: "*",
		Type: parser_types.InfixFunction,
	}

	NumericOverMethod = &BuiltInField{
		Name: "/",
		Type: parser_types.InfixFunction,
	}

	NumericModuloMethod = &BuiltInField{
		Name: "%",
		Type: parser_types.InfixFunction,
	}

	NumericLessThanMethod = &BuiltInField{
		Name: "<",
		Type: parser_types.InfixFunction,
	}

	NumericLessThanOrEqualToMethod = &BuiltInField{
		Name: "<=",
		Type: parser_types.InfixFunction,
	}

	NumericGreaterThanMethod = &BuiltInField{
		Name: ">",
		Type: parser_types.InfixFunction,
	}

	NumericGreaterThanOrEqualToMethod = &BuiltInField{
		Name: ">=",
		Type: parser_types.InfixFunction,
	}
)

// Implemented on bool
var (
	BooleanNotMethod = &BuiltInField{
		Name: "!",
		Type: parser_types.PrefixFunction,
	}

	BooleanAndMethod = &BuiltInField{
		Name: "&&",
		Type: parser_types.InfixFunction,
	}

	BooleanOrMethod = &BuiltInField{
		Name: "||",
		Type: parser_types.InfixFunction,
	}
)

// Implemented on tuple
var (
	TupleGetMethod = &BuiltInField{
		Name: "get",
		Type: parser_types.NormalFunction,
	}

	TupleLengthField = &BuiltInField{
		Name: "length",
		Type: nil,
	}
)

// Implemented on structs
var StructConstructorMethod = &BuiltInField{
	Name: "__constructor__",
	Type: parser_types.NormalFunction,
}

// Implemented on libraries
var (
	LibraryGetMethod = &BuiltInField{
		Name: "get",
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
	ImportLibraryFunctionID
	ModuleFunctionID
	TupleFunctionID
	StructFunctionID
)
