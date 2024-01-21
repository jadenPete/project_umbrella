package runtime_executor

import (
	"bytes"
	"encoding/binary"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/loader"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/built_in_definitions"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/bytecode_function"
)

func ExecuteBytecode(
	bytecode *bytecode_generator.Bytecode,
	loaderChannel *loader.LoaderChannel,
) value.Value {
	constants := make([]value.Value, 0, len(bytecode.Constants))

	for _, constant := range bytecode.Constants {
		constants = append(constants, newValueFromConstant(constant))
	}

	return bytecode_function.
		NewBytecodeFunction(0, &bytecode_function.BytecodeFunctionEvaluator{
			Constants:       constants,
			ContainingScope: nil,
			BlockGraph:      newBlockGraphFromBytecode(bytecode),
		}).
		Evaluate(&runtime.Runtime{
			LoaderChannel: loaderChannel,
		})
}

func newBlockGraphFromBytecode(
	bytecode *bytecode_generator.Bytecode,
) *runtime.BytecodeFunctionBlockGraph {
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
				ConsolidatedGraph: common.NewConsolidatedGraph[runtime.BytecodeFunctionBlock](),
				ValueID:           -1,
				FirstValueID:      0,
				ParameterCount:    0,
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
				dependentBlockID = scopeStack[i].blockGraph.Length() - 1
			}

			if dependencyBlockID != dependentBlockID {
				scopeStack[i].blockGraph.AddEdge(dependencyBlockID, dependentBlockID)
			}
		}
	}

	currentScope := func() *runtimeConstructorScope {
		return scopeStack[len(scopeStack)-1]
	}

	addSingleValuedBlock := func(blockFromValueID func(int) runtime.BytecodeFunctionBlock) {
		valueID := currentScope().nextValueID

		currentScope().nextValueID++
		currentScope().valueIDBlockMap[valueID] =
			currentScope().blockGraph.AddNode(blockFromValueID(valueID))
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
				ConsolidatedGraph: common.NewConsolidatedGraph[runtime.BytecodeFunctionBlock](),
				ValueID:           0,
				FirstValueID:      0,
				ParameterCount:    instruction.Arguments[0],
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
				GetNode(currentScope().functionsSeen).(*runtime.BytecodeFunctionBlockGraph)

			scopeBlockGraph.ValueID = currentScope().blockGraph.FirstValueID +
				currentScope().blockGraph.ParameterCount +
				currentScope().functionsSeen

			scopeBlockGraph.FirstValueID = currentScope().nextValueID

			scopeFunctionCount := scopeBlockGraph.Length()
			scopeNextValueID :=
				scopeBlockGraph.FirstValueID + scopeBlockGraph.ParameterCount + scopeFunctionCount

			scopeValueIDBlockMap := make(map[int]int, scopeBlockGraph.Length())

			for i := 0; i < scopeBlockGraph.Length(); i++ {
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

	result := scopeStack[0].blockGraph
	result.ConsolidateRecursively()

	return result
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
