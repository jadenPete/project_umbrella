/*
 * The Runtime:
 *
 * The runtime is responsible for executing given bytecode concurrently. It does so by locating
 * dependencies among values and using `Graph.Evaluate` to continually evaluate leaf dependencies
 * and prune them from the graph.
 *
 * The list of instructions is first partitioned into "blocks" of instructions which can be executed
 * sequentially. Most instructions will occupy their own block. PUSH_ARG instructions are currently
 * the only exception; they occupy the same block as the VAL_FROM_CALL instruction to which they
 * correspond.
 *
 * Collectively, a block graph constitutes a function. In fact, a block need not be an
 * instruction list; it can also be a function itself, allowing functions to be nested and composed.
 * Dependencies between functions in different layers are modeled by "floating" them up, such that a
 * dependency between a block inside a function and another outside of it is indicated by a
 * dependency between the outer block and the function itself.
 *
 * One consequence of this is that before a function can be defined, every value on which it depends
 * need be evaluated.
 */
package runtime

import (
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/parser"
)

type bytecodeFunctionBlock interface {
	bytecodeFunctionBlock()
}

type bytecodeFunctionBlockGraph struct {
	graph *common.Graph[bytecodeFunctionBlock]
}

func (*bytecodeFunctionBlockGraph) bytecodeFunctionBlock()

type instructionList struct {
	instructions []*parser.Instruction
}

func (*instructionList) bytecodeFunctionBlock()

type runtime struct {
	constants      []value
	rootBlockGraph *bytecodeFunctionBlockGraph
}

func newRuntime(bytecode *parser.Bytecode) *runtime {
	type runtimeConstructorScope struct {
		firstValueID  int
		blockGraph    *bytecodeFunctionBlockGraph
		seenFunctions int
	}

	scopeStack := []*runtimeConstructorScope{
		{
			firstValueID: 0,
			blockGraph: &bytecodeFunctionBlockGraph{
				graph: common.NewGraph[bytecodeFunctionBlock](),
			},

			seenFunctions: 0,
		},
	}

	addDependencyForLatestValue := func(dependencyValueID int) {
		if _, ok := builtInValues[builtInValueID(dependencyValueID)]; ok {
			return
		}

		i := len(scopeStack) - 1

		for scopeStack[i].firstValueID > dependencyValueID {
			i--
		}

		relativeDependencyValueID := dependencyValueID - scopeStack[i].firstValueID
		relativeDependentValueID := len(scopeStack[i].blockGraph.graph.Nodes) - 1

		if relativeDependencyValueID != relativeDependentValueID {
			if dependents, ok := scopeStack[i].blockGraph.graph.Edges[relativeDependencyValueID]; ok {
				scopeStack[i].blockGraph.graph.Edges[relativeDependencyValueID] =
					append(dependents, relativeDependentValueID)
			} else {
				scopeStack[i].blockGraph.graph.Edges[relativeDependencyValueID] =
					[]int{relativeDependentValueID}
			}
		}
	}

	currentScope := func() *runtimeConstructorScope {
		return scopeStack[len(scopeStack)-1]
	}

	addInstructionListToCurrentFunction := func(instructions []*parser.Instruction) {
		currentScope().blockGraph.graph.Nodes = append(
			currentScope().blockGraph.graph.Nodes,
			&instructionList{instructions: instructions},
		)
	}

	// Hoist declared functions
	for _, instruction := range bytecode.Instructions {
		if instruction.Type == parser.PushFunctionInstruction {
			newBlockGraph := &bytecodeFunctionBlockGraph{
				graph: common.NewGraph[bytecodeFunctionBlock](),
			}

			currentScope().blockGraph.graph.Nodes =
				append(currentScope().blockGraph.graph.Nodes, newBlockGraph)

			scopeStack = append(scopeStack, &runtimeConstructorScope{
				blockGraph: newBlockGraph,
			})
		} else if instruction.Type == parser.PopFunctionInstruction {
			scopeStack = scopeStack[:len(scopeStack)-1]
		}
	}

	pushArgumentInstructions := make([]*parser.Instruction, 0)

	for _, instruction := range bytecode.Instructions {
		switch instruction.Type {
		case parser.PushArgumentInstruction:
			pushArgumentInstructions = append(pushArgumentInstructions, instruction)

		case parser.PushFunctionInstruction:
			scopeStack = append(scopeStack, &runtimeConstructorScope{
				firstValueID: currentScope().firstValueID + currentScope().seenFunctions + 1,
				blockGraph: currentScope().
					blockGraph.
					graph.
					Nodes[currentScope().seenFunctions].(*bytecodeFunctionBlockGraph),

				seenFunctions: 0,
			})

			scopeStack[len(scopeStack)-2].seenFunctions++

		case parser.PopFunctionInstruction:
			scopeStack = scopeStack[:len(scopeStack)-1]

		case parser.ValueCopyInstruction:
			addInstructionListToCurrentFunction([]*parser.Instruction{instruction})
			addDependencyForLatestValue(instruction.Arguments[0])

		case parser.ValueFromCallInstruction:
			pushArgumentInstructions = append(pushArgumentInstructions, instruction)

			addInstructionListToCurrentFunction(pushArgumentInstructions)
			addDependencyForLatestValue(instruction.Arguments[0])

			for i, pushArgumentInstruction := range pushArgumentInstructions {
				if i < len(pushArgumentInstructions)-1 {
					addDependencyForLatestValue(pushArgumentInstruction.Arguments[0])
				}
			}

			pushArgumentInstructions = make([]*parser.Instruction, 0)

		case parser.ValueFromConstantInstruction:
			addInstructionListToCurrentFunction([]*parser.Instruction{instruction})

		case parser.ValueFromStructValueInstruction:
			addInstructionListToCurrentFunction([]*parser.Instruction{instruction})
			addDependencyForLatestValue(instruction.Arguments[0])
		}
	}

	runtime := &runtime{
		constants:      make([]value, 0),
		rootBlockGraph: scopeStack[0].blockGraph,
	}

	for _, constant := range bytecode.Constants {
		runtime.constants = append(runtime.constants, newValueFromConstant(constant))
	}

	return runtime
}

func (runtime *runtime) execute() {
	(&bytecodeFunction{
		scope:      nil,
		blockGraph: runtime.rootBlockGraph,
	}).evaluate(runtime)
}

func ExecuteBytecode(bytecode *parser.Bytecode) {
	newRuntime(bytecode).execute()
}
