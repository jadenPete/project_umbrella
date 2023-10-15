package main

import (
	"os"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/entry_errors"
	"project_umbrella/interpreter/parser"
	"project_umbrella/interpreter/runtime"
)

func executeSource(source string) {
	concreteTree, err := parser.ParseString(source)

	if err != nil {
		// TODO: Make error reporting better
		println(err.Error())

		os.Exit(1)
	}

	abstractTree := concreteTree.Abstract()

	runtime.ExecuteBytecode(bytecode_generator.ExpressionToBytecodeFromCache(abstractTree, source))
}

func main() {
	if len(os.Args) < 2 {
		errors.RaiseError(entry_errors.FileNotSpecified)
	}

	fileContent, err := os.ReadFile(os.Args[1])

	if err != nil {
		errors.RaiseError(entry_errors.FileNotOpened)
	}

	executeSource(string(fileContent))
}
