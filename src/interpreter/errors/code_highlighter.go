package errors

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"project_umbrella/interpreter/common"
)

const contextLines = 3
const lineNumberMinimumLength = 2
const lineNumberRightPadding = 2
const tabSpaces = 4

type context struct {
	contextStartLine    int
	contextLines        []string
	highlightedPosition *tabAdjustedPosition
}

/*
 * `tabAdjustedPosition` is one-indexed and inclusive on both ends. Unlike `errors.Position`, it
 * represents the error range as a line-column pair.
 */
type tabAdjustedPosition struct {
	startLine   int
	startColumn int
	endLine     int
	endColumn   int
}

func newTabAdjustedPosition(lines []string, position *Position) *tabAdjustedPosition {
	adjustedColumn := func(line int, column int, isStart bool) int {
		beforeColumnLength := column

		if isStart {
			beforeColumnLength--
		}

		beforeColumn := string([]rune(lines[line-1])[:beforeColumnLength])
		tabCountBeforeColumn := strings.Count(beforeColumn, "\t")

		return column + tabCountBeforeColumn*(tabSpaces-1)
	}

	offset := 0
	startLine := 1

	for offset+len(lines[startLine-1]) < position.Start {
		offset += len(lines[startLine-1]) + 1
		startLine++
	}

	startColumn := adjustedColumn(startLine, position.Start-offset+1, true)
	endLine := startLine

	for offset+len(lines[endLine-1]) < position.End {
		offset += len(lines[endLine-1]) + 1
		endLine++
	}

	endColumn := adjustedColumn(endLine, position.End-offset, false)

	return &tabAdjustedPosition{
		startLine:   startLine,
		startColumn: startColumn,
		endLine:     endLine,
		endColumn:   endColumn,
	}
}

func newContext(adjustedLines []string, position *tabAdjustedPosition) *context {
	contextStartLine := common.Max(position.startLine-contextLines, 1)

	return &context{
		contextStartLine:    contextStartLine,
		contextLines:        adjustedLines[contextStartLine-1 : position.endLine],
		highlightedPosition: position,
	}
}

func (context_ *context) formattedLineNumber(lineNumber string) string {
	lineNumberLength := common.Max(
		len(strconv.Itoa(context_.highlightedPosition.endLine)),
		lineNumberMinimumLength,
	)

	return fmt.Sprintf(
		" %*s%s│ ",
		lineNumberLength,
		lineNumber,
		strings.Repeat(" ", lineNumberRightPadding),
	)
}

func (context_ *context) String(lineDelimiterFromLineNumber func(int) string) string {
	var result strings.Builder

	for i, line := range context_.contextLines {
		lineNumber := context_.contextStartLine + i
		lineDelimiter := ""

		if lineDelimiterFromLineNumber != nil {
			lineDelimiter = lineDelimiterFromLineNumber(lineNumber)
		}

		result.WriteString(context_.formattedLineNumber(strconv.Itoa(lineNumber)))
		result.WriteString(fmt.Sprintf("%s%s\n", line, lineDelimiter))
	}

	return result.String()
}

func tabAdjustedCodeLines(lines []string) []string {
	adjustedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		adjustedLines = append(adjustedLines, strings.ReplaceAll(line, "\t", strings.Repeat(" ", tabSpaces)))
	}

	return adjustedLines
}

func highlightedCode(code string, position *Position) string {
	lines := strings.Split(code, "\n")
	adjustedLines := tabAdjustedCodeLines(lines)
	adjustedPosition := newTabAdjustedPosition(lines, position)
	context_ := newContext(adjustedLines, adjustedPosition)
	emptyLineNumber := context_.formattedLineNumber("")

	if adjustedPosition.startLine == adjustedPosition.endLine {
		return fmt.Sprintf(
			"%s%s%s%s\n",
			context_.String(nil),
			emptyLineNumber,
			strings.Repeat(" ", adjustedPosition.startColumn-1),
			strings.Repeat("^", adjustedPosition.endColumn-adjustedPosition.startColumn+1),
		)
	}

	maximumContextLineLength := math.MinInt

	for _, line := range context_.contextLines {
		lineLength := len([]rune(line))

		if lineLength < maximumContextLineLength {
			maximumContextLineLength = lineLength
		}
	}

	belowContextLeftPadding := strings.Repeat(" ", adjustedPosition.endColumn-1)
	belowContextRightPaddingLength := maximumContextLineLength - adjustedPosition.endColumn + 3

	return fmt.Sprintf(
		"%[1]s%[2]s%[3]s^%[4]s║\n%[2]s%[3]s╘%[5]s╝\n",
		context_.String(func(lineNumber int) string {
			lineLength := len([]rune(context_.contextLines[lineNumber-context_.contextStartLine]))
			padding := strings.Repeat(" ", maximumContextLineLength-lineLength+1)

			if lineNumber == context_.contextStartLine {
				return fmt.Sprintf("%s<─╖", padding)
			}

			return fmt.Sprintf("%s  ║", padding)
		}),

		emptyLineNumber,
		belowContextLeftPadding,
		strings.Repeat(" ", belowContextRightPaddingLength),
		emptyLineNumber,
		strings.Repeat("═", belowContextRightPaddingLength),
	)
}
