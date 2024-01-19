package bytecode_function

import (
	"reflect"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/built_in_definitions"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
	"project_umbrella/interpreter/runtime/value_util"
)

type BytecodeFunctionEvaluator struct {
	Constants       []value.Value
	ContainingScope *scope
	BlockGraph      *runtime.BytecodeFunctionBlockGraph
}

// TODO: Make this concurrent
func (evaluator *BytecodeFunctionEvaluator) Evaluator(
	runtime_ *runtime.Runtime,
	arguments ...value.Value,
) value.Value {
	firstValueID := 0

	if evaluator.ContainingScope != nil {
		firstValueID = evaluator.BlockGraph.FirstValueID
	}

	scope_ := &scope{
		parent:       evaluator.ContainingScope,
		firstValueID: firstValueID,
		values:       map[int]value.Value{},
	}

	for i, argument := range arguments {
		scope_.values[scope_.firstValueID+i] = argument
	}

	isAcyclic := evaluator.BlockGraph.Evaluate(
		func(consolidatedNode *common.ConsolidatedGraphNode[runtime.BytecodeFunctionBlock]) {
			functions := []*runtime.BytecodeFunctionBlockGraph{}
			instructionList := runtime.InstructionList(nil)

			for _, i := range consolidatedNode.Nodes() {
				switch node := evaluator.BlockGraph.ConsolidatedGraph.GetNode(i).(type) {
				case *runtime.BytecodeFunctionBlockGraph:
					if instructionList != nil {
						errors.RaiseError(runtime_errors.ValueCycle)
					}

					functions = append(functions, node)

				case runtime.InstructionList:
					if len(functions) > 0 {
						errors.RaiseError(runtime_errors.ValueCycle)
					}

					if instructionList != nil {
						errors.RaiseError(runtime_errors.ValueCycle)
					}

					instructionList = node
				}
			}

			if len(functions) > 0 {
				scope_.addFunctions(evaluator, functions)
			} else {
				scope_.addInstructionList(runtime_, evaluator, instructionList)
			}
		},
	)

	if !isAcyclic {
		errors.RaiseError(runtime_errors.ValueCycle)
	}

	if len(scope_.values) == 0 {
		errors.RaiseError(runtime_errors.EmptyFunctionBlockGraph)
	}

	lastValueID := 0

	for valueID := range scope_.values {
		if valueID > lastValueID {
			lastValueID = valueID
		}
	}

	return scope_.values[lastValueID]
}

type scope struct {
	parent       *scope
	firstValueID int
	values       map[int]value.Value
}

func (scope_ *scope) addFunctions(
	evaluator *BytecodeFunctionEvaluator,
	functions []*runtime.BytecodeFunctionBlockGraph,
) {
	for _, blockGraph := range functions {
		scope_.values[blockGraph.ValueID] = NewBytecodeFunction(
			blockGraph.ParameterCount,
			&BytecodeFunctionEvaluator{
				Constants:       evaluator.Constants,
				ContainingScope: scope_,
				BlockGraph:      blockGraph,
			},
		)
	}
}

func (scope_ *scope) addInstructionList(
	runtime_ *runtime.Runtime,
	evaluator *BytecodeFunctionEvaluator,
	instructionList runtime.InstructionList,
) {
	callArguments := []value.Value{}

	for _, element := range instructionList {
		switch element.Instruction.Type {
		case bytecode_generator.PushArgumentInstruction:
			callArguments =
				append(callArguments, scope_.getValue(element.Instruction.Arguments[0]))

		case bytecode_generator.ValueCopyInstruction:
			scope_.values[element.InstructionValueID] =
				scope_.getValue(element.Instruction.Arguments[0])

		case bytecode_generator.ValueFromCallInstruction:
			function_, ok :=
				scope_.getValue(element.Instruction.Arguments[0]).(*function.Function)

			if !ok {
				errors.RaiseError(runtime_errors.NonFunctionCalled)
			}

			scope_.values[element.InstructionValueID] =
				function_.Evaluate(runtime_, callArguments...)

		case bytecode_generator.ValueFromConstantInstruction:
			scope_.values[element.InstructionValueID] =
				evaluator.Constants[element.Instruction.Arguments[0]]

		case bytecode_generator.ValueFromStructValueInstruction:
			value_ := scope_.getValue(element.Instruction.Arguments[0])
			fieldNameConstant := evaluator.Constants[element.Instruction.Arguments[1]]
			fieldNameValue, ok := fieldNameConstant.(value_types.StringValue)

			if !ok {
				errors.RaiseError(runtime_errors.NonStringFieldName)
			}

			selectType := parser_types.SelectType(element.Instruction.Arguments[2])

			scope_.values[element.InstructionValueID] = value_util.LookupField(
				runtime_,
				value_,
				string(fieldNameValue),
				selectType,
			)
		}
	}
}

func (scope_ *scope) getValue(valueID int) value.Value {
	if builtInValue, ok :=
		built_in_definitions.BuiltInValues[built_in_declarations.BuiltInValueID(valueID)]; ok {
		return builtInValue
	}

	currentScope := scope_

	for currentScope.firstValueID > valueID {
		currentScope = currentScope.parent
	}

	return currentScope.values[valueID]
}

func NewBytecodeFunction(
	parameterCount int,
	evaluator *BytecodeFunctionEvaluator,
) *function.Function {
	name := "(function)"

	return &function.Function{
		FunctionEvaluator: evaluator,
		ArgumentValidator: function.NewFixedFunctionArgumentValidator(
			name,
			common.Repeat[reflect.Type](nil, parameterCount)...,
		),

		Name:  name,
		Type_: parser_types.NormalFunction,
	}
}
