package function

import (
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
)

type BuiltInFunctionEvaluator func(*runtime.Runtime, ...value.Value) value.Value

func (evaluator BuiltInFunctionEvaluator) Evaluator(
	runtime_ *runtime.Runtime,
	arguments ...value.Value,
) value.Value {
	return evaluator(runtime_, arguments...)
}

const BuiltInFunctionName = "(built-in function)"

func NewBuiltInFunction(
	argumentValidator FunctionArgumentValidator,
	evaluator BuiltInFunctionEvaluator,
	type_ *parser_types.FunctionType,
) *Function {
	return &Function{
		FunctionEvaluator: evaluator,
		ArgumentValidator: argumentValidator,
		Name:              BuiltInFunctionName,
		Type_:             type_,
	}
}
