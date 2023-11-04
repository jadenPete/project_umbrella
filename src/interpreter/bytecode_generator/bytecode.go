/*
 * Bytecode Translation
 * --------------------
 *
 * The Bytecode Translator:
 *
 * The bytecode translator is responsible for converting the AST generated by the parser into a
 * sequence of bytecode instructions and associated constant values (referred to as the
 * constant list).
 *
 * Each type of instruction accepts a fixed number of arguments that customize the behaviour of the
 * instruction. Instructions manipulate internal data structures managed by the runtime, and
 * although they're ordered sequentially, they're designed to be executed in parallel.
 *
 * Those data structures are:
 * - The constant list (described above)
 * - The value list
 *
 * NOTE: Most instructions generate values to be appended to the value list, but not all
 * (e.g. `PUSH_ARG` doesn't generate a value).
 *
 * The Value List:
 *
 * Despite its name, the value list is a tree-based data structure composed of scopes, each
 * holding values and corresponding to a function. Because functions can be nested and called from
 * anywhere, it's important that they're able to reference:
 * - Values within themselves
 * - Themselves
 * - Values within their ancestor functions
 *
 * The value list's design enables this. Each scope contains a reference to its parent scope—
 * corresponding to the function in which the scope's function is defined—and stores, in order:
 * - The function's arguments
 * - The function's functions
 * - The function's values
 *
 * Whenever an instruction generates a value, it's said to append that value to the value list,
 * which is assigned a value ID (note that value IDs are a concept intrinsic to the runtime and
 * arent't actually stored in values themselves). Specifically, the scope's running value ID is
 * assigned to the value, and then that running value ID is incremented.
 *
 * Technically, the running program is modeled as a function whose scope is parentless and whose
 * value ID is initially 0.
 *
 * Function Hoisting:
 *
 * The only exception to this incrementing rule is functions; because functions are hoisted, they're
 * assigned value IDs first, although their scope's initial value ID is relative to where they're
 * defined.
 *
 * For example, if the first function is defined after four values, that function's value ID will be
 * nonetheless be 0, but the first value ID within it will follow the value IDs of the four values
 * defined before it.
 *
 * Function hoisting is helpful because it allows functions (and code outside of them) to reference
 * themselves regardless of mutual order or where they're defined. Furthermore, this model of
 * function hoisting is especially helpful because it enables functions to reference values before
 * themselves, permitting global state.
 *
 * Built-In Values and Fields:
 *
 * Built-in values are negative; the following are accessible.
 * - print (-1)
 * - println (-2)
 * - unit (-3)
 * - false (-4)
 * - true (-5)
 *
 * Likewise, built-in fields are negative; the following are accessible on the following types.
 * - __to_str__ (-1) (every type)
 * - == (-2) (every type)
 * - != (-3) (every type)
 *
 * - + (-4) (str, int, float)
 * - - (-5) (int, float)
 * - * (-6) (int, float)
 * - / (-7) (int, float)
 * - % (-8) (int, float)
 * - < (-9) (int, float)
 * - <= (-10) (int, float)
 * - > (-11) (int, float)
 * - >= (-12) (int, float)
 *
 * - ! (-13) (bool)
 * - && (-14) (bool)
 * - || (-15) (bool)
 *
 * Instructions:
 *
 * The following instructions (listed alongside their IDs and parameters) are available.
 *
 * PUSH_ARG (1) (VAL_ID):
 * 	Push `VAL_ID` to the argument stack to be passed to the next called function. Sequentially, the
 * 	argument stack can be thought of as being cleared on every function call.
 *
 * VAL_FROM_CALL (2) (VAL_ID):
 * 	Call the function referred to by `VAL_ID` with the arguments in the argument stack and push the
 * 	returned value to the value list.
 *
 * 	If `VAL_ID` doesn't refer to a function, the runtime will panic.
 *
 * VAL_FROM_CONST (3) (CONST_ID):
 * 	Retrieve the constant referred to by `CONST_ID` from the constant list and push it to the
 *  value list.
 *
 * VAL_FROM_STRUCT_VAL (4) (VAL_ID, FIELD_ID):
 * 	Retrieve the field identified by `FIELD_ID` from the struct instance referred to by `VAL_ID`,
 * 	pushing it to the value list.
 *
 * 	If `VAL_ID` doesn't refer to a struct instance, the runtime will panic.
 *
 * PUSH_FN (5) (ARG_COUNT):
 *  Push a function accepting `ARG_COUNT` arguments to the function stack.
 *
 * POP_FN (6):
 *  Pop the current function from the function stack.
 *
 * VAL_COPY (7) (VAL_ID):
 *  Retrieve the value referred to by `VAL_ID` from the value list and push it to the value list
 *  again.
 */
package bytecode_generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"github.com/ugorji/go/codec"

	"project_umbrella/interpreter/bytecode_generator/built_ins"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/parser_errors"
	"project_umbrella/interpreter/parser"
	"project_umbrella/interpreter/parser/parser_types"
)

const checksumSize = 32

var builtInFields = map[string]*builtInField{
	"__to_str__": {built_ins.ToStringMethodID, parser_types.NormalField},
	"==":         {built_ins.EqualsMethodID, parser_types.InfixField},
	"!=":         {built_ins.NotEqualsMethodID, parser_types.InfixField},

	"+":  {built_ins.PlusMethodID, parser_types.InfixField},
	"-":  {built_ins.MinusMethodID, parser_types.InfixPrefixField},
	"*":  {built_ins.TimesMethodID, parser_types.InfixField},
	"/":  {built_ins.OverMethodID, parser_types.InfixField},
	"%":  {built_ins.ModuloMethodID, parser_types.InfixField},
	"<":  {built_ins.LessThanMethodID, parser_types.InfixField},
	"<=": {built_ins.LessThanOrEqualToMethodID, parser_types.InfixField},
	">":  {built_ins.GreaterThanMethodID, parser_types.InfixField},
	">=": {built_ins.GreaterThanOrEqualToMethodID, parser_types.InfixField},

	"!":  {built_ins.NotMethodID, parser_types.PrefixField},
	"&&": {built_ins.AndMethodID, parser_types.InfixField},
	"||": {built_ins.OrMethodID, parser_types.InfixField},
}

var builtInValues = map[string]built_ins.BuiltInValueID{
	"print":       built_ins.PrintFunctionID,
	"println":     built_ins.PrintlnFunctionID,
	"unit":        built_ins.UnitValueID,
	"false":       built_ins.FalseValueID,
	"true":        built_ins.TrueValueID,
	"__if_else__": built_ins.IfElseFunctionID,
}

func sourceChecksum(fileContent string) [checksumSize]byte {
	return sha256.Sum256([]byte(fileContent))
}

type builtInField struct {
	id        built_ins.BuiltInFieldID
	fieldType parser_types.FieldType
}

type Bytecode struct {
	sourceChecksum [checksumSize]byte
	Constants      []Constant
	Instructions   []*Instruction
}

func DecodeBytecode(encoded []byte) *Bytecode {
	var handle codec.MsgpackHandle

	bytecode := &Bytecode{}

	codec.NewDecoderBytes(encoded, &handle).Decode(bytecode)

	return bytecode
}

func (bytecode *Bytecode) Encode() []byte {
	var handle codec.MsgpackHandle
	var output []byte

	if codec.NewEncoderBytes(&output, &handle).Encode(
		&map[string]interface{}{
			"constants":    bytecode.Constants,
			"instructions": bytecode.Instructions,
		},
	) != nil {
		errors.RaiseError(parser_errors.BytecodeEncodingFailed)
	}

	return output
}

type BytecodeTranslator struct {
	fileContent   string
	constantIDMap map[Constant]int
	instructions  []*Instruction
	scopeStack    []*scope
}

func NewBytecodeTranslator(fileContent string) *BytecodeTranslator {
	return &BytecodeTranslator{
		fileContent:   fileContent,
		constantIDMap: map[Constant]int{},
		instructions:  []*Instruction{},
		scopeStack: []*scope{
			{
				constantValueIDMap:   map[int]int{},
				identifierValueIDMap: map[string]int{},
				functionValueIDMap:   map[*parser.Function]int{},
				nextValueID:          0,
			},
		},
	}
}

func (translator *BytecodeTranslator) currentScope() *scope {
	return translator.scopeStack[len(translator.scopeStack)-1]
}

func (translator *BytecodeTranslator) ExpressionToBytecode(expression parser.Expression) *Bytecode {
	expressionList, ok := expression.(*parser.ExpressionList)

	if !ok {
		errors.RaiseError(parser_errors.InvalidRootExpression)
	}

	translator.valueIDForExpression(expressionList)

	return translator.generateBytecode()
}

func (translator *BytecodeTranslator) generateBytecode() *Bytecode {
	bytecode := &Bytecode{
		sourceChecksum: sourceChecksum(translator.fileContent),
		Constants:      make([]Constant, 0, len(translator.constantIDMap)),
		Instructions:   translator.instructions,
	}

	constantsSet := make([]bool, 0, len(translator.constantIDMap))

	for constant, i := range translator.constantIDMap {
		if i >= len(bytecode.Constants) {
			bytecode.Constants = common.Resize(bytecode.Constants, i+1)
		}

		if i >= len(constantsSet) {
			constantsSet = common.Resize(constantsSet, i+1)
		}

		bytecode.Constants[i] = constant

		constantsSet[i] = true
	}

	for _, isSet := range constantsSet {
		if !isSet {
			errors.RaiseError(parser_errors.NonexhaustiveConstantIDMap)
		}
	}

	return bytecode
}

func (translator *BytecodeTranslator) valueIDForAssignment(
	assignment *parser.Assignment,
) int {
	valueID := translator.valueIDForExpression(assignment.Value)

	/*
	 * We check for value overloading in a separate pass because we don't want to leave
	 * `translator.identifierValueIDMap` in a bad state, in case the caller decides to recover.
	 */
	for _, nameExpression := range assignment.Names {
		if _, ok := translator.valueIDForNonBuiltInIdentifierInScope(nameExpression); ok {
			errors.RaisePositionalError(
				&errors.PositionalError{
					Error:    parser_errors.ValueReassigned,
					Position: assignment.Position(),
				},

				translator.fileContent,
			)
		}
	}

	for _, nameExpression := range assignment.Names {
		translator.currentScope().identifierValueIDMap[nameExpression.Value] = valueID
	}

	return valueID
}

func (translator *BytecodeTranslator) valueIDForCall(call *parser.Call) int {
	functionValueID := translator.valueIDForExpression(call.Function)

	for _, argument := range call.Arguments {
		translator.instructions = append(translator.instructions, &Instruction{
			Type:      PushArgumentInstruction,
			Arguments: []int{translator.valueIDForExpression(argument)},
		})
	}

	translator.instructions = append(translator.instructions, &Instruction{
		Type:      ValueFromCallInstruction,
		Arguments: []int{functionValueID},
	})

	result := translator.currentScope().nextValueID

	translator.currentScope().nextValueID++

	return result
}

func (translator *BytecodeTranslator) valueIDForConstant(constant Constant) int {
	var constantID int

	if result, ok := translator.constantIDMap[constant]; ok {
		constantID = result
	} else {
		constantID = len(translator.constantIDMap)

		translator.constantIDMap[constant] = constantID
	}

	var valueID int

	if result, ok := translator.valueIDForNonBuiltInConstantInScope(constantID); ok {
		valueID = result
	} else {
		translator.instructions = append(translator.instructions, &Instruction{
			Type:      ValueFromConstantInstruction,
			Arguments: []int{constantID},
		})

		valueID = translator.currentScope().nextValueID

		translator.currentScope().nextValueID++
		translator.currentScope().constantValueIDMap[constantID] = valueID
	}

	return valueID
}

func (translator *BytecodeTranslator) valueIDForExpression(expression parser.Expression) int {
	var result int

	switch expression := expression.(type) {
	case *parser.Assignment:
		result = translator.valueIDForAssignment(expression)

	case *parser.Call:
		result = translator.valueIDForCall(expression)

	case *parser.ExpressionList:
		result = translator.valueIDForExpressionList(expression)

	case *parser.Float:
		var buffer bytes.Buffer

		if err := binary.Write(&buffer, binary.LittleEndian, expression.Value); err != nil {
			panic(err)
		}

		result = translator.valueIDForConstant(Constant{
			Type:    FloatConstant,
			Encoded: buffer.String(),
		})

	case *parser.Function:
		result = translator.valueIDForFunction(expression)

	case *parser.Identifier:
		result = translator.valueIDForIdentifier(expression)

	case *parser.Integer:
		var buffer bytes.Buffer

		if err := binary.Write(&buffer, binary.LittleEndian, expression.Value); err != nil {
			panic(err)
		}

		result = translator.valueIDForConstant(Constant{
			Type:    IntegerConstant,
			Encoded: buffer.String(),
		})

	case *parser.Select:
		result = translator.valueIDForSelect(expression)

	case *parser.String:
		result = translator.valueIDForConstant(Constant{
			Type:    StringConstant,
			Encoded: expression.Value,
		})
	}

	return result
}

func (translator *BytecodeTranslator) valueIDForExpressionList(
	expressionList *parser.ExpressionList,
) int {
	// Hoist declared functions
	stack := []parser.Expression{expressionList}

	for len(stack) > 0 {
		expression := stack[len(stack)-1]

		stack = stack[:len(stack)-1]

		if function, ok := expression.(*parser.Function); ok {
			functionValueID := translator.currentScope().nextValueID

			if function.Name != nil {
				translator.currentScope().identifierValueIDMap[function.Name.Value] =
					functionValueID
			}

			translator.currentScope().functionValueIDMap[function] = functionValueID
			translator.currentScope().nextValueID++
		} else {
			children := expression.Children()

			for i := len(children) - 1; i >= 0; i-- {
				stack = append(stack, children[i])
			}
		}

	}

	returnValueID := int(builtInValues["unit"])

	for _, subexpression := range expressionList.Children_ {
		returnValueID = translator.valueIDForExpression(subexpression)
	}

	/*
	 * The specification requires that functions return the value of their last statement. In
	 * implementation, they actually return the last value ID in their value list. In most cases,
	 * these are equivalent; however, the evaluation of some statements doesn't append a new value
	 * to the value list.
	 *
	 * For example, because identifiers are really lookups to the values to which they refer, their
	 * values are never unique and therefore never the last on the value list. To ensure functions
	 * return the value of their last statement, we end every function with a
	 * `VAL_COPY` instruction, which appends a copy of the value provided to it to the value list.
	 */
	translator.instructions = append(translator.instructions, &Instruction{
		Type:      ValueCopyInstruction,
		Arguments: []int{returnValueID},
	})

	return returnValueID
}

func (translator *BytecodeTranslator) valueIDForFunction(function *parser.Function) int {
	translator.instructions = append(translator.instructions, &Instruction{
		Type:      PushFunctionInstruction,
		Arguments: []int{len(function.Parameters)},
	})

	scope := &scope{
		constantValueIDMap:   map[int]int{},
		identifierValueIDMap: make(map[string]int, len(function.Parameters)),
		functionValueIDMap:   map[*parser.Function]int{},
		nextValueID:          translator.currentScope().nextValueID,
	}

	for _, parameter := range function.Parameters {
		scope.identifierValueIDMap[parameter.Value] = scope.nextValueID
		scope.nextValueID++
	}

	translator.scopeStack = append(translator.scopeStack, scope)
	translator.valueIDForExpression(function.Body)
	translator.scopeStack = translator.scopeStack[:len(translator.scopeStack)-1]
	translator.instructions = append(translator.instructions, &Instruction{
		Type: PopFunctionInstruction,
	})

	/*
	 * If the function is generated during the bytecode translation phase (as is the case in
	 * `valueIDForIf`), it will be anonymous and we can't look up its value ID in
	 * `translator.identifierValueIDMap`.
	 *
	 * Nonetheless, all functions, including those, are hoisted, requiring their value IDs to be
	 * recorded before they're evaluated by the bytecode translator. We do so in
	 * `translator.functionValueIDMap`.
	 */
	return translator.currentScope().functionValueIDMap[function]
}

func (translator *BytecodeTranslator) valueIDForIdentifier(identifier *parser.Identifier) int {
	if valueID, ok := translator.valueIDForNonBuiltInIdentifierInScope(identifier); ok {
		return valueID
	}

	valueID, ok := builtInValues[identifier.Value]

	if !ok {
		errors.RaisePositionalError(
			&errors.PositionalError{
				Error:    parser_errors.UnknownValue(identifier.Value),
				Position: identifier.Position(),
			},

			translator.fileContent,
		)
	}

	return int(valueID)
}

func (translator *BytecodeTranslator) valueIDForNonBuiltInConstantInScope(constantID int) (int, bool) {
	for i := len(translator.scopeStack) - 1; i >= 0; i-- {
		if valueID, ok := translator.scopeStack[i].constantValueIDMap[constantID]; ok {
			return valueID, true
		}
	}

	return 0, false
}

func (translator *BytecodeTranslator) valueIDForNonBuiltInIdentifierInScope(
	identifier *parser.Identifier,
) (int, bool) {
	for i := len(translator.scopeStack) - 1; i >= 0; i-- {
		if valueID, ok := translator.scopeStack[i].identifierValueIDMap[identifier.Value]; ok {
			return valueID, true
		}
	}

	return 0, false
}

func (translator *BytecodeTranslator) valueIDForSelect(select_ *parser.Select) int {
	valueID := translator.valueIDForExpression(select_.Value)

	fieldName := select_.Field.Value
	field, ok := builtInFields[fieldName]

	if !ok {
		errors.RaisePositionalError(
			&errors.PositionalError{
				Error:    parser_errors.UnknownField(fieldName),
				Position: select_.Field.Position(),
			},

			translator.fileContent,
		)
	}

	if !field.fieldType.CanSelectBy(select_.Type) {
		errors.RaisePositionalError(
			&errors.PositionalError{
				Error: parser_errors.MethodCalledImproperly(
					translator.fileContent[select_.Value.Position().Start:select_.Value.Position().End],
					fieldName,
					field.fieldType,
					select_.Type,
				),

				Position: select_.Field.Position(),
			},

			translator.fileContent,
		)
	}

	translator.instructions = append(
		translator.instructions,
		&Instruction{
			Type:      ValueFromStructValueInstruction,
			Arguments: []int{valueID, int(field.id)},
		},
	)

	result := translator.currentScope().nextValueID

	translator.currentScope().nextValueID++

	return result
}

type Constant struct {
	Type    ConstantType
	Encoded string
}

type ConstantType int

const (
	StringConstant ConstantType = iota + 1
	IntegerConstant
	FloatConstant
)

type Instruction struct {
	Type      InstructionType
	Arguments []int
}

type InstructionType int

const (
	PushArgumentInstruction InstructionType = iota + 1
	ValueFromCallInstruction
	ValueFromConstantInstruction
	ValueFromStructValueInstruction
	PushFunctionInstruction
	PopFunctionInstruction
	ValueCopyInstruction
)

type scope struct {
	constantValueIDMap   map[int]int
	identifierValueIDMap map[string]int
	functionValueIDMap   map[*parser.Function]int
	nextValueID          int
}
