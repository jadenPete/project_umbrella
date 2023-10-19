/*
 * Bytecode Translation:
 *
 * The bytecode translator is responsible for converting the AST generated by the parser into a list
 * of bytecode instructions and associated constant values (referred to as the constant list).
 * Values generated by instructions have an implicit value ID which increments and begins at 0.
 * Collectively, these values are referred to as the value list.
 *
 * Each value list belongs exclusively to a scope, which is created when a function is called.
 * Each scope contains a pointer to a parent scope, allowing functions to refer to values outside
 * themselves. Technically, the running program is modeled as a function whose scope is parentless.
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
 * - + (-2) (str, int, float)
 * - - (-3) (int, float)
 * - * (-4) (int, float)
 * - / (-5) (int, float)
 *
 * Functions are declared by pushing to a function stack, executing a series of instructions that
 * constitute the function, and popping from the function stack. Because each function's scope is
 * independent, the last value ID after popping from from the function stack equal to that before
 * pushing to the function stack.
 *
 * Before functions are executed, their arguments are appended to the value list. Additionally,
 * functions are accessible to themselves and others in their scope, regardless of order, because
 * functions are hoisted and appended to the value list before they are parsed.
 *
 * The following instructions (listed alongside their IDs and parameters) are available.
 *
 * PUSH_ARG (1) (VAL_ID):
 * 	Push VAL_ID to the argument stack to be passed to the next called function. Sequentially, the
 * 	argument stack can be thought of as being cleared on every function call.
 *
 * VAL_FROM_CALL (2) (VAL_ID):
 * 	Call the function referred to by VAL_ID with the arguments in the argument stack and push the
 * 	returned value to the value list.
 *
 * 	If VAL_ID doesn't refer to a function, the runtime will panic.
 *
 * VAL_FROM_CONST (3) (CONST_ID):
 * 	Retrieve the constant referred to by `CONST_ID` from the constant list and push it to the
 *  value list.
 *
 * VAL_FROM_STRUCT_VAL (4) (VAL_ID, FIELD_ID):
 * 	Retrieve the field identified by FIELD_ID from the struct instance referred to by VAL_ID,
 * 	pushing it to the value list.
 *
 * 	If VAL_ID doesn't refer to a struct instance, the runtime will panic.
 *
 * PUSH_FN (5):
 *  Push a function to the function stack.
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

	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/parser_errors"
	"project_umbrella/interpreter/parser"
	"project_umbrella/interpreter/parser/parser_types"
)

const checksumSize = 32

var builtInFields = map[string]*builtInField{
	"__to_str__": {-1, parser_types.NormalField},
	"+":          {-2, parser_types.InfixField},
	"-":          {-3, parser_types.InfixPrefixField},
	"*":          {-4, parser_types.InfixField},
	"/":          {-5, parser_types.InfixField},
}

var builtInValues = map[string]int{
	"print":   -1,
	"println": -2,
	"unit":    -3,
	"false":   -4,
	"true":    -5,
}

func sourceChecksum(fileContent string) [checksumSize]byte {
	return sha256.Sum256([]byte(fileContent))
}

type builtInField struct {
	id        int
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
		constantIDMap: make(map[Constant]int),
		instructions:  make([]*Instruction, 0),
		scopeStack: []*scope{
			{
				constantValueIDMap:   make(map[int]int),
				identifierValueIDMap: make(map[string]int),
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
	for _, subexpression := range expressionList.Children {
		if function, ok := subexpression.(*parser.Function); ok {
			translator.currentScope().identifierValueIDMap[function.Name.Value] =
				translator.currentScope().nextValueID

			translator.currentScope().nextValueID++
		}
	}

	isEmpty := true

	var returnValueID int

	addValueCopyInstruction := func() {
		translator.instructions = append(translator.instructions, &Instruction{
			Type:      ValueCopyInstruction,
			Arguments: []int{returnValueID},
		})
	}

	for _, subexpression := range expressionList.Children {
		isEmpty = false
		returnValueID = translator.valueIDForExpression(subexpression)

		/*
		 * Parsing an identifier doesn't append a value to the value list, but we need only do that
		 * if the identifier is an standalone expression
		 * (in which case it's possible that it refers to a value other than the last one).
		 */
		if _, ok := subexpression.(*parser.Identifier); ok {
			addValueCopyInstruction()
		}
	}

	if isEmpty {
		returnValueID = builtInValues["unit"]

		addValueCopyInstruction()
	}

	return returnValueID
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

	return valueID
}

func (translator *BytecodeTranslator) valueIDForFunction(function *parser.Function) int {
	translator.instructions = append(translator.instructions, &Instruction{
		Type: PushFunctionInstruction,
	})

	scope := &scope{
		constantValueIDMap:   make(map[int]int),
		identifierValueIDMap: make(map[string]int, len(function.Parameters)),
		nextValueID:          translator.currentScope().nextValueID,
	}

	for _, parameter := range function.Parameters {
		scope.identifierValueIDMap[parameter.Value] = scope.nextValueID
		scope.nextValueID++
	}

	translator.scopeStack = append(translator.scopeStack, scope)
	translator.valueIDForExpression(function.Value)
	translator.scopeStack = translator.scopeStack[:len(translator.scopeStack)-1]
	translator.instructions = append(translator.instructions, &Instruction{
		Type: PopFunctionInstruction,
	})

	return translator.currentScope().identifierValueIDMap[function.Name.Value]
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
			Arguments: []int{valueID, field.id},
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
	nextValueID          int
}
