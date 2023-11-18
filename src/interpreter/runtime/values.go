package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/bytecode_generator/built_ins"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/runtime_errors"
	"project_umbrella/interpreter/parser/parser_types"
)

type valueDefinition struct {
	fields map[string]value
}

func newNumberDefinition[Value integerValue | floatValue](
	value_ Value,
	valueTypeName string,
) *valueDefinition {
	valueType := reflect.TypeOf(value_)

	return &valueDefinition{
		fields: map[string]value{
			built_ins.ToStringMethod.Name: newToStringMethod(
				func(*runtime) string {
					switch value_ := any(value_).(type) {
					case integerValue:
						return fmt.Sprintf("%d", value_)

					case floatValue:
						return fmt.Sprintf("%g", value_)

					default:
						return ""
					}
				},
			),

			built_ins.PlusMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.PlusMethod.Name, valueType),
				func(_ *runtime, arguments ...value) value {
					return value(value_ + arguments[0].(Value))
				},

				built_ins.PlusMethod.Type,
			),

			built_ins.MinusMethod.Name: newMinusMethod(value_),
			built_ins.TimesMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.TimesMethod.Name, valueType),
				func(_ *runtime, arguments ...value) value {
					return value(value_ * arguments[0].(Value))
				},

				built_ins.TimesMethod.Type,
			),

			built_ins.OverMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.OverMethod.Name, valueType),
				func(_ *runtime, arguments ...value) value {
					rightHandSide := arguments[0].(Value)

					if rightHandSide == 0 {
						errors.RaiseError(
							runtime_errors.DivisionByZero(valueTypeName, built_ins.OverMethod.Name),
						)
					}

					return value(value_ / rightHandSide)
				},

				built_ins.OverMethod.Type,
			),

			built_ins.ModuloMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.ModuloMethod.Name, valueType),
				func(_ *runtime, arguments ...value) value {
					modulus := arguments[0].(Value)

					if modulus == 0 {
						errors.RaiseError(
							runtime_errors.DivisionByZero(valueTypeName, built_ins.ModuloMethod.Name),
						)
					}

					switch value_ := any(value_).(type) {
					case integerValue:
						return value_ % integerValue(modulus)

					case floatValue:
						return floatValue(math.Mod(float64(value_), float64(modulus)))

					default:
						return nil
					}
				},

				built_ins.ModuloMethod.Type,
			),

			built_ins.LessThanMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.LessThanMethod.Name, valueType),
				func(_ *runtime, arguments ...value) value {
					return booleanValue(value_ < arguments[0].(Value))
				},

				built_ins.LessThanMethod.Type,
			),

			built_ins.LessThanOrEqualToMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(
					built_ins.LessThanOrEqualToMethod.Name,
					valueType,
				),

				func(_ *runtime, arguments ...value) value {
					return booleanValue(value_ <= arguments[0].(Value))
				},

				built_ins.LessThanOrEqualToMethod.Type,
			),

			built_ins.GreaterThanMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.GreaterThanMethod.Name, valueType),
				func(_ *runtime, arguments ...value) value {
					return booleanValue(value_ > arguments[0].(Value))
				},

				built_ins.GreaterThanMethod.Type,
			),

			built_ins.GreaterThanOrEqualToMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(
					built_ins.GreaterThanOrEqualToMethod.Name,
					valueType,
				),

				func(_ *runtime, arguments ...value) value {
					return booleanValue(value_ >= arguments[0].(Value))
				},

				built_ins.GreaterThanOrEqualToMethod.Type,
			),
		},
	}
}

type value interface {
	definition() *valueDefinition
}

func newValueFromConstant(constant bytecode_generator.Constant) value {
	switch constant.Type {
	case bytecode_generator.FloatConstant:
		var value float64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return floatValue(value)

	case bytecode_generator.IntegerConstant:
		var value int64

		buffer := bytes.NewBufferString(constant.Encoded)

		if err := binary.Read(buffer, binary.LittleEndian, &value); err != nil {
			panic(err)
		}

		return integerValue(value)

	case bytecode_generator.StringConstant:
		return stringValue{constant.Encoded}
	}

	return nil
}

var builtInValues = map[built_ins.BuiltInValueID]value{
	built_ins.PrintFunctionID: newBuiltInFunction(
		newVariadicFunctionArgumentValidator("print", nil),
		func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "", arguments...)
		},

		parser_types.NormalFunction,
	),

	built_ins.PrintlnFunctionID: newBuiltInFunction(
		newVariadicFunctionArgumentValidator("println", nil),
		func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "\n", arguments...)
		},

		parser_types.NormalFunction,
	),

	built_ins.UnitValueID:  unitValue{},
	built_ins.FalseValueID: booleanValue(false),
	built_ins.TrueValueID:  booleanValue(true),
	built_ins.IfElseFunctionID: newBuiltInFunction(
		newFixedFunctionArgumentValidator(
			"__if_else__",
			reflect.TypeOf(*new(booleanValue)),
			reflect.TypeOf(&function{}),
			reflect.TypeOf(&function{}),
		),

		ifElse,
		parser_types.NormalFunction,
	),

	built_ins.TupleFunctionID: newBuiltInFunction(
		newVariadicFunctionArgumentValidator("__tuple__", nil),
		tuple,
		parser_types.NormalFunction,
	),

	built_ins.StructFunctionID: newBuiltInFunction(
		newFixedFunctionArgumentValidator(
			"__struct__",
			reflect.TypeOf(stringValue{}),
			reflect.TypeOf(&function{}),
			reflect.TypeOf(&function{}),
			reflect.TypeOf(tupleValue{}),
		),

		struct_,
		parser_types.NormalFunction,
	),
}

func builtInEquals(runtime_ *runtime, value1 value, value2 value) booleanValue {
	tuple1, ok1 := value1.(tupleValue)
	tuple2, ok2 := value2.(tupleValue)

	if ok1 && ok2 {
		if len(tuple1.elements) != len(tuple2.elements) {
			return false
		}

		for i, element := range tuple1.elements {
			if !callEqualsMethod(runtime_, element, tuple2.elements[i]) {
				return false
			}
		}

		return true
	}

	return value1 == value2
}

func builtInStructFields(
	structName string,
	structConstructor *function,
	structArgumentNames []string,
	structArgumentValues []value,
) map[string]value {
	equalsMethodEvaluator := func(runtime_ *runtime, arguments ...value) booleanValue {
		return structEquals(
			runtime_,
			structConstructor,
			structArgumentNames,
			structArgumentValues,
			arguments...,
		)
	}

	return map[string]value{
		built_ins.ToStringMethod.Name: newBuiltInFunction(
			newFixedFunctionArgumentValidator(built_ins.ToStringMethod.Name),
			func(runtime_ *runtime, arguments ...value) value {
				argumentsAsStrings := make([]string, 0, len(structArgumentValues))

				for _, argument := range structArgumentValues {
					argumentsAsStrings =
						append(argumentsAsStrings, callToStringMethod(runtime_, argument).content)
				}

				return stringValue{
					content: fmt.Sprintf(
						"%s(%s)",
						structName,
						strings.Join(argumentsAsStrings, ", "),
					),
				}
			},

			built_ins.ToStringMethod.Type,
		),

		built_ins.EqualsMethod.Name: newBuiltInFunction(
			newFixedFunctionArgumentValidator(
				built_ins.EqualsMethod.Name,
				reflect.TypeOf(&function{}),
			),

			func(runtime_ *runtime, arguments ...value) value {
				return equalsMethodEvaluator(runtime_, arguments...)
			},

			built_ins.EqualsMethod.Type,
		),

		built_ins.NotEqualsMethod.Name: newBuiltInFunction(
			newFixedFunctionArgumentValidator(
				built_ins.EqualsMethod.Name,
				reflect.TypeOf(&function{}),
			),

			func(runtime_ *runtime, arguments ...value) value {
				return !equalsMethodEvaluator(runtime_, arguments...)
			},

			built_ins.NotEqualsMethod.Type,
		),

		built_ins.StructConstructorMethod.Name: structConstructor,
	}
}

func callBuiltInMethod[ReturnValue value](
	runtime_ *runtime,
	value_ value,
	methodName string,
	returnValueTypeName string,
	arguments ...value,
) ReturnValue {
	method, ok := lookupField(runtime_, value_, methodName, parser_types.NormalSelect).(*function)

	if !ok {
		errors.RaiseError(runtime_errors.NonFunctionCalled)
	}

	result, ok := method.evaluate(runtime_, arguments...).(ReturnValue)

	if !ok {
		errors.RaiseError(
			runtime_errors.UniversalMethodReturnedIncorrectValue(methodName, returnValueTypeName),
		)
	}

	return result
}

func callEqualsMethod(runtime_ *runtime, value1 value, value2 value) booleanValue {
	return callBuiltInMethod[booleanValue](
		runtime_,
		value1,
		built_ins.EqualsMethod.Name,
		"boolean",
		value2,
	)
}

func callToStringMethod(runtime_ *runtime, value_ value) stringValue {
	return callBuiltInMethod[stringValue](runtime_, value_, built_ins.ToStringMethod.Name, "str")
}

func ifElse(runtime_ *runtime, arguments ...value) value {
	var branchIndex int

	if arguments[0].(booleanValue) {
		branchIndex = 1
	} else {
		branchIndex = 2
	}

	return arguments[branchIndex].(*function).evaluate(runtime_)
}

func print(runtime_ *runtime, suffix string, arguments ...value) unitValue {
	serialized := make([]string, 0, len(arguments))

	for _, argument := range arguments {
		serialized = append(serialized, callToStringMethod(runtime_, argument).content)
	}

	fmt.Print(strings.Join(serialized, " ") + suffix)

	return unitValue{}
}

func struct_(runtime_ *runtime, arguments ...value) value {
	allFields := map[stringValue]value{}
	argumentFields := arguments[3].(tupleValue)

	raiseIncorrectArgumentTypeError := func(i int) {
		errors.RaiseError(runtime_errors.IncorrectBuiltInFunctionArgumentType("__struct__", i))
	}

	populateFields := func(fieldEntries tupleValue, i int) {
		for _, element := range fieldEntries.elements {
			entry, ok := element.(tupleValue)

			if !ok || len(entry.elements) != 2 {
				raiseIncorrectArgumentTypeError(i)
			}

			name, ok := entry.elements[0].(stringValue)

			if !ok {
				raiseIncorrectArgumentTypeError(i)
			}

			allFields[name] = entry.elements[1]
		}
	}

	result := newBuiltInFunction(
		newFixedFunctionArgumentValidator(builtInFunctionName, reflect.TypeOf(stringValue{})),
		func(_ *runtime, resultArguments ...value) value {
			fieldName := resultArguments[0].(stringValue)
			fieldValue, ok := allFields[fieldName]

			if !ok {
				errors.RaiseError(runtime_errors.UnknownField(fieldName.content))
			}

			return fieldValue
		},

		&parser_types.FunctionType{
			IsInfix:          false,
			IsPrefix:         false,
			IsStructInstance: true,
		},
	)

	fieldFactory := arguments[2].(*function)
	fieldEntries, ok := fieldFactory.evaluate(runtime_, result).(tupleValue)

	if !ok {
		raiseIncorrectArgumentTypeError(2)
	}

	populateFields(fieldEntries, 2)
	populateFields(argumentFields, 3)

	structName := arguments[0].(stringValue).content
	structConstructor := arguments[1].(*function)
	argumentFieldNames := make([]string, 0, len(argumentFields.elements))
	argumentFieldValues := make([]value, 0, len(argumentFields.elements))

	for _, element := range argumentFields.elements {
		entry := element.(tupleValue)

		argumentFieldNames = append(argumentFieldNames, entry.elements[0].(stringValue).content)
		argumentFieldValues = append(argumentFieldValues, entry.elements[1])
	}

	for fieldName, fieldValue := range builtInStructFields(
		structName,
		structConstructor,
		argumentFieldNames,
		argumentFieldValues,
	) {
		key := stringValue{
			content: fieldName,
		}

		allFields[key] = fieldValue
	}

	return result
}

func structEquals(
	runtime_ *runtime,
	leftConstructor *function,
	leftArgumentNames []string,
	leftArgumentValues []value,
	arguments ...value,
) booleanValue {
	rightHandSide := arguments[0].(*function)

	if !rightHandSide.type_.IsStructInstance {
		return false
	}

	rightConstructor := rightHandSide.evaluate(runtime_, stringValue{
		content: built_ins.StructConstructorMethod.Name,
	})

	if leftConstructor != rightConstructor {
		return false
	}

	for i := 0; i < len(leftArgumentNames); i++ {
		leftArgument := leftArgumentValues[i]
		rightArgument := rightHandSide.evaluate(runtime_, stringValue{
			leftArgumentNames[i],
		})

		if !callEqualsMethod(runtime_, leftArgument, rightArgument) {
			return false
		}
	}

	return true
}

func tuple(runtime_ *runtime, arguments ...value) value {
	return tupleValue{
		elements: arguments,
	}
}

type booleanValue bool

func (value_ booleanValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[string]value{
			built_ins.ToStringMethod.Name: newToStringMethod(
				func(*runtime) string {
					return fmt.Sprintf("%t", value_)
				},
			),

			built_ins.NotMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.NotMethod.Name),
				func(runtime_ *runtime, arguments ...value) value {
					return !value_
				},

				built_ins.NotMethod.Type,
			),

			built_ins.AndMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.AndMethod.Name, reflect.TypeOf(value_)),
				func(runtime_ *runtime, arguments ...value) value {
					return value_ && arguments[0].(booleanValue)
				},

				built_ins.AndMethod.Type,
			),

			built_ins.OrMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator(built_ins.OrMethod.Name, reflect.TypeOf(value_)),
				func(runtime_ *runtime, arguments ...value) value {
					return value_ || arguments[0].(booleanValue)
				},

				built_ins.OrMethod.Type,
			),
		},
	}
}

type functionArgumentValidator func(argumentTypes []reflect.Type) *errors.Error

func newFixedFunctionArgumentValidator(
	name string,
	parameterTypes ...reflect.Type,
) functionArgumentValidator {
	return functionArgumentValidator(
		func(argumentTypes []reflect.Type) *errors.Error {
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
		},
	)
}

func newIntersectionFunctionArgumentValidator(
	err func([]reflect.Type) *errors.Error,
	validators ...functionArgumentValidator,
) functionArgumentValidator {
	return functionArgumentValidator(
		func(argumentTypes []reflect.Type) *errors.Error {
			for _, validator := range validators {
				if err := validator(argumentTypes); err == nil {
					return nil
				}
			}

			return err(argumentTypes)
		},
	)
}

func newVariadicFunctionArgumentValidator(
	name string,
	parameterType reflect.Type,
) functionArgumentValidator {
	return functionArgumentValidator(
		func(argumentTypes []reflect.Type) *errors.Error {
			if parameterType == nil {
				return nil
			}

			for i, argumentType := range argumentTypes {
				if !argumentType.AssignableTo(parameterType) {
					return runtime_errors.IncorrectBuiltInFunctionArgumentType(name, i)
				}
			}

			return nil
		},
	)
}

type functionEvaluator interface {
	evaluator(*runtime, ...value) value
}

type builtInFunctionEvaluator func(*runtime, ...value) value

func (evaluator builtInFunctionEvaluator) evaluator(runtime_ *runtime, arguments ...value) value {
	return evaluator(runtime_, arguments...)
}

type bytecodeFunctionEvaluator struct {
	containingScope *scope
	blockGraph      *bytecodeFunctionBlockGraph
}

func lookupField(
	runtime_ *runtime,
	value_ value,
	fieldName string,
	selectType parser_types.SelectType,
) value {
	var universalMethodConstructors = map[string]func(value) *function{
		built_ins.EqualsMethod.Name:    newEqualsMethod,
		built_ins.NotEqualsMethod.Name: newNotEqualsMethod,
	}

	var result value

	if function_, ok := value_.(*function); ok && function_.type_.IsStructInstance {
		result = function_.evaluate(runtime_, stringValue{
			content: fieldName,
		})
	} else if field, ok := value_.definition().fields[fieldName]; ok {
		result = field
	} else if methodConstructor, ok := universalMethodConstructors[fieldName]; ok {
		result = methodConstructor(value_)
	} else {
		errors.RaiseError(runtime_errors.UnknownField(fieldName))
	}

	if function_, ok := result.(*function); ok {
		if !function_.type_.CanSelectBy(selectType) {
			errors.RaiseError(
				runtime_errors.MethodCalledImproperly(
					callToStringMethod(runtime_, value_).content,
					fieldName,
					function_.type_,
					selectType,
				),
			)
		}
	}

	return result
}

// TODO: Make this concurrent
func (evaluator *bytecodeFunctionEvaluator) evaluator(runtime_ *runtime, arguments ...value) value {
	firstValueID := 0

	if evaluator.containingScope != nil {
		firstValueID = evaluator.blockGraph.firstValueID
	}

	scope := &scope{
		parent:       evaluator.containingScope,
		firstValueID: firstValueID,
		values:       map[int]value{},
	}

	for i, argument := range arguments {
		scope.values[scope.firstValueID+i] = argument
	}

	isAcyclic := evaluator.blockGraph.Evaluate(func(i int) {
		callArguments := []value{}

		switch node := evaluator.blockGraph.Nodes[i].(type) {
		case *bytecodeFunctionBlockGraph:
			scope.values[node.valueID] = newBytecodeFunction(
				node.parameterCount,
				&bytecodeFunctionEvaluator{
					containingScope: scope,
					blockGraph:      node,
				},
			)

		case instructionList:
			for _, element := range node {
				switch element.instruction.Type {
				case bytecode_generator.PushArgumentInstruction:
					callArguments =
						append(callArguments, scope.getValue(element.instruction.Arguments[0]))

				case bytecode_generator.ValueCopyInstruction:
					scope.values[element.instructionValueID] =
						scope.getValue(element.instruction.Arguments[0])

				case bytecode_generator.ValueFromCallInstruction:
					function_, ok := scope.getValue(element.instruction.Arguments[0]).(*function)

					if !ok {
						errors.RaiseError(runtime_errors.NonFunctionCalled)
					}

					scope.values[element.instructionValueID] =
						function_.evaluate(runtime_, callArguments...)

				case bytecode_generator.ValueFromConstantInstruction:
					scope.values[element.instructionValueID] =
						runtime_.constants[element.instruction.Arguments[0]]

				case bytecode_generator.ValueFromStructValueInstruction:
					value_ := scope.getValue(element.instruction.Arguments[0])
					fieldNameConstant := runtime_.constants[element.instruction.Arguments[1]]
					fieldNameValue, ok := fieldNameConstant.(stringValue)

					if !ok {
						errors.RaiseError(runtime_errors.NonStringFieldName)
					}

					fieldName := fieldNameValue.content
					selectType := parser_types.SelectType(element.instruction.Arguments[2])

					scope.values[element.instructionValueID] =
						lookupField(runtime_, value_, fieldName, selectType)
				}
			}
		}
	})

	if !isAcyclic {
		errors.RaiseError(runtime_errors.ValueCycle)
	}

	if len(scope.values) == 0 {
		errors.RaiseError(runtime_errors.EmptyFunctionBlockGraph)
	}

	lastValueID := 0

	for valueID := range scope.values {
		if valueID > lastValueID {
			lastValueID = valueID
		}
	}

	return scope.values[lastValueID]
}

type function struct {
	functionEvaluator

	argumentValidator functionArgumentValidator
	name              string
	type_             *parser_types.FunctionType
}

func (function_ *function) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[string]value{
			built_ins.ToStringMethod.Name: newToStringMethod(
				func(*runtime) string {
					return function_.name
				},
			),
		},
	}
}

func (function_ *function) evaluate(runtime_ *runtime, arguments ...value) value {
	argumentTypes := make([]reflect.Type, 0, len(arguments))

	for _, argument := range arguments {
		argumentTypes = append(argumentTypes, reflect.TypeOf(argument))
	}

	if err := function_.argumentValidator(argumentTypes); err != nil {
		errors.RaiseError(err)
	}

	return function_.evaluator(runtime_, arguments...)
}

const builtInFunctionName = "(built-in function)"

func newBuiltInFunction(
	argumentValidator functionArgumentValidator,
	evaluator func(*runtime, ...value) value,
	type_ *parser_types.FunctionType,
) *function {
	return &function{
		functionEvaluator: builtInFunctionEvaluator(evaluator),
		argumentValidator: argumentValidator,
		name:              builtInFunctionName,
		type_:             type_,
	}
}

func newBytecodeFunction(
	parameterCount int,
	evaluator *bytecodeFunctionEvaluator,
) *function {
	name := "(function)"

	return &function{
		functionEvaluator: evaluator,
		argumentValidator: newFixedFunctionArgumentValidator(
			name,
			common.Repeat[reflect.Type](nil, parameterCount)...,
		),

		name:  name,
		type_: parser_types.NormalFunction,
	}
}

func newEqualsMethod(value_ value) *function {
	return newBuiltInFunction(
		newFixedFunctionArgumentValidator(built_ins.EqualsMethod.Name, nil),
		func(runtime_ *runtime, arguments ...value) value {
			return builtInEquals(runtime_, value_, arguments[0])
		},

		built_ins.EqualsMethod.Type,
	)
}

func newNotEqualsMethod(value_ value) *function {
	return newBuiltInFunction(
		newFixedFunctionArgumentValidator(built_ins.NotEqualsMethod.Name, nil),
		func(runtime_ *runtime, arguments ...value) value {
			return !builtInEquals(runtime_, value_, arguments[0])
		},

		built_ins.NotEqualsMethod.Type,
	)
}

func newMinusMethod[Value integerValue | floatValue](value_ Value) *function {
	return newBuiltInFunction(
		newIntersectionFunctionArgumentValidator(
			func(argumentTypes []reflect.Type) *errors.Error {
				if len(argumentTypes) == 1 {
					return runtime_errors.IncorrectBuiltInFunctionArgumentType(
						built_ins.MinusMethod.Name,
						0,
					)
				}

				return runtime_errors.IncorrectCallArgumentCount(
					"0-1",
					true,
					len(argumentTypes),
				)
			},

			newFixedFunctionArgumentValidator(built_ins.MinusMethod.Name),
			newFixedFunctionArgumentValidator(built_ins.MinusMethod.Name, reflect.TypeOf(value_)),
		),

		func(_ *runtime, arguments ...value) value {
			if len(arguments) == 0 {
				return value(-value_)
			}

			return value(value_ - arguments[0].(Value))
		},

		built_ins.MinusMethod.Type,
	)
}

func newToStringMethod(result func(*runtime) string) *function {
	return newBuiltInFunction(
		newFixedFunctionArgumentValidator(built_ins.ToStringMethod.Name),
		func(runtime_ *runtime, arguments ...value) value {
			return stringValue{result(runtime_)}
		},

		built_ins.ToStringMethod.Type,
	)
}

type floatValue float64

func (value_ floatValue) definition() *valueDefinition {
	return newNumberDefinition(value_, "float")
}

type integerValue int64

func (value_ integerValue) definition() *valueDefinition {
	return newNumberDefinition(value_, "int")
}

type stringValue struct {
	content string
}

func (value_ stringValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[string]value{
			built_ins.ToStringMethod.Name: newToStringMethod(
				func(*runtime) string {
					return value_.content
				},
			),

			built_ins.PlusMethod.Name: newBuiltInFunction(
				newFixedFunctionArgumentValidator("+", reflect.TypeOf(stringValue{})),
				func(_ *runtime, arguments ...value) value {
					return stringValue{value_.content + arguments[0].(stringValue).content}
				},

				built_ins.PlusMethod.Type,
			),
		},
	}
}

type scope struct {
	parent       *scope
	firstValueID int
	values       map[int]value
}

func (scope_ *scope) getValue(valueID int) value {
	if builtInValue, ok := builtInValues[built_ins.BuiltInValueID(valueID)]; ok {
		return builtInValue
	}

	currentScope := scope_

	for currentScope.firstValueID > valueID {
		currentScope = currentScope.parent
	}

	return currentScope.values[valueID]
}

type tupleValue struct {
	elements []value
}

func (value_ tupleValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[string]value{
			built_ins.ToStringMethod.Name: newToStringMethod(
				func(runtime_ *runtime) string {
					var insideParentheses string

					switch len(value_.elements) {
					case 0:
						insideParentheses = ","

					case 1:
						insideParentheses = fmt.Sprintf(
							"%s,",
							callToStringMethod(runtime_, value_.elements[0]).content,
						)

					default:
						elementsAsStrings := make([]string, 0, len(value_.elements))

						for _, element := range value_.elements {
							elementsAsStrings = append(
								elementsAsStrings,
								callToStringMethod(runtime_, element).content,
							)
						}

						insideParentheses = strings.Join(elementsAsStrings, ", ")
					}

					return fmt.Sprintf("(%s)", insideParentheses)
				},
			),
		},
	}
}

type unitValue struct{}

func (value_ unitValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[string]value{
			built_ins.ToStringMethod.Name: newToStringMethod(
				func(*runtime) string {
					return "(unit)"
				},
			),
		},
	}
}
