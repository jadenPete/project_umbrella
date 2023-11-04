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
)

type valueDefinition struct {
	fields map[built_ins.BuiltInFieldID]value
}

func newNumberDefinition[Value integerValue | floatValue](
	value_ Value,
	valueTypeName string,
) *valueDefinition {
	valueType := reflect.TypeOf(value_)

	return &valueDefinition{
		fields: map[built_ins.BuiltInFieldID]value{
			built_ins.ToStringMethodID: newToStringMethod(
				func() string {
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

			built_ins.PlusMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("+", valueType),
				func(_ *runtime, arguments ...value) value {
					return value(value_ + arguments[0].(Value))
				},
			),

			built_ins.MinusMethodID: newMinusMethod(value_),
			built_ins.TimesMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("*", valueType),
				func(_ *runtime, arguments ...value) value {
					return value(value_ * arguments[0].(Value))
				},
			),

			built_ins.OverMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("/", valueType),
				func(_ *runtime, arguments ...value) value {
					rightHandSide := arguments[0].(Value)

					if rightHandSide == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero(valueTypeName, "/"))
					}

					return value(value_ / rightHandSide)
				},
			),

			built_ins.ModuloMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("%", valueType),
				func(_ *runtime, arguments ...value) value {
					modulus := arguments[0].(Value)

					if modulus == 0 {
						errors.RaiseError(runtime_errors.DivisionByZero(valueTypeName, "%"))
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
	),

	built_ins.PrintlnFunctionID: newBuiltInFunction(
		newVariadicFunctionArgumentValidator("println", nil),
		func(runtime_ *runtime, arguments ...value) value {
			return print(runtime_, "\n", arguments...)
		},
	),

	built_ins.UnitValueID:  unitValue{},
	built_ins.FalseValueID: booleanValue(false),
	built_ins.TrueValueID:  booleanValue(true),
	built_ins.IfElseFunctionID: newBuiltInFunction(
		newFixedFunctionArgumentValidator(
			"if_else",
			reflect.TypeOf(*new(booleanValue)),
			reflect.TypeOf(&function{}),
			reflect.TypeOf(&function{}),
		),

		ifElse,
	),
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
		serialized = append(serialized, toString(runtime_, argument))
	}

	fmt.Print(strings.Join(serialized, " ") + suffix)

	return unitValue{}
}

func toString(runtime_ *runtime, value_ value) string {
	resultingValue, ok := value_.
		definition().fields[built_ins.ToStringMethodID].(*function).
		evaluate(runtime_).(stringValue)

	if !ok {
		errors.RaiseError(runtime_errors.ToStringMethodReturnedNonString)
	}

	return resultingValue.content
}

type booleanValue bool

func (value_ booleanValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[built_ins.BuiltInFieldID]value{
			built_ins.ToStringMethodID: newToStringMethod(func() string {
				return fmt.Sprintf("%t", value_)
			}),

			built_ins.NotMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("!"),
				func(runtime_ *runtime, arguments ...value) value {
					return !value_
				},
			),

			built_ins.AndMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("&&", reflect.TypeOf(value_)),
				func(runtime_ *runtime, arguments ...value) value {
					return value_ && arguments[0].(booleanValue)
				},
			),

			built_ins.OrMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("||", reflect.TypeOf(value_)),
				func(runtime_ *runtime, arguments ...value) value {
					return value_ || arguments[0].(booleanValue)
				},
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
					structValue := scope.getValue(element.instruction.Arguments[0])
					fieldID := built_ins.BuiltInFieldID(element.instruction.Arguments[1])

					if field, ok := structValue.definition().fields[fieldID]; ok {
						scope.values[element.instructionValueID] = field
					} else if methodConstructor, ok := universalMethodConstructors[fieldID]; ok {
						scope.values[element.instructionValueID] = methodConstructor(structValue)
					} else {
						errors.RaiseError(
							runtime_errors.UnrecognizedFieldID(
								toString(runtime_, structValue),
								int(fieldID),
							),
						)
					}
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
}

func (function_ *function) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[built_ins.BuiltInFieldID]value{
			built_ins.ToStringMethodID: newToStringMethod(func() string {
				return function_.name
			}),
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

func newBuiltInFunction(
	argumentValidator functionArgumentValidator,
	evaluator func(*runtime, ...value) value,
) *function {
	return &function{
		functionEvaluator: builtInFunctionEvaluator(evaluator),
		argumentValidator: argumentValidator,
		name:              "(built-in function)",
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

		name: name,
	}
}

func newEqualsMethod(value_ value) *function {
	return newBuiltInFunction(
		newFixedFunctionArgumentValidator("==", nil),
		func(runtime_ *runtime, arguments ...value) value {
			return booleanValue(value_ == arguments[0])
		},
	)
}

func newNotEqualsMethod(value_ value) *function {
	return newBuiltInFunction(
		newFixedFunctionArgumentValidator("!=", nil),
		func(runtime_ *runtime, arguments ...value) value {
			return booleanValue(value_ != arguments[0])
		},
	)
}

func newMinusMethod[Value integerValue | floatValue](value_ Value) *function {
	return newBuiltInFunction(
		newIntersectionFunctionArgumentValidator(
			func(argumentTypes []reflect.Type) *errors.Error {
				return runtime_errors.IncorrectCallArgumentCount("0-1", len(argumentTypes))
			},

			newFixedFunctionArgumentValidator("-"),
			newFixedFunctionArgumentValidator("-", reflect.TypeOf(value_)),
		),

		func(_ *runtime, arguments ...value) value {
			if len(arguments) == 0 {
				return value(-value_)
			}

			return value(value_ - arguments[0].(Value))
		},
	)
}

func newToStringMethod(result func() string) *function {
	return newBuiltInFunction(
		newFixedFunctionArgumentValidator("__to_str__"),
		func(runtime_ *runtime, arguments ...value) value {
			return stringValue{result()}
		},
	)
}

var universalMethodConstructors = map[built_ins.BuiltInFieldID]func(value) *function{
	built_ins.EqualsMethodID:    newEqualsMethod,
	built_ins.NotEqualsMethodID: newNotEqualsMethod,
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
		fields: map[built_ins.BuiltInFieldID]value{
			built_ins.ToStringMethodID: newToStringMethod(
				func() string {
					return value_.content
				},
			),

			built_ins.PlusMethodID: newBuiltInFunction(
				newFixedFunctionArgumentValidator("+", reflect.TypeOf(stringValue{})),
				func(_ *runtime, arguments ...value) value {
					return stringValue{value_.content + arguments[0].(stringValue).content}
				},
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

type unitValue struct{}

func (value_ unitValue) definition() *valueDefinition {
	return &valueDefinition{
		fields: map[built_ins.BuiltInFieldID]value{
			built_ins.ToStringMethodID: newToStringMethod(func() string {
				return "(unit)"
			}),
		},
	}
}
