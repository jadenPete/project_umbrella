package value_types

import "project_umbrella/interpreter/runtime/value"

type UnitValue struct{}

func (UnitValue) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{},
	}
}
