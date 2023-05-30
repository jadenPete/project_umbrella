/*
 * The Runtime:
 *
 * The runtime is responsible for executing given bytecode concurrently. It does so by locating
 * dependencies among values and using `Graph.Evaluate` to continually compute leaf dependencies.
 *
 * The list of instructions is partitioned into "blocks" of instructions which can be executed
 * independently. Most instructions will occupy their own block. PUSH_ARG instructions are currently
 * the only exception; they'll occupy the same block as the VAL_FROM_CALL instruction to which they
 * correspond.
 */
package runtime

import (
	"fmt"
	"strings"

	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/parser"
)

type builtInFunctionId int

const (
	printFunctionId   builtInFunctionId = -1
	printlnFunctionId builtInFunctionId = -2
)

type builtInMethodId int

const (
	toStringMethodId builtInMethodId = -1
)

type valueDefinition struct {
	methods map[builtInMethodId]function
}

func generateStaticToString(result string) *builtInFunction {
	return &builtInFunction{
		evaluator: func(runtime *runtime, arguments ...value) value {
			return &stringValue{
				content: result,
			}
		},
	}
}

type value interface {
	definition() *valueDefinition
}

func newValueFromConstant(constant parser.Constant) value {
	switch constant.Type {
	case parser.StringConstant:
		return &stringValue{
			content: constant.Encoded,
		}
	}

	return nil
}

type function interface {
	evaluate(runtime *runtime, arguments ...value) value
}

type builtInFunction struct {
	evaluator func(runtime *runtime, arguments ...value) value
}

func (builtInFunction *builtInFunction) definition() *valueDefinition {
	return &valueDefinition{
		methods: map[builtInMethodId]function{
			toStringMethodId: generateStaticToString("(built-in function)"),
		},
	}
}

func (function *builtInFunction) evaluate(runtime *runtime, arguments ...value) value {
	return function.evaluator(runtime, arguments...)
}

func print(runtime *runtime, suffix string, arguments ...value) *unitValue {
	serialized := make([]string, 0, len(arguments))

	for _, argument := range arguments {
		if argumentString, ok := argument.
			definition().methods[toStringMethodId].
			evaluate(runtime, argument).(*stringValue); ok {
			serialized = append(serialized, argumentString.content)
		} else {
			panic("Runtime error: __to_str__ returned a non-string.")
		}
	}

	fmt.Print(strings.Join(serialized, " ") + suffix)

	return &unitValue{}
}

var builtInFunctions = map[builtInFunctionId]function{
	printFunctionId: &builtInFunction{
		evaluator: func(runtime *runtime, arguments ...value) value {
			return print(runtime, "", arguments...)
		},
	},

	printlnFunctionId: &builtInFunction{
		evaluator: func(runtime *runtime, arguments ...value) value {
			return print(runtime, "\n", arguments...)
		},
	},
}

type bytecodeFunction struct {
	valueInstructions *common.Graph[[]*parser.Instruction]
}

func (bytecodeFunction *bytecodeFunction) definition() *valueDefinition {
	return &valueDefinition{
		methods: map[builtInMethodId]function{
			toStringMethodId: generateStaticToString("(function)"),
		},
	}
}

// TODO: Make this concurrent
func (function *bytecodeFunction) evaluate(runtime *runtime, arguments ...value) value {
	nodes := function.valueInstructions.Nodes
	values := make(map[int]value)

	for i, argument := range arguments {
		values[i] = argument
	}

	function.valueInstructions.Evaluate(func(i int) {
		callArguments := make([]value, 0)

		for _, instruction := range nodes[i] {
			switch instruction.Type {
			case parser.PushArgumentInstruction:
				callArguments = append(callArguments, values[instruction.Arguments[0]])

			case parser.ValueFromCallInstruction:
				if function, ok := builtInFunctions[builtInFunctionId(instruction.Arguments[0])]; ok {
					values[i] = function.evaluate(runtime, callArguments...)

					return
				}

				panic(
					fmt.Sprintf(
						"Runtime error: %d is not a recognized built-in function ID",
						instruction.Arguments[0],
					),
				)

			case parser.ValueFromConstantInstruction:
				values[i] = runtime.constants[instruction.Arguments[0]]

				return

			case parser.ValueFromStructValueInstruction:
				panic("Runtime error: the VAL_FROM_STRUCT_VAL instruction isn't supported yet.")
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

func (stringValue *stringValue) definition() *valueDefinition {
	return &valueDefinition{
		methods: map[builtInMethodId]function{
			toStringMethodId: &builtInFunction{
				evaluator: func(runtime *runtime, arguments ...value) value {
					return arguments[0]
				},
			},
		},
	}
}

type unitValue struct{}

func (value *unitValue) definition() *valueDefinition {
	return &valueDefinition{
		methods: map[builtInMethodId]function{
			toStringMethodId: generateStaticToString("(unit)"),
		},
	}
}

type runtime struct {
	constants    []value
	rootFunction *bytecodeFunction
}

func newRuntime(bytecode *parser.Bytecode) *runtime {
	runtime := &runtime{
		constants: make([]value, 0),
		rootFunction: &bytecodeFunction{
			valueInstructions: common.NewGraph[[]*parser.Instruction](),
		},
	}

	for _, constant := range bytecode.Constants {
		runtime.constants = append(runtime.constants, newValueFromConstant(constant))
	}

	valueInstructions := runtime.rootFunction.valueInstructions

	pushArgumentInstructions := make([]*parser.Instruction, 0)

	for _, instruction := range bytecode.Instructions {
		switch instruction.Type {
		case parser.PushArgumentInstruction:
			pushArgumentInstructions = append(pushArgumentInstructions, instruction)
		case parser.ValueFromCallInstruction:
			pushArgumentInstructions = append(pushArgumentInstructions, instruction)
			valueInstructions.Nodes = append(valueInstructions.Nodes, pushArgumentInstructions)

			dependent := len(valueInstructions.Nodes) - 1

			if _, ok := builtInFunctions[builtInFunctionId(instruction.Arguments[0])]; !ok {
				valueInstructions.Edges[instruction.Arguments[0]] = []int{dependent}
			}

			for i, pushArgumentInstruction := range pushArgumentInstructions {
				if i < len(pushArgumentInstructions)-1 {
					dependency := pushArgumentInstruction.Arguments[0]

					valueInstructions.Edges[dependency] =
						append(valueInstructions.Edges[dependency], dependent)
				}
			}

			pushArgumentInstructions = make([]*parser.Instruction, 0)
		case parser.ValueFromConstantInstruction:
			valueInstructions.Nodes =
				append(valueInstructions.Nodes, []*parser.Instruction{instruction})
		case parser.ValueFromStructValueInstruction:
			valueInstructions.Nodes =
				append(valueInstructions.Nodes, []*parser.Instruction{instruction})

			dependent := len(valueInstructions.Nodes) - 1

			valueInstructions.Edges[dependent] = []int{instruction.Arguments[0]}
		}
	}

	return runtime
}

func (runtime *runtime) execute() {
	runtime.rootFunction.evaluate(runtime)
}

func ExecuteBytecode(bytecode *parser.Bytecode) {
	newRuntime(bytecode).execute()
}
