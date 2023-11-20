package value_types

import "project_umbrella/interpreter/runtime/value"

type TupleValue struct {
	Elements []value.Value
}

func (value_ TupleValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{},
	}
}
