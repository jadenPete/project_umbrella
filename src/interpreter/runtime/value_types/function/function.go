package function

import (
	"reflect"
	"strconv"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
)

type Function struct {
	FunctionEvaluator

	ArgumentValidator FunctionArgumentValidator
	Name              string
	Type_             *parser_types.FunctionType
}

func (function *Function) Definition() *value.ValueDefinition {
	return &value.ValueDefinition{
		Fields: map[string]value.Value{},
	}
}

func (function *Function) Evaluate(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
	argumentTypes := make([]reflect.Type, 0, len(arguments))

	for _, argument := range arguments {
		argumentTypes = append(argumentTypes, reflect.TypeOf(argument))
	}

	if err := function.ArgumentValidator(argumentTypes); err != nil {
		errors.RaiseError(err)
	}

	return function.Evaluator(runtime_, arguments...)
}

type FunctionArgumentValidator func(argumentTypes []reflect.Type) *errors.Error

func NewFixedFunctionArgumentValidator(
	name string,
	parameterTypes ...reflect.Type,
) FunctionArgumentValidator {
	return func(argumentTypes []reflect.Type) *errors.Error {
		if len(argumentTypes) != len(parameterTypes) {
			return runtime_errors.IncorrectCallArgumentCount(
				strconv.Itoa(len(parameterTypes)),
				len(parameterTypes) != 1,
				len(argumentTypes),
			)
		}

		for i, parameterType := range parameterTypes {
			if parameterType != nil && !argumentTypes[i].AssignableTo(parameterType) {
				return runtime_errors.IncorrectBuiltInFunctionArgumentType(name, i)
			}
		}

		return nil
	}
}

func NewIntersectionFunctionArgumentValidator(
	err func([]reflect.Type) *errors.Error,
	validators ...FunctionArgumentValidator,
) FunctionArgumentValidator {
	return func(argumentTypes []reflect.Type) *errors.Error {
		for _, validator := range validators {
			if err := validator(argumentTypes); err == nil {
				return nil
			}
		}

		return err(argumentTypes)
	}
}

func NewVariadicFunctionArgumentValidator(
	name string,
	parameterType reflect.Type,
) FunctionArgumentValidator {
	return func(argumentTypes []reflect.Type) *errors.Error {
		if parameterType == nil {
			return nil
		}

		for i, argumentType := range argumentTypes {
			if !argumentType.AssignableTo(parameterType) {
				return runtime_errors.IncorrectBuiltInFunctionArgumentType(name, i)
			}
		}

		return nil
	}
}

type FunctionEvaluator interface {
	Evaluator(*runtime.Runtime, ...value.Value) value.Value
}
