package value

type Value interface {
	Definition() *ValueDefinition
}

type ValueDefinition struct {
	Fields map[string]Value
}
