package value_types

import "project_umbrella/interpreter/runtime/value"

type TupleValue struct {
	Elements []value.Value
}

func (value_ *TupleValue) Definition() *value.ValueDefinition {
	return orderedCollectionDefinition(
		value_,
		"tuple",
		len(value_.Elements),
		func(i int) value.Value {
			return value_.Elements[i]
		},

		func(other *TupleValue) *TupleValue {
			return &TupleValue{
				Elements: append(value_.Elements, other.Elements...),
			}
		},

		func() *TupleValue {
			return &TupleValue{
				Elements: []value.Value{},
			}
		},

		func(start int, end int) *TupleValue {
			return &TupleValue{
				Elements: value_.Elements[start:end],
			}
		},
	)
}
