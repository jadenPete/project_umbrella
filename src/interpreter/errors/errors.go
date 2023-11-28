package errors

import (
	"fmt"
	"os"
)

type Error struct {
	Section     string
	Code        int
	Name        string
	Description string
}

/*
 * `Position` represents the zero-indexed, inclusive-exclusive range within the source file
 * within which the error occurred.
 */
type Position struct {
	Filename string
	Start    int
	End      int
}

type PositionalError struct {
	Error    *Error
	Position *Position
}

func RaiseError(error_ *Error) {
	description := ""

	if error_.Description != "" {
		description = fmt.Sprintf("\n%s\n", error_.Description)
	}

	fmt.Fprintf(
		os.Stderr,
		"Error (%s-%d): %s\n%s",
		error_.Section,
		error_.Code,
		error_.Name,
		description,
	)

	os.Exit(1)
}

func RaisePositionalError(error_ *PositionalError) {
	description := ""

	if error_.Error.Description != "" {
		description = fmt.Sprintf("\n%s", error_.Error.Description)
	}

	RaiseError(
		&Error{
			Section: error_.Error.Section,
			Code:    error_.Error.Code,
			Name:    error_.Error.Name,
			Description: fmt.Sprintf(
				"%s:\n%s%s",
				error_.Position.Filename,
				highlightedSource(error_.Position),
				description,
			),
		},
	)
}
