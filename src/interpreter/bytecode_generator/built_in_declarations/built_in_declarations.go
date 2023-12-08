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

// Implemented on str and tuple
var (
	OrderedGetMethod = &BuiltInField{
		Name: "get",
		Type: parser_types.NormalFunction,
	}

	OrderedLengthField = &BuiltInField{
		Name: "length",
		Type: nil,
	}

	OrderedPlusMethod = &BuiltInField{
		Name: "+",
		Type: parser_types.InfixFunction,
	}

	OrderedSliceMethod = &BuiltInField{
		Name: "slice",
		Type: parser_types.NormalFunction,
	}
)

// Implemented on str
var (
	StringCodepointMethod = &BuiltInField{
		Name: "codepoint",
		Type: parser_types.NormalFunction,
	}

	StringSplit = &BuiltInField{
		Name: "split",
		Type: parser_types.NormalFunction,
	}

	StringStrip = &BuiltInField{
		Name: "strip",
		Type: parser_types.NormalFunction,
	}
)

// Implemented on int
var (
	IntegerToCharacterMethod = &BuiltInField{
		Name: "to_character",
		Type: parser_types.NormalFunction,
	}

	IntegerToFloatMethod = &BuiltInField{
		Name: "to_float",
		Type: parser_types.NormalFunction,
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

// Implemented on float
var (
	FloatCeilingMethod = &BuiltInField{
		Name: "ceil",
		Type: parser_types.NormalFunction,
	}

	FloatFloorMethod = &BuiltInField{
		Name: "floor",
		Type: parser_types.NormalFunction,
	}

	FloatToIntegerMethod = &BuiltInField{
		Name: "to_int",
		Type: parser_types.NormalFunction,
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

// Implemented on structs
var (
	StructIsInstanceOfMethod = &BuiltInField{
		Name: "__is_instance_of__",
		Type: parser_types.NormalFunction,
	}

	StructConstructorMethod = &BuiltInField{
		Name: "__constructor__",
		Type: parser_types.NormalFunction,
	}
)

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
