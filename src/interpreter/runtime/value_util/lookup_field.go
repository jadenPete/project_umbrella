package value_util

import (
	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
)

func callUniversalMethod[ReturnValue value.Value](
	runtime_ *runtime.Runtime,
	value_ value.Value,
	methodName string,
	returnValueTypeName string,
	arguments ...value.Value,
) ReturnValue {
	method, ok :=
		LookupField(runtime_, value_, methodName, parser_types.NormalSelect).(*function.Function)

	if !ok {
		errors.RaiseError(runtime_errors.NonFunctionCalled)
	}

	result, ok := method.Evaluate(runtime_, arguments...).(ReturnValue)

	if !ok {
		errors.RaiseError(
			runtime_errors.UniversalMethodReturnedIncorrectValue(methodName, returnValueTypeName),
		)
	}

	return result
}

func CallEqualsMethod(
	runtime_ *runtime.Runtime,
	value1 value.Value,
	value2 value.Value,
) value_types.BooleanValue {
	return callUniversalMethod[value_types.BooleanValue](
		runtime_,
		value1,
		built_in_declarations.UniversalEqualsMethod.Name,
		"boolean",
		value2,
	)
}

func CallToStringMethod(runtime_ *runtime.Runtime, value_ value.Value) value_types.StringValue {
	return callUniversalMethod[value_types.StringValue](
		runtime_,
		value_,
		built_in_declarations.UniversalToStringMethod.Name,
		"str",
	)
}

func LookupField(
	runtime_ *runtime.Runtime,
	value_ value.Value,
	fieldName string,
	selectType parser_types.SelectType,
) value.Value {
	var universalMethodConstructors = map[string]func(value.Value) *function.Function{
		built_in_declarations.UniversalEqualsMethod.Name:    newEqualsMethod,
		built_in_declarations.UniversalNotEqualsMethod.Name: newNotEqualsMethod,
		built_in_declarations.UniversalToStringMethod.Name:  newToStringMethod,
	}

	var result value.Value

	if function_, ok := value_.(*function.Function); ok && function_.Type_.IsLookup {
		result = function_.Evaluate(runtime_, value_types.StringValue(fieldName))
	} else if field, ok := value_.Definition().Fields[fieldName]; ok {
		result = field
	} else if methodConstructor, ok := universalMethodConstructors[fieldName]; ok {
		result = methodConstructor(value_)
	} else {
		errors.RaiseError(runtime_errors.UnknownField(fieldName))
	}

	if function_, ok := result.(*function.Function); ok {
		if !function_.Type_.CanSelectBy(selectType) {
			errors.RaiseError(
				runtime_errors.MethodCalledImproperly(
					string(CallToStringMethod(runtime_, value_)),
					fieldName,
					function_.Type_,
					selectType,
				),
			)
		}
	}

	return result
}
