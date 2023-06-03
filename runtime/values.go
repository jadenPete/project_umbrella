package runtime

import (
	"fmt"
	"strings"

	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/parser"
)

type builtInFieldId int

const (
	toStringMethodId builtInFieldId = -1
	plusMethodId     builtInFieldId = -2
)

type builtInFunctionId int

const (
	printFunctionId   builtInFunctionId = -1
	printlnFunctionId builtInFunctionId = -2
)

type valueDefinition struct {
	fields map[builtInFieldId]value
}

type value interface {
	definition() *valueDefinition
}

func newValueFromConstant(constant parser.Constant) value {
	switch constant.Type {
	case parser.StringConstant:
		return stringValue{
			content: constant.Encoded,
		}
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

func generateStaticToString(result string) *builtInFunction {
	return &builtInFunction{
		argumentCount: 0,
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			return stringValue{
				content: result,
			}
		},
	}
}

func (function_ *builtInFunction) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldId]value{
			toStringMethodId: generateStaticToString("(built-in function)"),
		},
	}
}

func (function *builtInFunction) evaluate(runtime_ *runtime, arguments ...value) value {
	if !function.isVariadic && len(arguments) != function.argumentCount {
		panic(
			fmt.Sprintf(
				"Runtime error: A function that accepts %d arguments was called with %d",
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
		definition().fields[toStringMethodId].(function).
		evaluate(runtime_).(stringValue); ok {
		return resultingValue.content
	}

	panic("Runtime error: __to_str__ returned a non-string.")
}

var builtInFunctions = map[builtInFunctionId]function{
	printFunctionId: &builtInFunction{
		isVariadic: true,
		evaluator: func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "", arguments...)
		},
	},

	printlnFunctionId: &builtInFunction{
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
		fields: map[builtInFieldId]value{
			toStringMethodId: generateStaticToString("(function)"),
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

				if builtInFunction_, ok := builtInFunctions[builtInFunctionId(instruction.Arguments[0])]; ok {
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
				fieldID := builtInFieldId(instruction.Arguments[1])

				if field, ok := structValue.definition().fields[fieldID]; ok {
					values[i] = field

					return
				}

				panic(
					fmt.Sprintf(
						"Runtime error: %d is not a recognized field ID for the value %s",
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

type stringValue struct {
	content string
}

func (value_ stringValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldId]value{
			toStringMethodId: &builtInFunction{
				argumentCount: 0,
				evaluator: func(runtime_ *runtime, arguments ...value) value {
					return value_
				},
			},

			plusMethodId: &builtInFunction{
				argumentCount: 1,
				evaluator: func(runtime_ *runtime, arguments ...value) value {
					return stringValue{value_.content + arguments[0].(stringValue).content}
				},
			},
		},
	}
}

type unitValue struct{}

func (value_ *unitValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[builtInFieldId]value{
			toStringMethodId: generateStaticToString("(unit)"),
		},
	}
}
