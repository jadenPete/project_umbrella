from tests import output_from_code

def test_parenthesized_arithmetic() -> None:
	assert output_from_code("println((2 + 2) * 4)\n") == "16\n"
	assert output_from_code("println((4 - 2) / (2 * 2))\n") == "0\n"

def test_parenthesized_boolean_logic() -> None:
	assert \
		output_from_code("println((true && !false) || (true && false))\n") == \
		output_from_code("println(!(true && false))\n") == "true\n"

def test_parenthesized_calls() -> None:
	assert output_from_code("println(((1).+(1)))\n") == "2\n"
	assert output_from_code('(println)("Hello, world!")\n') == "Hello, world!\n"

def test_parenthesized_if_expressions() -> None:
	assert output_from_code(
		"""\
println((
	if true:
		"foo"
	else:
		"bar"
))
"""
	) == "foo\n"

	assert output_from_code(
		"""\
println(
	if(true):
		"foo"
	else:
		"bar"
)
"""
	) == "foo\n"

def test_parenthesized_primaries() -> None:
	assert output_from_code("println((0.5))\n") == "0.5\n"
	assert output_from_code('println((println))\n') == "(built-in function)\n"
	assert output_from_code("println((1))\n") == "1\n"
	assert output_from_code('println((("Hello, world!")))') == "Hello, world!\n"
	assert output_from_code('println(("foo") + "bar")\n') == "foobar\n"
	assert output_from_code('println(((true, false)))') == "(true, false)\n"

def test_parenthesized_selects() -> None:
	assert \
		output_from_code("println((println.__to_str__))\n") == \
		output_from_code("println((1).+)\n") == "(built-in function)\n"

	assert output_from_code("(1).(+)\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "."

  1  │ (1).(+)
     │    ^

"""

def test_invalid_parentheses() -> None:
	assert output_from_code(
		"""\
(
	fn do_nothing():
)
""",
		expected_return_code=1
	) == """\
Error (PARSER-1): The parser failed: unexpected token "("

  1  │ (
     │ ^

"""

	assert output_from_code("(meaning_of_life = 42)\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "("

  1  │ (meaning_of_life = 42)
     │ ^

"""

	assert output_from_code("(unit\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "("

  1  │ (unit
     │ ^

"""

	assert output_from_code("()\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "("

  1  │ ()
     │ ^

"""
