package built_ins

type BuiltInFieldID int

const (
	// Implemented on every type
	ToStringMethodID BuiltInFieldID = -iota - 1

	// Implemented on int and float
	PlusMethodID
	MinusMethodID
	TimesMethodID
	OverMethodID

	// Implemented on bool
	NotMethodID
	AndMethodID
	OrMethodID
)

type BuiltInValueID int

const (
	PrintFunctionID BuiltInValueID = -iota - 1
	PrintlnFunctionID
	UnitValueID
	FalseValueID
	TrueValueID
	IfElseFunctionID
)
