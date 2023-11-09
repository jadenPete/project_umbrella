from tests import output_from_code

def test_selects() -> None:
	assert output_from_code("println((1).__to_str__)\n") == "(built-in function)\n"
	assert output_from_code("println((1).+)\n") == "(built-in function)\n"

def test_nonexistent_fields() -> None:
	assert output_from_code('"Hello, world!".foo\n', expected_return_code=1) == """\
Error (PARSER-7): Unknown field: `foo`

  1  │ "Hello, world!".foo
     │                 ^^^

"""

def test_invalid_selects() -> None:
	assert output_from_code("println.\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "."

  1  │ println.
     │        ^

"""

	assert output_from_code("println..\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "."

  1  │ println..
     │        ^

"""

	assert output_from_code(".__str__\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "."

  1  │ .__str__
     │ ^

"""

def test_to_str() -> None:
	assert output_from_code("println(true.__to_str__())\n") == "true\n"
	assert output_from_code("println(true)\n") == "true\n"
	assert output_from_code("println(0.5)\n") == "0.5\n"
	assert output_from_code("println(println)\n") == "(built-in function)\n"
	assert output_from_code(
		"""\
fn do_nothing():

println(do_nothing)
"""
	) == "(function)\n"

	assert output_from_code("println(1)\n") == "1\n"
	assert output_from_code('println("Hello, world!")\n') == "Hello, world!\n"
	assert output_from_code("println(unit)") == "(unit)\n"

def _test_equals_case(value: str, different: str) -> None:
	assert output_from_code(
		f"""\
value = {value}
different = {different}

println((value == value) && !(value != value) && !(value == different) && (value != different))
"""
	) == "true\n"

def test_equals() -> None:
	_test_equals_case("true", "false")
	_test_equals_case("1.0", "0.5")
	_test_equals_case("1.0", "1")
	_test_equals_case("println", "print")

	assert output_from_code(
		"""\
fn do_nothing1():
fn do_nothing2():

println(do_nothing1 == do_nothing2)
"""
	) == "false\n"

	_test_equals_case("1", "2")
	_test_equals_case('"foo"', '"bar"')

	assert output_from_code("println(unit == unit)") == "true\n"
	assert output_from_code("println(unit != unit)") == "false\n"
