package runtime_executor

import (
	"bytes"
	"encoding/binary"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/built_in_definitions"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/bytecode_function"
)

func executeRuntime(runtime_ *runtime.Runtime) {
	bytecode_function.NewBytecodeFunction(0, &bytecode_function.BytecodeFunctionEvaluator{
		ContainingScope: nil,
		BlockGraph:      runtime_.RootBlockGraph,
	}).Evaluate(runtime_)
}

func ExecuteBytecode(bytecode *bytecode_generator.Bytecode) {
	executeRuntime(newRuntime(bytecode))
}

func newRuntime(bytecode *bytecode_generator.Bytecode) *runtime.Runtime {
	type runtimeConstructorScope struct {
		nextValueID              int
		valueIDBlockMap          map[int]int
		functionCount            int
		functionsSeen            int
		pushArgumentInstructions []*bytecode_generator.Instruction
		blockGraph               *runtime.BytecodeFunctionBlockGraph
	}

	scopeStack := []*runtimeConstructorScope{
		{
			nextValueID:     0,
			valueIDBlockMap: map[int]int{},
			functionCount:   0,
			functionsSeen:   0,
			blockGraph: &runtime.BytecodeFunctionBlockGraph{
				Graph:          common.NewGraph[runtime.BytecodeFunctionBlock](),
				ValueID:        -1,
				FirstValueID:   0,
				ParameterCount: 0,
			},
		},
	}

	addDependencyForLatestBlock := func(dependencyValueID int) {
		if _, ok :=
			built_in_definitions.BuiltInValues[built_in_declarations.BuiltInValueID(dependencyValueID)]; ok {
			return
		}

		i := len(scopeStack) - 1

		for scopeStack[i].blockGraph.FirstValueID > dependencyValueID {
			i--
		}

		// If the dependency is an argument, it won't be present in the value ID-block index map
		if dependencyBlockID, ok := scopeStack[i].valueIDBlockMap[dependencyValueID]; ok {
			var dependentBlockID int

			/*
			 * The last block isn't necessarily the one most recently handled, since functions are
			 * hoisted before any subsequent blocks are added.
			 */
			if i < len(scopeStack)-1 {
				dependentBlockID = scopeStack[i].functionsSeen - 1
			} else {
				dependentBlockID = len(scopeStack[i].blockGraph.Nodes) - 1
			}

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

	addSingleValuedBlock := func(blockFromValueID func(int) runtime.BytecodeFunctionBlock) {
		valueID := currentScope().nextValueID

		currentScope().nextValueID++
		currentScope().blockGraph.Nodes =
			append(currentScope().blockGraph.Nodes, blockFromValueID(valueID))

		currentScope().valueIDBlockMap[valueID] = len(currentScope().blockGraph.Nodes) - 1
	}

	addValuedInstruction := func(instruction *bytecode_generator.Instruction) {
		addSingleValuedBlock(
			func(valueID int) runtime.BytecodeFunctionBlock {
				return runtime.InstructionList{
					&runtime.InstructionListElement{
						Instruction:        instruction,
						InstructionValueID: valueID,
					},
				}
			},
		)
	}

	// Hoist declared functions
	for _, instruction := range bytecode.Instructions {
		if instruction.Type == bytecode_generator.PushFunctionInstruction {
			newBlockGraph := &runtime.BytecodeFunctionBlockGraph{
				Graph:          common.NewGraph[runtime.BytecodeFunctionBlock](),
				ValueID:        0,
				FirstValueID:   0,
				ParameterCount: instruction.Arguments[0],
			}

			addSingleValuedBlock(
				func(int) runtime.BytecodeFunctionBlock {
					return newBlockGraph
				},
			)

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
				Nodes[currentScope().functionsSeen].(*runtime.BytecodeFunctionBlockGraph)

			scopeBlockGraph.ValueID = currentScope().blockGraph.FirstValueID +
				currentScope().blockGraph.ParameterCount +
				currentScope().functionsSeen

			scopeBlockGraph.FirstValueID = currentScope().nextValueID

			scopeFunctionCount := len(scopeBlockGraph.Nodes)
			scopeNextValueID :=
				scopeBlockGraph.FirstValueID + scopeBlockGraph.ParameterCount + scopeFunctionCount

			scopeValueIDBlockMap := make(map[int]int, len(scopeBlockGraph.Nodes))

			for i := 0; i < len(scopeBlockGraph.Nodes); i++ {
				functionValueID := scopeBlockGraph.FirstValueID + scopeBlockGraph.ParameterCount + i

				scopeValueIDBlockMap[functionValueID] = i
			}

			currentScope().functionsSeen++

			scopeStack = append(scopeStack, &runtimeConstructorScope{
				nextValueID:              scopeNextValueID,
				valueIDBlockMap:          scopeValueIDBlockMap,
				functionCount:            scopeFunctionCount,
				functionsSeen:            0,
				pushArgumentInstructions: []*bytecode_generator.Instruction{},
				blockGraph:               scopeBlockGraph,
			})

		case bytecode_generator.PopFunctionInstruction:
			scopeStack = scopeStack[:len(scopeStack)-1]

		case bytecode_generator.ValueCopyInstruction:
			addValuedInstruction(instruction)
			addDependencyForLatestBlock(instruction.Arguments[0])

		case bytecode_generator.ValueFromCallInstruction:
			instructionList :=
				make(runtime.InstructionList, 0, len(currentScope().pushArgumentInstructions)+1)

			for _, instruction := range currentScope().pushArgumentInstructions {
				instructionList = append(instructionList, &runtime.InstructionListElement{
					Instruction:        instruction,
					InstructionValueID: -1,
				})
			}

			instructionList = append(instructionList, &runtime.InstructionListElement{
				Instruction:        instruction,
				InstructionValueID: currentScope().nextValueID,
			})

			addSingleValuedBlock(
				func(int) runtime.BytecodeFunctionBlock {
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

	runtime := &runtime.Runtime{
		Constants:      make([]value.Value, 0, len(bytecode.Constants)),
		RootBlockGraph: scopeStack[0].blockGraph,
	}

	for _, constant := range bytecode.Constants {
		runtime.Constants = append(runtime.Constants, newValueFromConstant(constant))
	}

	return runtime
}

func newValueFromConstant(constant bytecode_generator.Constant) value.Value {
	switch constant.Type {
	case bytecode_generator.FloatConstant:
		var value float64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return value_types.FloatValue(value)

	case bytecode_generator.IntegerConstant:
		var value int64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return value_types.IntegerValue(value)

	case bytecode_generator.StringConstant:
		return value_types.StringValue(constant.Encoded)
	}

	return nil
}
