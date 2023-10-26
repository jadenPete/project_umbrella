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
	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/bytecode_generator/built_ins"
	"project_umbrella/interpreter/common"
)

type bytecodeFunctionBlock interface {
	bytecodeFunctionBlock()
}

type bytecodeFunctionBlockGraph struct {
	*common.Graph[bytecodeFunctionBlock]

	firstValueID   int
	parameterCount int
}

func (*bytecodeFunctionBlockGraph) bytecodeFunctionBlock() {}

type instructionList []*instructionListElement
type instructionListElement struct {
	instruction        *bytecode_generator.Instruction
	instructionValueID int // Should be -1 if the instruction is valueless
}

func (instructionList) bytecodeFunctionBlock() {}

type runtime struct {
	constants      []value
	rootBlockGraph *bytecodeFunctionBlockGraph
}

func newRuntime(bytecode *bytecode_generator.Bytecode) *runtime {
	type runtimeConstructorScope struct {
		nextValueID              int
		valueIDBlockMap          map[int]int
		functionCount            int
		functionsSeen            int
		pushArgumentInstructions []*bytecode_generator.Instruction
		blockGraph               *bytecodeFunctionBlockGraph
	}

	scopeStack := []*runtimeConstructorScope{
		{
			nextValueID:     0,
			valueIDBlockMap: map[int]int{},
			functionCount:   0,
			functionsSeen:   0,
			blockGraph: &bytecodeFunctionBlockGraph{
				common.NewGraph[bytecodeFunctionBlock](),
				0,
				0,
			},
		},
	}

	addDependencyForLatestBlock := func(dependencyValueID int) {
		if _, ok := builtInValues[built_ins.BuiltInValueID(dependencyValueID)]; ok {
			return
		}

		i := len(scopeStack) - 1

		for scopeStack[i].blockGraph.firstValueID > dependencyValueID {
			i--
		}

		// If the dependency is an argument, it won't be present in the value ID-block index map
		if dependencyBlockID, ok := scopeStack[i].valueIDBlockMap[dependencyValueID]; ok {
			dependentBlockID := len(scopeStack[i].blockGraph.Nodes) - 1

			if dependencyBlockID != dependentBlockID {
				if dependents, ok := scopeStack[i].blockGraph.Edges[dependencyBlockID]; ok {
					scopeStack[i].blockGraph.Edges[dependencyBlockID] =
						append(dependents, dependentBlockID)
				} else {
					scopeStack[i].blockGraph.Edges[dependencyBlockID] = []int{dependentBlockID}
				}
			}
		}
	}

	currentScope := func() *runtimeConstructorScope {
		return scopeStack[len(scopeStack)-1]
	}

	addSingleValuedBlock :=
		func(blockFromValueID func(int) bytecodeFunctionBlock) bytecodeFunctionBlock {
			valueID := currentScope().nextValueID

			currentScope().nextValueID++

			block := blockFromValueID(valueID)

			currentScope().blockGraph.Nodes = append(currentScope().blockGraph.Nodes, block)
			currentScope().valueIDBlockMap[valueID] = len(currentScope().blockGraph.Nodes) - 1

			return block
		}

	addValuedInstruction := func(instruction *bytecode_generator.Instruction) {
		addSingleValuedBlock(
			func(valueID int) bytecodeFunctionBlock {
				return instructionList{
					&instructionListElement{
						instruction:        instruction,
						instructionValueID: valueID,
					},
				}
			},
		)
	}

	// Hoist declared functions
	for _, instruction := range bytecode.Instructions {
		if instruction.Type == bytecode_generator.PushFunctionInstruction {
			newBlockGraph := addSingleValuedBlock(
				func(valueID int) bytecodeFunctionBlock {
					return &bytecodeFunctionBlockGraph{
						common.NewGraph[bytecodeFunctionBlock](),
						0,
						instruction.Arguments[0],
					}
				},
			).(*bytecodeFunctionBlockGraph)

			currentScope().functionCount++

			scopeStack = append(scopeStack, &runtimeConstructorScope{
				nextValueID:              0,
				valueIDBlockMap:          map[int]int{},
				functionCount:            0,
				functionsSeen:            0,
				pushArgumentInstructions: []*bytecode_generator.Instruction{},
				blockGraph:               newBlockGraph,
			})
		} else if instruction.Type == bytecode_generator.PopFunctionInstruction {
			scopeStack = scopeStack[:len(scopeStack)-1]
		}
	}

	for _, instruction := range bytecode.Instructions {
		switch instruction.Type {
		case bytecode_generator.PushArgumentInstruction:
			currentScope().pushArgumentInstructions =
				append(currentScope().pushArgumentInstructions, instruction)

		case bytecode_generator.PushFunctionInstruction:
			scopeBlockGraph := currentScope().
				blockGraph.
				Nodes[currentScope().functionsSeen].(*bytecodeFunctionBlockGraph)

			scopeBlockGraph.firstValueID = currentScope().nextValueID
			scopeStack = append(scopeStack, &runtimeConstructorScope{
				nextValueID:              scopeBlockGraph.firstValueID,
				valueIDBlockMap:          map[int]int{},
				functionCount:            len(scopeBlockGraph.Nodes),
				functionsSeen:            0,
				pushArgumentInstructions: []*bytecode_generator.Instruction{},
				blockGraph:               scopeBlockGraph,
			})

			scopeStack[len(scopeStack)-2].functionsSeen++

		case bytecode_generator.PopFunctionInstruction:
			scopeStack = scopeStack[:len(scopeStack)-1]

		case bytecode_generator.ValueCopyInstruction:
			addValuedInstruction(instruction)
			addDependencyForLatestBlock(instruction.Arguments[0])

		case bytecode_generator.ValueFromCallInstruction:
			instructionList :=
				make(instructionList, 0, len(currentScope().pushArgumentInstructions)+1)

			for _, instruction := range currentScope().pushArgumentInstructions {
				instructionList = append(instructionList, &instructionListElement{
					instruction:        instruction,
					instructionValueID: -1,
				})
			}

			instructionList = append(instructionList, &instructionListElement{
				instruction:        instruction,
				instructionValueID: currentScope().nextValueID,
			})

			addSingleValuedBlock(
				func(int) bytecodeFunctionBlock {
					return instructionList
				},
			)

			addDependencyForLatestBlock(instruction.Arguments[0])

			for _, pushArgumentInstruction := range currentScope().pushArgumentInstructions {
				addDependencyForLatestBlock(pushArgumentInstruction.Arguments[0])
			}

			currentScope().pushArgumentInstructions = []*bytecode_generator.Instruction{}

		case bytecode_generator.ValueFromConstantInstruction:
			addValuedInstruction(instruction)

		case bytecode_generator.ValueFromStructValueInstruction:
			addValuedInstruction(instruction)
			addDependencyForLatestBlock(instruction.Arguments[0])
		}
	}

	runtime := &runtime{
		constants:      make([]value, 0, len(bytecode.Constants)),
		rootBlockGraph: scopeStack[0].blockGraph,
	}

	for _, constant := range bytecode.Constants {
		runtime.constants = append(runtime.constants, newValueFromConstant(constant))
	}

	return runtime
}

func (runtime *runtime) execute() {
	newBytecodeFunction(0, &bytecodeFunctionEvaluator{
		containingScope: nil,
		blockGraph:      runtime.rootBlockGraph,
	}).evaluate(runtime)
}

func ExecuteBytecode(bytecode *bytecode_generator.Bytecode) {
	newRuntime(bytecode).execute()
}
