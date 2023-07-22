package lexer_errors

import (
	"fmt"
	"project_umbrella/interpreter/errors"
)

func indentCharacterName(indentCharacter rune, indentLength int) string {
	var indentCharacterName string

	if indentCharacter == '\t' {
		indentCharacterName = "tab"
	} else {
		indentCharacterName = "space"
	}

	if indentLength == 1 {
		return indentCharacterName
	}

	return fmt.Sprintf("%ss", indentCharacterName)
}

func InconsistentIndentation(
	indentCharacter rune,
	indentLength int,
	incorrectIndentCharacter rune,
	incorrectIndentLength int,
) *errors.Error {
	return &errors.Error{
		Section: "LEXER",
		Code:    1,
		Name:    "Inconsistent indentation",
		Description: fmt.Sprintf(
			"You indent using %d %s, but this line is prefixed with %d %s.",
			indentLength,
			indentCharacterName(indentCharacter, indentLength),
			incorrectIndentLength,
			indentCharacterName(incorrectIndentCharacter, incorrectIndentLength),
		),
	}
}

var LexerFailed = &errors.Error{
	Section: "LEXER",
	Code:    2,
	Name:    "The lexer failed",
}
