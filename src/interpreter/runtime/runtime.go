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
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/loader"
	"project_umbrella/interpreter/runtime/value"
)

type BytecodeFunctionBlock interface {
	BytecodeFunctionBlock()
}

type BytecodeFunctionBlockGraph struct {
	*common.Graph[BytecodeFunctionBlock]

	ValueID        int // Should be -1 if this is the root block graph
	FirstValueID   int
	ParameterCount int
}

func (*BytecodeFunctionBlockGraph) BytecodeFunctionBlock() {}

type InstructionList []*InstructionListElement
type InstructionListElement struct {
	Instruction        *bytecode_generator.Instruction
	InstructionValueID int // Should be -1 if the instruction is valueless
}

func (InstructionList) BytecodeFunctionBlock() {}

type Runtime struct {
	Constants      []value.Value
	RootBlockGraph *BytecodeFunctionBlockGraph
	LoaderChannel  *loader.LoaderChannel
}
