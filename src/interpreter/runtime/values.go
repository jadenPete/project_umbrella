package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"project_umbrella/interpreter/bytecode_generator"
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
	printFunctionID: &builtInFunction{
		argumentValidator: variadicBuiltInFunction("print", nil),
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "", arguments...)
		},
	},

	printlnFunctionID: &builtInFunction{
		argumentValidator: variadicBuiltInFunction("println", nil),
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "\n", arguments...)
		},
	},

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
		definition().fields[toStringMethodID].(function).
		evaluate(runtime_).(stringValue)

	if !ok {
		errors.RaiseError(runtime_errors.ToStrReturnedNonString)
	}

	return resultingValue.content
}

type booleanValue struct {
	value bool
}

func (value_ booleanValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: generateToStringMethod(func() string {
				return fmt.Sprintf("%t", value_.value)
			}),
		},
	}
}

type function interface {
	value

	evaluate(runtime_ *runtime, arguments ...value) value
}

type builtInFunction struct {
	argumentValidator *builtInFunctionArgumentValidator
	evaluator         func(runtime_ *runtime, arguments ...value) value
}

func generateMinusMethod[Number value](
	numberType reflect.Type,
	negated func() Number,
	subtracted func(Number) Number,
) *builtInFunction {
	return &builtInFunction{
		argumentValidator: intersectionBuiltInFunction(
			func(argumentTypes []reflect.Type) *errors.Error {
				return runtime_errors.IncorrectCallArgumentCount("1-2", len(argumentTypes))
			},

			fixedBuiltInFunction("-", numberType),
			fixedBuiltInFunction("-", numberType, numberType),
		),

		evaluator: func(_ *runtime, arguments ...value) value {
			if len(arguments) == 0 {
				return negated()
			}

			return subtracted(arguments[0].(Number))
		},
	}
}

func generateToStringMethod(result func() string) *builtInFunction {
	return &builtInFunction{
		argumentValidator: fixedBuiltInFunction("__to_str__"),
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			return stringValue{result()}
		},
	}
}

func (function_ *builtInFunction) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: generateToStringMethod(func() string {
				return "(built-in function)"
			}),
		},
	}
}

func (function *builtInFunction) evaluate(runtime_ *runtime, arguments ...value) value {
	argumentTypes := make([]reflect.Type, 0, len(arguments))

	for _, argument := range arguments {
		argumentTypes = append(argumentTypes, reflect.TypeOf(argument))
	}

	if err := function.argumentValidator.validArguments(argumentTypes); err != nil {
		errors.RaiseError(err)
	}

	return function.evaluator(runtime_, arguments...)
}

type builtInFunctionArgumentValidator struct {
	validArguments func(argumentTypes []reflect.Type) *errors.Error
}

func fixedBuiltInFunction(
	name string,
	parameterTypes ...reflect.Type,
) *builtInFunctionArgumentValidator {
	return &builtInFunctionArgumentValidator{
		validArguments: func(argumentTypes []reflect.Type) *errors.Error {
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
	}
}

func intersectionBuiltInFunction(
	err func([]reflect.Type) *errors.Error,
	validators ...*builtInFunctionArgumentValidator,
) *builtInFunctionArgumentValidator {
	return &builtInFunctionArgumentValidator{
		validArguments: func(argumentTypes []reflect.Type) *errors.Error {
			for _, validator := range validators {
				if err := validator.validArguments(argumentTypes); err == nil {
					return nil
				}
			}

			return err(argumentTypes)
		},
	}
}

func variadicBuiltInFunction(
	name string,
	parameterType reflect.Type,
) *builtInFunctionArgumentValidator {
	return &builtInFunctionArgumentValidator{
		validArguments: func(argumentTypes []reflect.Type) *errors.Error {
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
	}
}

type bytecodeFunction struct {
	scope      *scope
	valueID    int
	blockGraph *bytecodeFunctionBlockGraph
}

func (function_ *bytecodeFunction) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: generateToStringMethod(func() string {
				return "(function)"
			}),
		},
	}
}

// TODO: Make this concurrent
func (bytecodeFunction_ *bytecodeFunction) evaluate(runtime_ *runtime, arguments ...value) value {
	firstValueID := 0

	if bytecodeFunction_.scope != nil {
		firstValueID = bytecodeFunction_.valueID + 1
	}

	scope := &scope{
		parent:       bytecodeFunction_.scope,
		firstValueID: firstValueID,
		values:       make(map[int]value),
	}

	for i, argument := range arguments {
		scope.values[scope.firstValueID+i] = argument
	}

	nodes := bytecodeFunction_.blockGraph.graph.Nodes

	bytecodeFunction_.blockGraph.graph.Evaluate(func(i int) {
		callArguments := make([]value, 0)

		switch node := nodes[i].(type) {
		case *bytecodeFunctionBlockGraph:
			scope.values[scope.firstValueID+i] = &bytecodeFunction{
				scope:      scope,
				valueID:    scope.firstValueID + i,
				blockGraph: node,
			}

		case *instructionList:
			for _, instruction := range node.instructions {
				switch instruction.Type {
				case bytecode_generator.PushArgumentInstruction:
					callArguments =
						append(callArguments, scope.getValue(instruction.Arguments[0]))

				case bytecode_generator.ValueCopyInstruction:
					scope.values[scope.firstValueID+i] =
						scope.getValue(instruction.Arguments[0])

				case bytecode_generator.ValueFromCallInstruction:
					scope.values[scope.firstValueID+i] = scope.
						getValue(instruction.Arguments[0]).(function).
						evaluate(runtime_, callArguments...)

				case bytecode_generator.ValueFromConstantInstruction:
					scope.values[scope.firstValueID+i] =
						runtime_.constants[instruction.Arguments[0]]

					return

				case bytecode_generator.ValueFromStructValueInstruction:
					structValue := scope.getValue(instruction.Arguments[0])
					fieldID := builtInFieldID(instruction.Arguments[1])

					if field, ok := structValue.definition().fields[fieldID]; ok {
						scope.values[scope.firstValueID+i] = field

						return
					}

					errors.RaiseError(
						runtime_errors.UnrecognizedFieldID(
							toString(runtime_, structValue),
							int(fieldID),
						),
					)
				}
			}
		}
	})

	if len(scope.values) != len(nodes) {
		errors.RaiseError(runtime_errors.ValueCycle)
	}

	if len(scope.values) == 0 {
		errors.RaiseError(runtime_errors.EmptyFunctionBlockGraph)
	}

	return scope.values[len(scope.values)-1]
}

type floatValue struct {
	value float64
}

func (value_ floatValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: generateToStringMethod(func() string {
				return fmt.Sprintf("%g", value_.value)
			}),

			plusMethodID: &builtInFunction{
				argumentValidator: fixedBuiltInFunction("+", reflect.TypeOf(floatValue{})),
				evaluator: func(_ *runtime, arguments ...value) value {
					return floatValue{value_.value + arguments[0].(floatValue).value}
				},
			},

			minusMethodID: generateMinusMethod(
				reflect.TypeOf(floatValue{}),
				func() floatValue {
					return floatValue{-value_.value}
				},

				func(rightHandSide floatValue) floatValue {
					return floatValue{value_.value - rightHandSide.value}
				},
			),

			timesMethodID: &builtInFunction{
				argumentValidator: fixedBuiltInFunction("*", reflect.TypeOf(floatValue{})),
				evaluator: func(_ *runtime, arguments ...value) value {
					return floatValue{value_.value * arguments[0].(floatValue).value}
				},
			},

			overMethodID: &builtInFunction{
				argumentValidator: fixedBuiltInFunction("/", reflect.TypeOf(floatValue{})),
				evaluator: func(_ *runtime, arguments ...value) value {
					rightHandSide := arguments[0].(floatValue).value

					if rightHandSide == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero("float"))
					}

					return floatValue{value_.value / rightHandSide}
				},
			},
		},
	}
}

type integerValue struct {
	value int64
}

func (value_ integerValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: generateToStringMethod(func() string {
				return fmt.Sprintf("%d", value_.value)
			}),

			plusMethodID: &builtInFunction{
				argumentValidator: fixedBuiltInFunction("+", reflect.TypeOf(integerValue{})),
				evaluator: func(_ *runtime, arguments ...value) value {
					return integerValue{value_.value + arguments[0].(integerValue).value}
				},
			},

			minusMethodID: generateMinusMethod(
				reflect.TypeOf(integerValue{}),
				func() integerValue {
					return integerValue{-value_.value}
				},

				func(rightHandSide integerValue) integerValue {
					return integerValue{value_.value - rightHandSide.value}
				},
			),

			timesMethodID: &builtInFunction{
				argumentValidator: fixedBuiltInFunction("*", reflect.TypeOf(integerValue{})),
				evaluator: func(_ *runtime, arguments ...value) value {
					return integerValue{value_.value * arguments[0].(integerValue).value}
				},
			},

			overMethodID: &builtInFunction{
				argumentValidator: fixedBuiltInFunction("/", reflect.TypeOf(integerValue{})),
				evaluator: func(_ *runtime, arguments ...value) value {
					rightHandSide := arguments[0].(integerValue).value

					if rightHandSide == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero("int"))
					}

					return integerValue{value_.value / rightHandSide}
				},
			},
		},
	}
}

type stringValue struct {
	content string
}

func (value_ stringValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: generateToStringMethod(
				func() string {
					return value_.content
				},
			),

			plusMethodID: &builtInFunction{
				argumentValidator: fixedBuiltInFunction("+", reflect.TypeOf(stringValue{})),
				evaluator: func(_ *runtime, arguments ...value) value {
					return stringValue{value_.content + arguments[0].(stringValue).content}
				},
			},
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
			toStringMethodID: generateToStringMethod(func() string {
				return "(unit)"
			}),
		},
	}
}
