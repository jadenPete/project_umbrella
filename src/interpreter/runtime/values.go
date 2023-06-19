package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"project_umbrella/interpreter/common"
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

type builtInFunctionID int

const (
	printFunctionID   builtInFunctionID = -1
	printlnFunctionID builtInFunctionID = -2
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
				panic(
					fmt.Sprintf(
						"Runtime error: Expected the right-hand side of %[1]s#%[2]s to be of type %[1]s.",
						typeName,
						methodName,
					),
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
		panic(
			fmt.Sprintf(
				"Runtime error: A function that accepts %d arguments was called with %d arguments.",
				len(arguments),
				function.argumentCount,
			),
		)
	}

	return function.evaluator(runtime_, arguments...)
}

func print(runtime_ *runtime, suffix string, arguments ...value) *unitValue {
	serialized := make([]string, 0, len(arguments))

	for _, argument := range arguments {
		serialized = append(serialized, toString(runtime_, argument))
	}

	fmt.Print(strings.Join(serialized, " ") + suffix)

	return &unitValue{}
}

func toString(runtime_ *runtime, value_ value) string {
	if resultingValue, ok := value_.
		definition().fields[toStringMethodID].(function).
		evaluate(runtime_).(stringValue); ok {
		return resultingValue.content
	}

	panic("Runtime error: __to_str__ returned a non-string.")
}

var builtInFunctions = map[builtInFunctionID]function{
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
}

type bytecodeFunction struct {
	valueInstructions *common.Graph[[]*parser.Instruction]
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
	nodes := bytecodeFunction_.valueInstructions.Nodes
	values := make(map[int]value)

	for i, argument := range arguments {
		values[i] = argument
	}

	bytecodeFunction_.valueInstructions.Evaluate(func(i int) {
		callArguments := make([]value, 0)

		for _, instruction := range nodes[i] {
			switch instruction.Type {
			case parser.PushArgumentInstruction:
				callArguments = append(callArguments, values[instruction.Arguments[0]])

			case parser.ValueFromCallInstruction:
				var function_ function

				if builtInFunction_, ok := builtInFunctions[builtInFunctionID(instruction.Arguments[0])]; ok {
					function_ = builtInFunction_
				} else {
					function_ = values[instruction.Arguments[0]].(function)
				}

				values[i] = function_.evaluate(runtime_, callArguments...)

			case parser.ValueFromConstantInstruction:
				values[i] = runtime_.constants[instruction.Arguments[0]]

				return

			case parser.ValueFromStructValueInstruction:
				structValue := values[instruction.Arguments[0]]
				fieldID := builtInFieldID(instruction.Arguments[1])

				if field, ok := structValue.definition().fields[fieldID]; ok {
					values[i] = field

					return
				}

				panic(
					fmt.Sprintf(
						"Runtime error: %d is not a recognized field ID for the value `%s`.",
						fieldID,
						toString(runtime_, structValue),
					),
				)
			}
		}
	})

	if len(values) != len(nodes) {
		panic("Runtime error: A cycle was found in the value tree.")
	}

	return values[len(nodes)-1]
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
						panic(
							"Runtime error: Expected the right-hand side of float#/ to be nonzero.",
						)
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
						panic(
							"Runtime error: Expected the right-hand side of int#/ to be nonzero.",
						)
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

type unitValue struct{}

func (value_ *unitValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldID]value{
			toStringMethodID: generateToStringMethod(func() string {
				return "(unit)"
			}),
		},
	}
}
