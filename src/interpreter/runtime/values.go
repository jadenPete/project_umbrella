package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser"
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
)

type valueDefinition struct {
	fields map[builtInFieldID]value
}

type value interface {
	definition() *valueDefinition
}

func newValueFromConstant(constant parser.Constant) value {
	switch constant.Type {
	case parser.FloatConstant:
		var value float64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return floatValue{value}

	case parser.IntegerConstant:
		var value int64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return integerValue{value}

	case parser.StringConstant:
		return stringValue{constant.Encoded}
	}

	return nil
}

type function interface {
	value

	evaluate(runtime_ *runtime, arguments ...value) value
}

type builtInFunction struct {
	argumentCount int
	isVariadic    bool
	evaluator     func(runtime_ *runtime, arguments ...value) value
}

func generateInfixMethod[T value](
	operation func(rightHandSide T) T,
	typeName string,
	methodName string,
) *builtInFunction {
	return &builtInFunction{
		argumentCount: 1,
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			argument, ok := arguments[0].(T)

			if !ok {
				errors.RaiseError(
					runtime_errors.IncorrectBuiltInInfixMethodArgumentType(typeName, methodName),
				)
			}

			return operation(argument)
		},
	}
}

func generateToStringMethod(result func() string) *builtInFunction {
	return &builtInFunction{
		argumentCount: 0,
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
	if !function.isVariadic && len(arguments) != function.argumentCount {
		errors.RaiseError(
			runtime_errors.IncorrectCallArgumentCount(function.argumentCount, len(arguments)),
		)
	}

	return function.evaluator(runtime_, arguments...)
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

var builtInValues = map[builtInValueID]value{
	printFunctionID: &builtInFunction{
		isVariadic: true,
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "", arguments...)
		},
	},

	printlnFunctionID: &builtInFunction{
		isVariadic: true,
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "\n", arguments...)
		},
	},

	unitValueID: unitValue{},
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

		nodes[i].Visit(&bytecodeFunctionBlockVisitor{
			VisitBytecodeFunctionBlockGraph: func(blockGraph *bytecodeFunctionBlockGraph) {
				scope.values[scope.firstValueID+i] = &bytecodeFunction{
					scope:      scope,
					valueID:    scope.firstValueID + i,
					blockGraph: blockGraph,
				}
			},

			VisitInstructionList: func(instructionList_ *instructionList) {
				for _, instruction := range instructionList_.instructions {
					switch instruction.Type {
					case parser.PushArgumentInstruction:
						callArguments =
							append(callArguments, scope.getValue(instruction.Arguments[0]))

					case parser.ValueCopyInstruction:
						scope.values[scope.firstValueID+i] =
							scope.getValue(instruction.Arguments[0])

					case parser.ValueFromCallInstruction:
						scope.values[scope.firstValueID+i] = scope.
							getValue(instruction.Arguments[0]).(function).
							evaluate(runtime_, callArguments...)

					case parser.ValueFromConstantInstruction:
						scope.values[scope.firstValueID+i] =
							runtime_.constants[instruction.Arguments[0]]

						return

					case parser.ValueFromStructValueInstruction:
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
			},
		})
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

			plusMethodID: generateInfixMethod(
				func(rightHandSide floatValue) floatValue {
					return floatValue{value_.value + rightHandSide.value}
				},

				"float",
				"+",
			),

			minusMethodID: generateInfixMethod(
				func(rightHandSide floatValue) floatValue {
					return floatValue{value_.value - rightHandSide.value}
				},

				"float",
				"-",
			),

			timesMethodID: generateInfixMethod(
				func(rightHandSide floatValue) floatValue {
					return floatValue{value_.value * rightHandSide.value}
				},

				"float",
				"*",
			),

			overMethodID: generateInfixMethod(
				func(rightHandSide floatValue) floatValue {
					if rightHandSide.value == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero("float"))
					}

					return floatValue{value_.value / rightHandSide.value}
				},

				"float",
				"/",
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
			toStringMethodID: generateToStringMethod(func() string {
				return fmt.Sprintf("%d", value_.value)
			}),

			plusMethodID: generateInfixMethod(
				func(rightHandSide integerValue) integerValue {
					return integerValue{value_.value + rightHandSide.value}
				},

				"int",
				"+",
			),

			minusMethodID: generateInfixMethod(
				func(rightHandSide integerValue) integerValue {
					return integerValue{value_.value - rightHandSide.value}
				},

				"int",
				"-",
			),

			timesMethodID: generateInfixMethod(
				func(rightHandSide integerValue) integerValue {
					return integerValue{value_.value * rightHandSide.value}
				},

				"int",
				"*",
			),

			overMethodID: generateInfixMethod(
				func(rightHandSide integerValue) integerValue {
					if rightHandSide.value == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero("int"))
					}

					return integerValue{value_.value / rightHandSide.value}
				},

				"int",
				"/",
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
			toStringMethodID: &builtInFunction{
				argumentCount: 0,
				evaluator: func(runtime_ *runtime, arguments ...value) value {
					return value_
				},
			},

			plusMethodID: generateInfixMethod(
				func(rightHandSide stringValue) stringValue {
					return stringValue{value_.content + rightHandSide.content}
				},

				"str",
				"+",
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
			toStringMethodID: generateToStringMethod(func() string {
				return "(unit)"
			}),
		},
	}
}
