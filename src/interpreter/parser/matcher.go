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
	rangeIndices := make(map[int]int)

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
	ranges := make([][2]int, 0)
	replacements := make([]string, 0)

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
	type_ MatcherCode

	// The start index of the match in the input.
	start int

	// The end index of the match in the input.
	end int

	/*
	 * The match subgroups as captured by Go's regular expression engine.
	 *
	 * Note that these are not relative to the input, but to the start of the match. Furthermore, in
	 * hierarchical matching, these will be in terms of this match's children.
	 */
	subgroups [][2]int
}

type ExhaustiveMatchPattern struct {
	type_   MatcherCode
	matcher *Matcher
}

type ExhaustiveMatcher struct {
	patterns []*ExhaustiveMatchPattern
}

const unrecognizedCode MatcherCode = 0

func compiledExhaustiveMatchTreeArray(uncompiled []*common.Tree[*ExhaustiveMatch]) MatcherInput {
	input := make([]MatcherCode, 0, len(uncompiled))

	for _, tree := range uncompiled {
		input = append(input, tree.Value.type_)
	}

	return CompileInput(input)
}

func flattenedExhaustiveMatchTree(tree *common.BinaryTree[*ExhaustiveMatch]) []*ExhaustiveMatch {
	result := make([]*ExhaustiveMatch, 0)

	if tree.DFS(func(node *common.BinaryTree[*ExhaustiveMatch]) bool {
		if node.Value == nil {
			return false
		}

		if node.Value.type_ == unrecognizedCode {
			return true
		}

		result = append(result, node.Value)

		return false
	}) != nil {
		return nil
	}

	return result
}

/*
 * Perform linear matching as described in the package description.
 *
 * If the input couldn't be exhaustively matched against, this function returns `nil`.
 */
func (matcher *ExhaustiveMatcher) Match(input MatcherInput) []*ExhaustiveMatch {
	tree := &common.BinaryTree[*ExhaustiveMatch]{
		Value: &ExhaustiveMatch{
			type_: unrecognizedCode,
			start: 0,
			end:   len(input),
		},
	}

	stack := []*common.BinaryTree[*ExhaustiveMatch]{tree}

	for len(stack) > 0 {
		node := stack[len(stack)-1]

		stack = stack[:len(stack)-1]

		replacements := make([]*ExhaustiveMatch, 0)
		lastMatchEnd := 0

		for _, pattern := range matcher.patterns {
			type_, matcher := pattern.type_, pattern.matcher

			for _, match := range matcher.FindAllSubmatchIndex(
				input[node.Value.start:node.Value.end],
				-1,
			) {
				if match[0] > lastMatchEnd {
					replacements = append(replacements, &ExhaustiveMatch{
						type_: unrecognizedCode,
						start: node.Value.start + lastMatchEnd,
						end:   node.Value.start + match[0],
					})
				}

				subgroups := make([][2]int, 0)

				for i := 2; i < len(match); i += 2 {
					subgroups = append(subgroups, [2]int{match[i], match[i+1]})
				}

				replacements = append(replacements, &ExhaustiveMatch{
					type_:     type_,
					start:     node.Value.start + match[0],
					end:       node.Value.start + match[1],
					subgroups: subgroups,
				})

				lastMatchEnd = match[1]
			}

			if len(replacements) > 0 {
				break
			}
		}

		if len(replacements) == 0 {
			return nil
		}

		if node.Value.start+lastMatchEnd < node.Value.end {
			replacements = append(replacements, &ExhaustiveMatch{
				type_: unrecognizedCode,
				start: node.Value.start + lastMatchEnd,
				end:   node.Value.end,
			})
		}

		*node = *common.NewBalancedBinaryTreeFromSlice(replacements)

		node.DFS(func(descendent *common.BinaryTree[*ExhaustiveMatch]) bool {
			if descendent.Value != nil && descendent.Value.type_ == unrecognizedCode {
				stack = append(stack, descendent)
			}

			return false
		})
	}

	return flattenedExhaustiveMatchTree(tree)
}

/*
 * Perform hierarchical matching as described in the package description.
 *
 * If the input couldn't be exhaustively matched against, this function returns `nil`.
 */
func (matcher *ExhaustiveMatcher) MatchTree(input []MatcherCode) *common.Tree[*ExhaustiveMatch] {
	unmatched := make([]*common.Tree[*ExhaustiveMatch], 0, len(input))

	for i, code := range input {
		unmatched = append(unmatched, &common.Tree[*ExhaustiveMatch]{
			Children: []*common.Tree[*ExhaustiveMatch]{},
			Value: &ExhaustiveMatch{
				type_:     code,
				start:     i,
				end:       i + 1,
				subgroups: make([][2]int, 0),
			},
		})
	}

	for {
		var compiled MatcherInput

		changed := false
		recompile := true

		for _, pattern := range matcher.patterns {
			if recompile {
				compiled = compiledExhaustiveMatchTreeArray(unmatched)
				recompile = false
			}

			squashed := make([]*common.Tree[*ExhaustiveMatch], 0)

			i := 0

			for _, match := range pattern.matcher.FindAllSubmatchIndex(compiled, -1) {
				changed = true
				recompile = true

				subgroups := make([][2]int, 0)

				for i := 2; i < len(match); i += 2 {
					if match[i] == -1 {
						subgroups = append(subgroups, [2]int{-1, -1})
					} else {
						subgroups = append(
							subgroups,
							[2]int{match[i] - match[0], match[i+1] - match[0]},
						)
					}
				}

				var start int
				var end int

				if len(unmatched) > 0 {
					start = unmatched[match[0]].Value.start

					if match[0] == match[1] {
						end = start
					} else {
						end = unmatched[match[1]-1].Value.end
					}
				}

				squashed = append(squashed, unmatched[i:match[0]]...)
				squashed = append(squashed, &common.Tree[*ExhaustiveMatch]{
					Children: unmatched[match[0]:match[1]],
					Value: &ExhaustiveMatch{
						type_:     pattern.type_,
						start:     start,
						end:       end,
						subgroups: subgroups,
					},
				})

				i = match[1]
			}

			if recompile {
				unmatched = append(squashed, unmatched[i:]...)
			}
		}

		if !changed {
			break
		}
	}

	if len(unmatched) != 1 {
		return nil
	}

	return unmatched[0]
}
