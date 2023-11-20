package built_in_definitions

import (
	"fmt"
	"reflect"
	"strings"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
	"project_umbrella/interpreter/runtime/value_util"
)

var BuiltInValues = map[built_in_declarations.BuiltInValueID]value.Value{
	built_in_declarations.PrintFunctionID: function.NewBuiltInFunction(
		function.NewVariadicFunctionArgumentValidator("print", nil),
		func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
			return print(runtime_, "", arguments...)
		},

		parser_types.NormalFunction,
	),

	built_in_declarations.PrintlnFunctionID: function.NewBuiltInFunction(
		function.NewVariadicFunctionArgumentValidator("println", nil),
		func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
			return print(runtime_, "\n", arguments...)
		},

		parser_types.NormalFunction,
	),

	built_in_declarations.UnitValueID:  value_types.UnitValue{},
	built_in_declarations.FalseValueID: value_types.BooleanValue(false),
	built_in_declarations.TrueValueID:  value_types.BooleanValue(true),
	built_in_declarations.IfElseFunctionID: function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			"__if_else__",
			reflect.TypeOf(*new(value_types.BooleanValue)),
			reflect.TypeOf(&function.Function{}),
			reflect.TypeOf(&function.Function{}),
		),

		ifElse,
		parser_types.NormalFunction,
	),

	built_in_declarations.TupleFunctionID: function.NewBuiltInFunction(
		function.NewVariadicFunctionArgumentValidator("__tuple__", nil),
		tuple,
		parser_types.NormalFunction,
	),

	built_in_declarations.StructFunctionID: function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			"__struct__",
			reflect.TypeOf(value_types.StringValue{}),
			reflect.TypeOf(&function.Function{}),
			reflect.TypeOf(&function.Function{}),
			reflect.TypeOf(value_types.TupleValue{}),
		),

		struct_,
		parser_types.NormalFunction,
	),
}

func builtInStructFields(
	structName string,
	structConstructor *function.Function,
	structArgumentNames []string,
	structArgumentValues []value.Value,
) map[string]value.Value {
	equalsMethodEvaluator :=
		func(runtime_ *runtime.Runtime, arguments ...value.Value) value_types.BooleanValue {
			return structEquals(
				runtime_,
				structConstructor,
				structArgumentNames,
				structArgumentValues,
				arguments...,
			)
		}

	return map[string]value.Value{
		built_in_declarations.ToStringMethod.Name: function.NewBuiltInFunction(
			function.NewFixedFunctionArgumentValidator(built_in_declarations.ToStringMethod.Name),
			func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
				argumentsAsStrings := make([]string, 0, len(structArgumentValues))

				for _, argument := range structArgumentValues {
					argumentsAsStrings = append(
						argumentsAsStrings,
						value_util.CallToStringMethod(runtime_, argument).Content,
					)
				}

				return value_types.StringValue{
					Content: fmt.Sprintf(
						"%s(%s)",
						structName,
						strings.Join(argumentsAsStrings, ", "),
					),
				}
			},

			built_in_declarations.ToStringMethod.Type,
		),

		built_in_declarations.EqualsMethod.Name: function.NewBuiltInFunction(
			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.EqualsMethod.Name,
				reflect.TypeOf(&function.Function{}),
			),

			func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
				return equalsMethodEvaluator(runtime_, arguments...)
			},

			built_in_declarations.EqualsMethod.Type,
		),

		built_in_declarations.NotEqualsMethod.Name: function.NewBuiltInFunction(
			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.EqualsMethod.Name,
				reflect.TypeOf(&function.Function{}),
			),

			func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
				return !equalsMethodEvaluator(runtime_, arguments...)
			},

			built_in_declarations.NotEqualsMethod.Type,
		),

		built_in_declarations.StructConstructorMethod.Name: structConstructor,
	}
}

func ifElse(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
	var branchIndex int

	if arguments[0].(value_types.BooleanValue) {
		branchIndex = 1
	} else {
		branchIndex = 2
	}

	return arguments[branchIndex].(*function.Function).Evaluate(runtime_)
}

func print(runtime_ *runtime.Runtime, suffix string, arguments ...value.Value) value_types.UnitValue {
	serialized := make([]string, 0, len(arguments))

	for _, argument := range arguments {
		serialized = append(serialized, value_util.CallToStringMethod(runtime_, argument).Content)
	}

	fmt.Print(strings.Join(serialized, " ") + suffix)

	return value_types.UnitValue{}
}

func struct_(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
	allFields := map[value_types.StringValue]value.Value{}
	argumentFields := arguments[3].(value_types.TupleValue)

	raiseIncorrectArgumentTypeError := func(i int) {
		errors.RaiseError(runtime_errors.IncorrectBuiltInFunctionArgumentType("__struct__", i))
	}

	populateFields := func(fieldEntries value_types.TupleValue, i int) {
		for _, element := range fieldEntries.Elements {
			entry, ok := element.(value_types.TupleValue)

			if !ok || len(entry.Elements) != 2 {
				raiseIncorrectArgumentTypeError(i)
			}

			name, ok := entry.Elements[0].(value_types.StringValue)

			if !ok {
				raiseIncorrectArgumentTypeError(i)
			}

			allFields[name] = entry.Elements[1]
		}
	}

	result := function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			function.BuiltInFunctionName,
			reflect.TypeOf(value_types.StringValue{}),
		),

		func(_ *runtime.Runtime, resultArguments ...value.Value) value.Value {
			fieldName := resultArguments[0].(value_types.StringValue)
			fieldValue, ok := allFields[fieldName]

			if !ok {
				errors.RaiseError(runtime_errors.UnknownField(fieldName.Content))
			}

			return fieldValue
		},

		&parser_types.FunctionType{
			IsInfix:          false,
			IsPrefix:         false,
			IsStructInstance: true,
		},
	)

	fieldFactory := arguments[2].(*function.Function)
	fieldEntries, ok := fieldFactory.Evaluate(runtime_, result).(value_types.TupleValue)

	if !ok {
		raiseIncorrectArgumentTypeError(2)
	}

	populateFields(fieldEntries, 2)
	populateFields(argumentFields, 3)

	structName := arguments[0].(value_types.StringValue).Content
	structConstructor := arguments[1].(*function.Function)
	argumentFieldNames := make([]string, 0, len(argumentFields.Elements))
	argumentFieldValues := make([]value.Value, 0, len(argumentFields.Elements))

	for _, element := range argumentFields.Elements {
		entry := element.(value_types.TupleValue)

		argumentFieldNames =
			append(argumentFieldNames, entry.Elements[0].(value_types.StringValue).Content)

		argumentFieldValues = append(argumentFieldValues, entry.Elements[1])
	}

	for fieldName, fieldValue := range builtInStructFields(
		structName,
		structConstructor,
		argumentFieldNames,
		argumentFieldValues,
	) {
		key := value_types.StringValue{
			Content: fieldName,
		}

		allFields[key] = fieldValue
	}

	return result
}

func structEquals(
	runtime_ *runtime.Runtime,
	leftConstructor *function.Function,
	leftArgumentNames []string,
	leftArgumentValues []value.Value,
	arguments ...value.Value,
) value_types.BooleanValue {
	rightHandSide := arguments[0].(*function.Function)

	if !rightHandSide.Type_.IsStructInstance {
		return false
	}

	rightConstructor := rightHandSide.Evaluate(runtime_, value_types.StringValue{
		Content: built_in_declarations.StructConstructorMethod.Name,
	})

	if leftConstructor != rightConstructor {
		return false
	}

	for i := 0; i < len(leftArgumentNames); i++ {
		leftArgument := leftArgumentValues[i]
		rightArgument := rightHandSide.Evaluate(runtime_, value_types.StringValue{
			Content: leftArgumentNames[i],
		})

		if !value_util.CallEqualsMethod(runtime_, leftArgument, rightArgument) {
			return false
		}
	}

	return true
}

func tuple(_ *runtime.Runtime, arguments ...value.Value) value.Value {
	return value_types.TupleValue{
		Elements: arguments,
	}
}
