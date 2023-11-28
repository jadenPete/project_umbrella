package entry_errors

import (
	"fmt"

	"project_umbrella/interpreter/errors"
)

var FileNotSpecified = &errors.Error{
	Section: "ENTRY",
	Code:    1,
	Name:    "Please specify a file to run",
}

func FileNotOpened(path string) *errors.Error {
	return &errors.Error{
		Section: "ENTRY",
		Code:    2,
		Name:    fmt.Sprintf("Couldn't open the specified file: %s", path),
	}
}

func StartupFileNotOpened(path string) *errors.Error {
	return &errors.Error{
		Section: "ENTRY",
		Code:    3,
		Name:    fmt.Sprintf("Couldn't open the file in $KRAIT_STARTUP: %s", path),
	}
}
