package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
)

type builtInFieldID int

const (
	toStringMethodID builtInFieldID = -1
	plusMethodID     builtInFieldID = -2
	minusMethodID    builtInFieldID = -3
	timesMethodID    builtInFieldID = -4
	overMethodID     builtInFieldID = -5
)

type builtInValueID int

const (
	printFunctionID   builtInValueID = -1
	printlnFunctionID builtInValueID = -2
	unitValueID       builtInValueID = -3
	falseValueID      builtInValueID = -4
	trueValueID       builtInValueID = -5
)

type valueDefinition struct {
	fields map[builtInFieldID]value
}

type value interface {
	definition() *valueDefinition
}

func newValueFromConstant(constant bytecode_generator.Constant) value {
	switch constant.Type {
	case bytecode_generator.FloatConstant:
		var value float64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return floatValue{value}

	case bytecode_generator.IntegerConstant:
		var value int64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return integerValue{value}

	case bytecode_generator.StringConstant:
		return stringValue{constant.Encoded}
	}

	return nil
}

var builtInValues = map[builtInValueID]value{
	printFunctionID: newBuiltInFunction(
		newVariadicFunctionArgumentValidator("print", nil),
		func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "", arguments...)
		},
	),

	printlnFunctionID: newBuiltInFunction(
		newVariadicFunctionArgumentValidator("println", nil),
		func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "\n", arguments...)
		},
	),

	unitValueID: unitValue{},
	falseValueID: booleanValue{
		value: false,
	},

	trueValueID: booleanValue{
		value: true,
	},
}

func print(runtime_ *runtime, suffix string, arguments ...value) unitValue {
	serialized := make([]string, 0, len(arguments))

	for _, argument := range arguments {
		serialized = append(serialized, toString(runtime_, argument))
	}

	fmt.Print(strings.Join(serialized, " ") + suffix)

	return unitValue{}
}

func toString(runtime_ *runtime, value_ value) string {
	resultingValue, ok := value_.
		definition().fields[toStringMethodID].(*function).
		evaluate(runtime_).(stringValue)

	if !ok {
		errors.RaiseError(runtime_errors.ToStringMethodReturnedNonString)
	}

	return resultingValue.content
}

type booleanValue struct {
	value bool
}

func (value_ booleanValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: newToStringFunction(func() string {
				return fmt.Sprintf("%t", value_.value)
			}),
		},
	}
}

type functionArgumentValidator func(argumentTypes []reflect.Type) *errors.Error

func newFixedFunctionArgumentValidator(
	name string,
	parameterTypes ...reflect.Type,
) functionArgumentValidator {
	return functionArgumentValidator(
		func(argumentTypes []reflect.Type) *errors.Error {
			if len(argumentTypes) != len(parameterTypes) {
				return runtime_errors.IncorrectCallArgumentCount(
					strconv.Itoa(len(parameterTypes)),
					len(argumentTypes),
				)
			}

			for i, parameterType := range parameterTypes {
				if parameterType != nil && !argumentTypes[i].AssignableTo(parameterType) {
					return runtime_errors.IncorrectBuiltInFunctionArgumentType(name, i)
				}
			}

			return nil
		},
	)
}

func newIntersectionFunctionArgumentValidator(
	err func([]reflect.Type) *errors.Error,
	validators ...functionArgumentValidator,
) functionArgumentValidator {
	return functionArgumentValidator(
		func(argumentTypes []reflect.Type) *errors.Error {
			for _, validator := range validators {
				if err := validator(argumentTypes); err == nil {
					return nil
				}
			}

			return err(argumentTypes)
		},
	)
}

func newVariadicFunctionArgumentValidator(
	name string,
	parameterType reflect.Type,
) functionArgumentValidator {
	return functionArgumentValidator(
		func(argumentTypes []reflect.Type) *errors.Error {
			if parameterType == nil {
				return nil
			}

			for i, argumentType := range argumentTypes {
				if !argumentType.AssignableTo(parameterType) {
					return runtime_errors.IncorrectBuiltInFunctionArgumentType(name, i)
				}
			}

			return nil
		},
	)
}

type functionEvaluator interface {
	evaluator(*runtime, ...value) value
}

type builtInFunctionEvaluator func(*runtime, ...value) value

func (evaluator builtInFunctionEvaluator) evaluator(runtime_ *runtime, arguments ...value) value {
	return evaluator(runtime_, arguments...)
}

type bytecodeFunctionEvaluator struct {
	containingScope *scope
	blockGraph      *bytecodeFunctionBlockGraph
}

// TODO: Make this concurrent
func (evaluator *bytecodeFunctionEvaluator) evaluator(runtime_ *runtime, arguments ...value) value {
	firstValueID := 0

	if evaluator.containingScope != nil {
		firstValueID = evaluator.blockGraph.firstValueID
	}

	scope := &scope{
		parent:       evaluator.containingScope,
		firstValueID: firstValueID,
		values:       map[int]value{},
	}

	for i, argument := range arguments {
		scope.values[scope.firstValueID+i] = argument
	}

	isAcyclic := evaluator.blockGraph.Evaluate(func(i int) {
		callArguments := []value{}

		switch node := evaluator.blockGraph.Nodes[i].(type) {
		case *bytecodeFunctionBlockGraph:
			scope.values[node.firstValueID-1] = newBytecodeFunction(
				node.parameterCount,
				&bytecodeFunctionEvaluator{
					containingScope: scope,
					blockGraph:      node,
				},
			)

		case instructionList:
			for _, element := range node {
				switch element.instruction.Type {
				case bytecode_generator.PushArgumentInstruction:
					callArguments =
						append(callArguments, scope.getValue(element.instruction.Arguments[0]))

				case bytecode_generator.ValueCopyInstruction:
					scope.values[element.instructionValueID] =
						scope.getValue(element.instruction.Arguments[0])

				case bytecode_generator.ValueFromCallInstruction:
					function_, ok := scope.getValue(element.instruction.Arguments[0]).(*function)

					if !ok {
						errors.RaiseError(runtime_errors.NonFunctionCalled)
					}

					scope.values[element.instructionValueID] =
						function_.evaluate(runtime_, callArguments...)

				case bytecode_generator.ValueFromConstantInstruction:
					scope.values[element.instructionValueID] =
						runtime_.constants[element.instruction.Arguments[0]]

				case bytecode_generator.ValueFromStructValueInstruction:
					structValue := scope.getValue(element.instruction.Arguments[0])
					fieldID := builtInFieldID(element.instruction.Arguments[1])
					field, ok := structValue.definition().fields[fieldID]

					if !ok {
						errors.RaiseError(
							runtime_errors.UnrecognizedFieldID(
								toString(runtime_, structValue),
								int(fieldID),
							),
						)
					}

					scope.values[element.instructionValueID] = field
				}
			}
		}
	})

	if !isAcyclic {
		errors.RaiseError(runtime_errors.ValueCycle)
	}

	if len(scope.values) == 0 {
		errors.RaiseError(runtime_errors.EmptyFunctionBlockGraph)
	}

	lastValueID := 0

	for valueID := range scope.values {
		if valueID > lastValueID {
			lastValueID = valueID
		}
	}

	return scope.values[lastValueID]
}

type function struct {
	functionEvaluator

	argumentValidator functionArgumentValidator
	name              string
}

func (function_ *function) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: newToStringFunction(func() string {
				return function_.name
			}),
		},
	}
}

func (function_ *function) evaluate(runtime_ *runtime, arguments ...value) value {
	argumentTypes := make([]reflect.Type, 0, len(arguments))

	for _, argument := range arguments {
		argumentTypes = append(argumentTypes, reflect.TypeOf(argument))
	}

	if err := function_.argumentValidator(argumentTypes); err != nil {
		errors.RaiseError(err)
	}

	return function_.evaluator(runtime_, arguments...)
}

func newBuiltInFunction(
	argumentValidator functionArgumentValidator,
	evaluator func(*runtime, ...value) value,
) *function {
	return &function{
		functionEvaluator: builtInFunctionEvaluator(evaluator),
		argumentValidator: argumentValidator,
		name:              "(built-in function)",
	}
}

func newBytecodeFunction(
	parameterCount int,
	evaluator *bytecodeFunctionEvaluator,
) *function {
	name := "(function)"

	return &function{
		functionEvaluator: evaluator,
		argumentValidator: newFixedFunctionArgumentValidator(
			name,
			common.Repeat[reflect.Type](nil, parameterCount)...,
		),

		name: name,
	}
}

func newMinusMethod[Number value](
	numberType reflect.Type,
	negated func() Number,
	subtracted func(Number) Number,
) *function {
	return newBuiltInFunction(
		newIntersectionFunctionArgumentValidator(
			func(argumentTypes []reflect.Type) *errors.Error {
				return runtime_errors.IncorrectCallArgumentCount("0-1", len(argumentTypes))
			},

			newFixedFunctionArgumentValidator("-"),
			newFixedFunctionArgumentValidator("-", numberType),
		),

		func(_ *runtime, arguments ...value) value {
			if len(arguments) == 0 {
				return negated()
			}

			return subtracted(arguments[0].(Number))
		},
	)
}

func newToStringFunction(result func() string) *function {
	return newBuiltInFunction(
		newFixedFunctionArgumentValidator("__to_str__"),
		func(runtime_ *runtime, arguments ...value) value {
			return stringValue{result()}
		},
	)
}

type floatValue struct {
	value float64
}

func (value_ floatValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: newToStringFunction(func() string {
				return fmt.Sprintf("%g", value_.value)
			}),

			plusMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("+", reflect.TypeOf(floatValue{})),
				func(_ *runtime, arguments ...value) value {
					return floatValue{value_.value + arguments[0].(floatValue).value}
				},
			),

			minusMethodID: newMinusMethod(
				reflect.TypeOf(floatValue{}),
				func() floatValue {
					return floatValue{-value_.value}
				},

				func(rightHandSide floatValue) floatValue {
					return floatValue{value_.value - rightHandSide.value}
				},
			),

			timesMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("*", reflect.TypeOf(floatValue{})),
				func(_ *runtime, arguments ...value) value {
					return floatValue{value_.value * arguments[0].(floatValue).value}
				},
			),

			overMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("/", reflect.TypeOf(floatValue{})),
				func(_ *runtime, arguments ...value) value {
					rightHandSide := arguments[0].(floatValue).value

					if rightHandSide == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero("float"))
					}

					return floatValue{value_.value / rightHandSide}
				},
			),
		},
	}
}

type integerValue struct {
	value int64
}

func (value_ integerValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: newToStringFunction(func() string {
				return fmt.Sprintf("%d", value_.value)
			}),

			plusMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("+", reflect.TypeOf(integerValue{})),
				func(_ *runtime, arguments ...value) value {
					return integerValue{value_.value + arguments[0].(integerValue).value}
				},
			),

			minusMethodID: newMinusMethod(
				reflect.TypeOf(integerValue{}),
				func() integerValue {
					return integerValue{-value_.value}
				},

				func(rightHandSide integerValue) integerValue {
					return integerValue{value_.value - rightHandSide.value}
				},
			),

			timesMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("*", reflect.TypeOf(integerValue{})),
				func(_ *runtime, arguments ...value) value {
					return integerValue{value_.value * arguments[0].(integerValue).value}
				},
			),

			overMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("/", reflect.TypeOf(integerValue{})),
				func(_ *runtime, arguments ...value) value {
					rightHandSide := arguments[0].(integerValue).value

					if rightHandSide == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero("int"))
					}

					return integerValue{value_.value / rightHandSide}
				},
			),
		},
	}
}

type stringValue struct {
	content string
}

func (value_ stringValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: newToStringFunction(
				func() string {
					return value_.content
				},
			),

			plusMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("+", reflect.TypeOf(stringValue{})),
				func(_ *runtime, arguments ...value) value {
					return stringValue{value_.content + arguments[0].(stringValue).content}
				},
			),
		},
	}
}

type scope struct {
	parent       *scope
	firstValueID int
	values       map[int]value
}

func (scope_ *scope) getValue(valueID int) value {
	if builtInValue, ok := builtInValues[builtInValueID(valueID)]; ok {
		return builtInValue
	}

	currentScope := scope_

	for currentScope.firstValueID > valueID {
		currentScope = currentScope.parent
	}

	return currentScope.values[valueID]
}

type unitValue struct{}

func (value_ unitValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: newToStringFunction(func() string {
				return "(unit)"
			}),
		},
	}
}
