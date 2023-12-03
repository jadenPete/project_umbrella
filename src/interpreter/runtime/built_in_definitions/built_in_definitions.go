package built_in_definitions

import (
	"fmt"
	"reflect"
	"strings"

	"project_umbrella/interpreter/bytecode_generator/built_in_declarations"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/loader"
	"project_umbrella/interpreter/parser/parser_types"
	"project_umbrella/interpreter/runtime"
	"project_umbrella/interpreter/runtime/value"
	"project_umbrella/interpreter/runtime/value_types"
	"project_umbrella/interpreter/runtime/value_types/function"
	"project_umbrella/interpreter/runtime/value_util"
)

var BuiltInValues = map[built_in_declarations.BuiltInValueID]value.Value{
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

	built_in_declarations.ImportFunctionID: function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			"import",
			reflect.TypeOf(*new(value_types.StringValue)),
		),

		func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
			return import_(runtime_, loader.ModuleRequest, arguments...)
		},

		parser_types.NormalFunction,
	),

	built_in_declarations.ImportLibraryFunctionID: function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			"import_library",
			reflect.TypeOf(*new(value_types.StringValue)),
		),

		func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
			return import_(runtime_, loader.LibraryRequest, arguments...)
		},

		parser_types.NormalFunction,
	),

	built_in_declarations.ModuleFunctionID: function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			"__module__",
			reflect.TypeOf(&value_types.TupleValue{}),
		),

		module,
		parser_types.NormalFunction,
	),

	built_in_declarations.StructFunctionID: function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			"__struct__",
			reflect.TypeOf(*new(value_types.StringValue)),
			reflect.TypeOf(&function.Function{}),
			reflect.TypeOf(&function.Function{}),
			reflect.TypeOf(&value_types.TupleValue{}),
		),

		struct_,
		parser_types.NormalFunction,
	),

	built_in_declarations.TupleFunctionID: function.NewBuiltInFunction(
		function.NewVariadicFunctionArgumentValidator("__tuple__", nil),
		tuple,
		parser_types.NormalFunction,
	),

	built_in_declarations.UnitValueID: value_types.UnitValue{},
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
		built_in_declarations.UniversalEqualsMethod.Name: function.NewBuiltInFunction(
			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.UniversalEqualsMethod.Name,
				reflect.TypeOf(&function.Function{}),
			),

			func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
				return equalsMethodEvaluator(runtime_, arguments...)
			},

			built_in_declarations.UniversalEqualsMethod.Type,
		),

		built_in_declarations.UniversalNotEqualsMethod.Name: function.NewBuiltInFunction(
			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.UniversalEqualsMethod.Name,
				reflect.TypeOf(&function.Function{}),
			),

			func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
				return !equalsMethodEvaluator(runtime_, arguments...)
			},

			built_in_declarations.UniversalNotEqualsMethod.Type,
		),

		built_in_declarations.StructConstructorMethod.Name: structConstructor,
		built_in_declarations.UniversalToStringMethod.Name: function.NewBuiltInFunction(
			function.NewFixedFunctionArgumentValidator(
				built_in_declarations.UniversalToStringMethod.Name,
			),

			func(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
				argumentsAsStrings := make([]string, 0, len(structArgumentValues))

				for _, argument := range structArgumentValues {
					argumentsAsStrings = append(
						argumentsAsStrings,
						string(value_util.CallToStringMethod(runtime_, argument)),
					)
				}

				return value_types.StringValue(
					fmt.Sprintf(
						"%s(%s)",
						structName,
						strings.Join(argumentsAsStrings, ", "),
					),
				)
			},

			built_in_declarations.UniversalToStringMethod.Type,
		),
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

func import_(runtime_ *runtime.Runtime, type_ loader.LoaderRequestType, arguments ...value.Value) value.Value {
	runtime_.LoaderChannel.LoadRequest <- &loader.LoaderRequest{
		Type: type_,
		Name: string(arguments[0].(value_types.StringValue)),
	}

	return <-runtime_.LoaderChannel.LoadResponse
}

func module(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
	fields, ok := moduleOrStructFieldsToMap(arguments[0].(*value_types.TupleValue))

	if !ok {
		errors.RaiseError(runtime_errors.IncorrectBuiltInFunctionArgumentType("__module__", 0))
	}

	return newLookupFunction(fields)
}

func moduleOrStructFieldsToMap(
	fields *value_types.TupleValue,
) (map[value_types.StringValue]value.Value, bool) {
	result := make(map[value_types.StringValue]value.Value, len(fields.Elements))

	for _, element := range fields.Elements {
		field, ok := element.(*value_types.TupleValue)

		if !ok || len(field.Elements) != 2 {
			return nil, false
		}

		name, ok := field.Elements[0].(value_types.StringValue)

		if !ok {
			return nil, false
		}

		result[name] = field.Elements[1]
	}

	return result, true
}

func newLookupFunction(fields map[value_types.StringValue]value.Value) *function.Function {
	return function.NewBuiltInFunction(
		function.NewFixedFunctionArgumentValidator(
			function.BuiltInFunctionName,
			reflect.TypeOf(*new(value_types.StringValue)),
		),

		func(_ *runtime.Runtime, resultArguments ...value.Value) value.Value {
			fieldName := resultArguments[0].(value_types.StringValue)
			fieldValue, ok := fields[fieldName]

			if !ok {
				errors.RaiseError(runtime_errors.UnknownField(string(fieldName)))
			}

			return fieldValue
		},

		&parser_types.FunctionType{
			IsInfix:  false,
			IsPrefix: false,
			IsLookup: true,
		},
	)
}

func struct_(runtime_ *runtime.Runtime, arguments ...value.Value) value.Value {
	allFields := map[value_types.StringValue]value.Value{}

	raiseIncorrectArgumentTypeError := func(i int) {
		errors.RaiseError(runtime_errors.IncorrectBuiltInFunctionArgumentType("__struct__", i))
	}

	populateFields := func(fieldEntries *value_types.TupleValue, i int) {
		newFields, ok := moduleOrStructFieldsToMap(fieldEntries)

		if ok {
			for name, value := range newFields {
				allFields[name] = value
			}
		} else {
			raiseIncorrectArgumentTypeError(i)
		}
	}

	result := newLookupFunction(allFields)
	fieldFactory := arguments[2].(*function.Function)
	fieldEntries, ok := fieldFactory.Evaluate(runtime_, result).(*value_types.TupleValue)

	if !ok {
		raiseIncorrectArgumentTypeError(2)
	}

	argumentFieldEntries := arguments[3].(*value_types.TupleValue)

	populateFields(fieldEntries, 2)
	populateFields(argumentFieldEntries, 3)

	structName := string(arguments[0].(value_types.StringValue))
	structConstructor := arguments[1].(*function.Function)
	argumentFieldNames := make([]string, 0, len(argumentFieldEntries.Elements))
	argumentFieldValues := make([]value.Value, 0, len(argumentFieldEntries.Elements))

	for _, element := range argumentFieldEntries.Elements {
		entry := element.(*value_types.TupleValue)

		argumentFieldNames =
			append(argumentFieldNames, string(entry.Elements[0].(value_types.StringValue)))

		argumentFieldValues = append(argumentFieldValues, entry.Elements[1])
	}

	for fieldName, fieldValue := range builtInStructFields(
		structName,
		structConstructor,
		argumentFieldNames,
		argumentFieldValues,
	) {
		allFields[value_types.StringValue(fieldName)] = fieldValue
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

	if !rightHandSide.Type_.IsLookup {
		return false
	}

	rightConstructor := rightHandSide.Evaluate(
		runtime_,
		value_types.StringValue(built_in_declarations.StructConstructorMethod.Name),
	)

	if leftConstructor != rightConstructor {
		return false
	}

	for i := 0; i < len(leftArgumentNames); i++ {
		leftArgument := leftArgumentValues[i]
		rightArgument :=
			rightHandSide.Evaluate(runtime_, value_types.StringValue(leftArgumentNames[i]))

		if !value_util.CallEqualsMethod(runtime_, leftArgument, rightArgument) {
			return false
		}
	}

	return true
}

func tuple(_ *runtime.Runtime, arguments ...value.Value) value.Value {
	return &value_types.TupleValue{
		Elements: arguments,
	}
}
