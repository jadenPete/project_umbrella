package parser_types

type FunctionType struct {
	IsInfix  bool
	IsPrefix bool
	IsLookup bool
}

func (functionType FunctionType) CanSelectBy(selectType SelectType) bool {
	switch selectType {
	case NormalSelect:
		return true

	case InfixSelect:
		return functionType.IsInfix

	case PrefixSelect:
		return functionType.IsPrefix

	}

	return false
}

var NormalFunction = &FunctionType{
	IsInfix:  false,
	IsPrefix: false,
	IsLookup: false,
}

var InfixFunction = &FunctionType{
	IsInfix:  true,
	IsPrefix: false,
	IsLookup: false,
}

var InfixPrefixFunction = &FunctionType{
	IsInfix:  true,
	IsPrefix: true,
	IsLookup: false,
}

var PrefixFunction = &FunctionType{
	IsInfix:  false,
	IsPrefix: true,
	IsLookup: false,
}

type SelectType int

const (
	NormalSelect SelectType = iota + 1
	InfixSelect
	PrefixSelect
)
