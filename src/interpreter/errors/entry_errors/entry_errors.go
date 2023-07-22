package entry_errors

import "project_umbrella/interpreter/errors"

var FileNotSpecified = &errors.Error{
	Section: "ENTRY",
	Code:    1,
	Name:    "Please specify a file to run",
}

var FileNotOpened = &errors.Error{
	Section: "ENTRY",
	Code:    2,
	Name:    "Couldn't open the specified file",
}
