/*
 * Matching:
 *
 * parser provides advancing matching capabilities supporting the construction of
 * abstract syntax trees (ASTs).
 *
 * Specifically, this functionality is encapsulated in the `ExhaustiveMatcher` struct, which extends
 * Go's regular expression engine by allowing `MatcherCode`s to substituted in.
 *
 * `MatcherCode`s correspond to tokens in the lexer and expressions in the parser. Under the hood,
 * they're replaced with Unicode characters, meaning they can be treated as ordinary characters in
 * the regular expressions passed `CompileMatcher`. Although this is a bit of hack, it allows us to
 * leverage the full capabilities of Go's regular expression engine, without having to manufacture a
 * full-fledged parser from scratch.
 *
 * `Matcher` objects returned by `CompilerMatcher` can be paired with `MatchCode`s, corresponding to
 * the tokens or expressions they match, to produce `ExhaustiveMatchPattern` objects which
 * collectively comprise an `ExhaustiveMatcher` object.
 *
 * An `ExhaustiveMatcher` can process input (`MatchInput`) that's either a raw string or compiled
 * from a sequence of `MatchCode`s via `CompileInput`; the latter is necessary when matching
 * higher-order expression patterns on tokens.
 *
 * Processing can be done linearly or hierarchically. The first strategy iteratively applies the
 * patterns on the input until it can be entirely partitioned into a list of matches
 * (`ExhaustiveMatch`es). The second is more attuned to parsing; it repeatedly cycles through the
 * patterns, attempting to build a tree of `ExhaustiveMatch`es until the entire entire input is a
 * single match on smaller, non-overlapping submatches.
 */
package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"

	"project_umbrella/interpreter/common"
)

type MatcherCode int
type MatcherInput string
type Matcher struct {
	regex *regexp.Regexp
}

// The first 32 Unicode characters frequently require special flags to match or aren't matchable in
// regular expressions. Hence, we use every character after that.
const matcherCodeOffset MatcherCode = 32

var escapedQuantifierRegex = regexp.MustCompile(`\{(\{.*\})\}`)
var codeSubstitionRegex = regexp2.MustCompile(`(?<!\{)\{(\d+)\}(?!\})`, 0)

/*
 * Given a string and the start and end indices of various substrings,
 * replace those substrings with provided replacements.
 *
 * This function's time complexity is `O(n + m)`, where `n = len(source)` and `m = len(ranges)`.
 */
func replaceStringRanges(source string, ranges [][2]int, replacements []string) string {
	rangeIndices := make(map[int]int, len(ranges))

	for i, indices := range ranges {
		rangeIndices[indices[0]] = i
	}

	var result strings.Builder

	i := 0

	for i < len(source) {
		if j, ok := rangeIndices[i]; ok {
			result.WriteString(replacements[j])

			i = ranges[j][1]
		} else {
			result.WriteByte(source[i])

			i++
		}
	}

	return result.String()
}

/*
 * Compile a sequence of matcher codes into input passable to an `ExhaustiveMatcher`.
 *
 * If a raw string is to be matched upon, this function should not be used and the string should be
 * directly coerced into `MatcherInput`.
 */
func CompileInput(input []MatcherCode) MatcherInput {
	offset := make([]rune, 0, len(input))

	for _, code := range input {
		offset = append(offset, rune(code+matcherCodeOffset))
	}

	return MatcherInput(string(offset))
}

/*
 * Compile an extended regular expression, including one or more substitution codes and
 * corresponding substitutions, into a `Matcher` object.
 *
 * See `parser.go` for an example of this method's usage.
 */
func CompileMatcher(regex string, substitutions ...MatcherCode) *Matcher {
	ranges := [][2]int{}
	replacements := []string{}

	match, _ := codeSubstitionRegex.FindStringMatch(regex)

	for match != nil {
		substitutedCodeCapture := match.GroupByNumber(1).Capture

		i, err := strconv.Atoi(substitutedCodeCapture.String())

		if err != nil {
			panic(err)
		}

		ranges = append(ranges, [2]int{match.Group.Index, match.Group.Index + match.Group.Length})
		replacements = append(
			replacements,
			fmt.Sprintf("\\x{%x}", substitutions[i]+matcherCodeOffset),
		)

		match, _ = codeSubstitionRegex.FindNextMatch(match)
	}

	substituted := replaceStringRanges(regex, ranges, replacements)
	unescaped := escapedQuantifierRegex.ReplaceAllString(substituted, "$1")
	compiled := regexp.MustCompile(unescaped)

	return &Matcher{
		regex: compiled,
	}
}

func (matcher *Matcher) FindAllSubmatchIndex(input MatcherInput, maximum int) [][]int {
	return matcher.regex.FindAllStringSubmatchIndex(string(input), maximum)
}

type ExhaustiveMatch struct {
	Type MatcherCode

	// The start index of the match in the input.
	Start int

	// The end index of the match in the input.
	End int

	/*
	 * The match subgroups as captured by Go's regular expression engine.
	 *
	 * Note that these are not relative to the input, but to the start of the match. Furthermore, in
	 * hierarchical matching, these will be in terms of this match's children.
	 */
	Subgroups [][2]int
}

type ExhaustiveMatchPattern struct {
	Type    MatcherCode
	Matcher *Matcher
}

type ExhaustiveMatcher struct {
	Patterns []*ExhaustiveMatchPattern
}

const UnrecognizedMatcherCode MatcherCode = 0

func flattenedExhaustiveMatchTree(
	tree *common.BinaryTree[*ExhaustiveMatch],
) ([]*ExhaustiveMatch, bool) {
	result := []*ExhaustiveMatch{}

	if _, ok := tree.DFS(func(node *common.BinaryTree[*ExhaustiveMatch]) bool {
		if node.Value == nil {
			return false
		}

		if node.Value.Type == UnrecognizedMatcherCode {
			return true
		}

		result = append(result, node.Value)

		return false
	}); ok {
		return nil, false
	}

	return result, true
}

/*
 * `Match` and `MatchWithInitial` perform linear matching as described in the package description.
 *
 * If the input couldn't be exhaustively matched against, they return `nil`.
 */
func (matcher *ExhaustiveMatcher) Match(input MatcherInput) ([]*ExhaustiveMatch, bool) {
	return matcher.MatchWithInitial(input, []*ExhaustiveMatch{
		{
			Type:  UnrecognizedMatcherCode,
			Start: 0,
			End:   len(input),
		},
	})
}

func (matcher *ExhaustiveMatcher) MatchWithInitial(
	input MatcherInput,
	initialMatches []*ExhaustiveMatch,
) ([]*ExhaustiveMatch, bool) {
	tree := common.NewBalancedBinaryTreeFromSlice(initialMatches)
	stack := []*common.BinaryTree[*ExhaustiveMatch]{}

	appendUnrecognizedMatchesToStack := func(node *common.BinaryTree[*ExhaustiveMatch]) {
		node.DFS(func(node *common.BinaryTree[*ExhaustiveMatch]) bool {
			if node.Value != nil && node.Value.Type == UnrecognizedMatcherCode {
				stack = append(stack, node)
			}

			return false
		})
	}

	appendUnrecognizedMatchesToStack(tree)

	for len(stack) > 0 {
		node := stack[len(stack)-1]

		stack = stack[:len(stack)-1]

		replacements := []*ExhaustiveMatch{}
		lastMatchEnd := 0

		for _, pattern := range matcher.Patterns {
			type_, matcher := pattern.Type, pattern.Matcher

			for _, match := range matcher.FindAllSubmatchIndex(
				input[node.Value.Start:node.Value.End],
				-1,
			) {
				if match[0] > lastMatchEnd {
					replacements = append(replacements, &ExhaustiveMatch{
						Type:  UnrecognizedMatcherCode,
						Start: node.Value.Start + lastMatchEnd,
						End:   node.Value.Start + match[0],
					})
				}

				subgroups := [][2]int{}

				for i := 2; i < len(match); i += 2 {
					subgroups = append(subgroups, [2]int{match[i], match[i+1]})
				}

				replacements = append(replacements, &ExhaustiveMatch{
					Type:      type_,
					Start:     node.Value.Start + match[0],
					End:       node.Value.Start + match[1],
					Subgroups: subgroups,
				})

				lastMatchEnd = match[1]
			}

			if len(replacements) > 0 {
				break
			}
		}

		if len(replacements) == 0 {
			return nil, false
		}

		if node.Value.Start+lastMatchEnd < node.Value.End {
			replacements = append(replacements, &ExhaustiveMatch{
				Type:  UnrecognizedMatcherCode,
				Start: node.Value.Start + lastMatchEnd,
				End:   node.Value.End,
			})
		}

		*node = *common.NewBalancedBinaryTreeFromSlice(replacements)

		appendUnrecognizedMatchesToStack(node)
	}

	return flattenedExhaustiveMatchTree(tree)
}
