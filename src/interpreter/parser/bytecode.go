/*
 * Bytecode Translation:
 *
 * The bytecode translator is responsible for converting the AST generated by the parser into a list
 * of bytecode instructions and associated constant values (referred to as the constant list).
 * Values generated by instructions have an implicit value ID which increments and begins at 0.
 * Collectively, these values are referred to as the value list.
 *
 * Built-in values are negative; the following are accessible.
 * - print (-1)
 * - println (-2)
 *
 * Likewise, built-in fields are negative; the following are accessible.
 * - __to_str__ (-1) (available on every type)
 * - + (-2) (available on the types str, int, and float)
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
 * 	Retrieve the constant `CONST_ID` from the constant list and push it to the value list.
 *
 * VAL_FROM_STRUCT_VAL (4) (VAL_ID, FIELD_ID):
 * 	Retrieve the field identified by FIELD_ID from the struct instance referred to by VAL_ID,
 * 	pushing it to the value list.
 *
 * 	If VAL_ID doesn't refer to a struct instance, the runtime will panic.
 */
package parser

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/ugorji/go/codec"

	"project_umbrella/interpreter/common"
)

const ChecksumSize = 32

var builtInFields = map[string]*BuiltInField{
	"__to_str__": {-1, false},
	"+":          {-2, true},
}

var builtinValues = map[string]int{
	"print":   -1,
	"println": -2,
}

func sourceChecksum(fileContent string) [ChecksumSize]byte {
	return sha256.Sum256([]byte(fileContent))
}

type BuiltInField struct {
	id            int
	isInfixMethod bool
}

type Bytecode struct {
	sourceChecksum [ChecksumSize]byte
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
		panic("Internal parser error: Couldn't encode the resulting bytecode.")
	}

	return output
}

type BytecodeTranslator struct {
	constantIDMap        map[Constant]int
	constantValueIDMap   map[int]int
	identifierValueIDMap map[string]int
	nextValueID          int
	instructions         []*Instruction
}

func NewBytecodeTranslator() *BytecodeTranslator {
	return &BytecodeTranslator{
		constantIDMap:        make(map[Constant]int),
		constantValueIDMap:   make(map[int]int),
		identifierValueIDMap: make(map[string]int),
		nextValueID:          0,
		instructions:         make([]*Instruction, 0),
	}
}

func (translator *BytecodeTranslator) generateBytecode(fileContent string) *Bytecode {
	bytecode := &Bytecode{
		sourceChecksum: sourceChecksum(fileContent),
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
			panic("Internal parser error: Nonexhaustive constant ID map.")
		}
	}

	return bytecode
}

func (translator *BytecodeTranslator) ExpressionToBytecode(
	expression Expression,
	fileContent string,
) *Bytecode {
	expressionList, ok := expression.(*ExpressionListExpression)

	if !ok {
		panic("Internal parser error: Expected an expression list, but got a different expression.")
	}

	for _, subexpression := range expressionList.Children {
		translator.valueIDForExpression(subexpression)
	}

	return translator.generateBytecode(fileContent)
}

func (translator *BytecodeTranslator) valueIDForAssignment(assignment *AssignmentExpression) int {
	valueID := translator.valueIDForExpression(assignment.Value)

	/*
	 * We check for value overloading in a separate pass because we don't want to leave
	 * `translator.identifierValueIDMap` in a bad state, in case the caller decides to recover.
	 */
	for _, nameExpression := range assignment.Names {
		if _, ok := translator.identifierValueIDMap[nameExpression.Content]; ok {
			panic("Reassigning to an already declared value is impossible.")
		}
	}

	for _, nameExpression := range assignment.Names {
		translator.identifierValueIDMap[nameExpression.Content] = valueID
	}

	return valueID
}

func (translator *BytecodeTranslator) valueIDForCall(call *CallExpression) int {
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

	result := translator.nextValueID

	translator.nextValueID++

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

	if result, ok := translator.constantValueIDMap[constantID]; ok {
		valueID = result
	} else {
		translator.instructions = append(translator.instructions, &Instruction{
			Type:      ValueFromConstantInstruction,
			Arguments: []int{constantID},
		})

		valueID = translator.nextValueID

		translator.nextValueID++
		translator.constantValueIDMap[constantID] = valueID
	}

	return valueID
}

func (translator *BytecodeTranslator) valueIDForExpression(expression Expression) int {
	var result int

	expression.Visit(&ExpressionVisitor{
		func(assignment *AssignmentExpression) {
			result = translator.valueIDForAssignment(assignment)
		},

		func(expressionList *ExpressionListExpression) {
			panic("Parser error: encountered a non-top level expression list.")
		},

		func(call *CallExpression) {
			result = translator.valueIDForCall(call)
		},

		func(float *FloatExpression) {
			var buffer bytes.Buffer

			if err := binary.Write(&buffer, binary.LittleEndian, float.Value); err != nil {
				panic(err)
			}

			result = translator.valueIDForConstant(Constant{
				Type:    FloatConstant,
				Encoded: buffer.String(),
			})
		},

		func(identifier *IdentifierExpression) {
			result = translator.valueIDForIdentifier(identifier)
		},

		func(integer *IntegerExpression) {
			var buffer bytes.Buffer

			if err := binary.Write(&buffer, binary.LittleEndian, integer.Value); err != nil {
				panic(err)
			}

			result = translator.valueIDForConstant(Constant{
				Type:    IntegerConstant,
				Encoded: buffer.String(),
			})
		},

		func(select_ *SelectExpression) {
			result = translator.valueIDForSelect(select_)
		},

		func(string_ *StringExpression) {
			result = translator.valueIDForConstant(Constant{
				Type:    StringConstant,
				Encoded: string_.Content,
			})
		},
	})

	return result
}

func (translator *BytecodeTranslator) valueIDForIdentifier(identifier *IdentifierExpression) int {
	if valueID, ok := translator.identifierValueIDMap[identifier.Content]; ok {
		return valueID
	}

	if valueID, ok := builtinValues[identifier.Content]; ok {
		return valueID
	}

	panic(fmt.Sprintf("Unknown value: `%s`", identifier.Content))
}

func (translator *BytecodeTranslator) valueIDForSelect(select_ *SelectExpression) int {
	valueID := translator.valueIDForExpression(select_.Value)

	fieldName := select_.Field.Content
	field, ok := builtInFields[fieldName]

	if !ok {
		panic(fmt.Sprintf("Unknown field: `%s`", fieldName))
	}

	if select_.IsInfix && !field.isInfixMethod {
		panic(fmt.Sprintf("`%s` cannot is not an infix method and cannot be called so.", fieldName))
	}

	translator.instructions = append(
		translator.instructions,
		&Instruction{
			Type:      ValueFromStructValueInstruction,
			Arguments: []int{valueID, field.id},
		},
	)

	result := translator.nextValueID

	translator.nextValueID++

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
)
