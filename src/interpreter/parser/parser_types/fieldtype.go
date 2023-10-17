package parser_types

type FieldType struct {
	IsInfix  bool
	IsPrefix bool
}

func (fieldType FieldType) CanSelectBy(selectType SelectType) bool {
	switch selectType {
	case NormalSelect:
		return true

	case InfixSelect:
		return fieldType.IsInfix

	case PrefixSelect:
		return fieldType.IsPrefix

	}

	return false
}

var NormalField = FieldType{
	IsInfix:  false,
	IsPrefix: false,
}

var InfixField = FieldType{
	IsInfix:  true,
	IsPrefix: false,
}

var InfixPrefixField = FieldType{
	IsInfix:  true,
	IsPrefix: true,
}

var PrefixField = FieldType{
	IsInfix:  false,
	IsPrefix: true,
}

type SelectType int

const (
	NormalSelect SelectType = iota + 1
	InfixSelect
	PrefixSelect
)
