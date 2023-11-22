package main

import (
	"os"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/entry_errors"
	"project_umbrella/interpreter/loader/module_loader"
)

func main() {
	if len(os.Args) < 2 {
		errors.RaiseError(entry_errors.FileNotSpecified)
	}

	module_loader.NewModuleLoader().LoadFile(os.Args[1])
}
