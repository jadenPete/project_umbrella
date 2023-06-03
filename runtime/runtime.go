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
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/parser"
)

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

			dependency := instruction.Arguments[0]

			valueInstructions.Edges[dependency] =
				append(valueInstructions.Edges[dependency], len(valueInstructions.Nodes)-1)
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
