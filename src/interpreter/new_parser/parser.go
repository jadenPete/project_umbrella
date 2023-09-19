package main

// Union expressions

type Statement interface {
	statement()
}

type Expression interface {
	Statement

	expression()
}

type Primary interface {
	primary()
}

// Statements

type Assignment struct {
	Name  *Identifier `parser:"@IdentifierToken (IndentToken | OutdentToken | NewlineToken)* '=' (IndentToken | OutdentToken | NewlineToken)*"`
	Value Expression  `parser:" (@@"`
	Tail  *Assignment `parser:"| @@)"`
}

func (*Assignment) statement() {}

type Function struct {
	Declaration   *FunctionDeclaration   `parser:"@@ ':' NewlineToken"`
	StatementList *IndentedStatementList `parser:"@@"`
}

func (*Function) statement() {}

type FunctionDeclaration struct {
	Name       *Identifier         `parser:"'fn' (IndentToken | OutdentToken | NewlineToken)* @IdentifierToken (IndentToken | OutdentToken | NewlineToken)* '(' (IndentToken | OutdentToken | NewlineToken)*"`
	Parameters *FunctionParameters `parser:"@@? (IndentToken | OutdentToken | NewlineToken)* ')'"`
}

type FunctionParameters struct {
	Head *Identifier         `parser:"@IdentifierToken"`
	Tail *FunctionParameters `parser:"((IndentToken | OutdentToken | NewlineToken)* ',' (IndentToken | OutdentToken | NewlineToken)* @@)?"`
}

type IndentedStatementList struct {
	Value *StatementList `parser:"IndentToken @@ (OutdentToken | EOF)"`
}

type StatementList struct {
	Head Statement      `parser:"NewlineToken* @@"`
	Tail *StatementList `parser:"NewlineToken @@ | NewlineToken*"`
}

// Multi-token expressions

type InfixAddition struct {
	Left  *InfixMultiplication  `parser:"@@"`
	Right []*InfixAdditionRight `parser:"@@*"`
}

func (*InfixAddition) statement()  {}
func (*InfixAddition) expression() {}

type InfixAdditionRight struct {
	OperatorOne *Identifier          `parser:"(((IndentToken | OutdentToken)* @('+' | '-') (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *Identifier          `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @('+' | '-') (IndentToken | OutdentToken)*))"`
	Operand     *InfixMultiplication `parser:"@@"`
}

type InfixMultiplication struct {
	Left  *InfixMiscellaneous         `parser:"@@"`
	Right []*InfixMultiplicationRight `parser:"@@*"`
}

type InfixMultiplicationRight struct {
	OperatorOne *Identifier         `parser:"(((IndentToken | OutdentToken)* @('*' | '/') (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *Identifier         `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @('*' | '/') (IndentToken | OutdentToken)*))"`
	Operand     *InfixMiscellaneous `parser:"@@"`
}

type InfixMiscellaneous struct {
	Left  *Call                      `parser:"@@"`
	Right []*InfixMiscellaneousRight `parser:"@@*"`
}

type InfixMiscellaneousRight struct {
	OperatorOne *Identifier `parser:"(((IndentToken | OutdentToken)* @IdentifierToken (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *Identifier `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @IdentifierToken (IndentToken | OutdentToken)*))"`
	Operand     *Call       `parser:"@@"`
}

type Call struct {
	Left  *Select      `parser:"@@"`
	Right []*CallRight `parser:"@@*"`
}

type CallRight struct {
	Arguments *CallArguments `parser:"(IndentToken | OutdentToken)* '(' (IndentToken | OutdentToken | NewlineToken)* @@? (IndentToken | OutdentToken | NewlineToken)* ')'"`
	Select    *SelectRight   `parser:"| @@"`
}

type CallArguments struct {
	Head Expression     `parser:"@@"`
	Tail *CallArguments `parser:" ((IndentToken | OutdentToken | NewlineToken)* ',' (IndentToken | OutdentToken | NewlineToken)* @@)?"`
}

type Select struct {
	Left  Primary        `parser:"@@"`
	Right []*SelectRight `parser:"@@*"`
}

type SelectRight struct {
	Field *Identifier `parser:"(IndentToken | OutdentToken | NewlineToken)* '.' (IndentToken | OutdentToken | NewlineToken)* @IdentifierToken"`
}

// Single-token expressions

type Float struct {
	Value float64 `parser:"@FloatToken"`
}

func (*Float) primary() {}

type Identifier struct {
	Value string `parser:"@IdentifierToken"`
}

/*
 * This method is necessary for `Identifier` to implement the `participle.Capturable` interface,
 * allowing specific identifier tokens to be matched into the `Identifier` terminal struct.
 *
 * Note that because of a bug in Participle, `Identifier` can't be captured with `@@`:
 * https://github.com/alecthomas/participle/issues/140
 */
func (identifier *Identifier) Capture(values []string) error {
	identifier.Value = values[0]

	return nil
}

func (*Identifier) primary() {}

type Integer struct {
	Value int64 `parser:"@IntegerToken"`
}

func (*Integer) primary() {}

type String struct {
	Value string `parser:"@StringToken"`
}

func (*String) primary() {}
