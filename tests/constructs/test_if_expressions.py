from tests import output_from_code

def test_if_expressions() -> None:
	assert output_from_code(
		"""\
if ("foo" + "bar") == "foobar":
	println(true)
"""
	) == "true\n"

	assert output_from_code("if true:\n") == ""
	assert output_from_code("if:\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "if"

  1  │ if:
     │ ^^

"""

	assert output_from_code("if true\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "if"

  1  │ if true
     │ ^^

"""

def test_if_else_expressions() -> None:
	assert output_from_code(
		"""\
if ("foo" + "bar") == "foobar":
	println(true)
else:
	println(false)
"""
	) == "true\n"

	assert output_from_code(
		"""\
if ("foo" + "bar") == "bizz":
	println(true)
else:
	println(false)
"""
	) == "false\n"

	assert output_from_code(
		"""\
if ("foo" + "bar") == "foobar":
	println(true)
else:
"""
	) == "true\n"


	assert output_from_code(
		"""\
if false:
else:
else:
""",
		expected_return_code=1
	) == """\
Error (PARSER-1): The parser failed: unexpected token "else"

  1  │ if false:
  2  │ else:
  3  │ else:
     │ ^^^^

"""

	assert output_from_code(
		"""\
if false:
else
""",
		expected_return_code=1
	) == """\
Error (PARSER-1): The parser failed: unexpected token "else"

  1  │ if false:
  2  │ else
     │ ^^^^

"""

def test_if_else_if_expressions() -> None:
	assert output_from_code(
		"""\
if true:
	println(1)
else if true:
	println(2)
else:
	println(3)
"""
	) == "1\n"

	assert output_from_code(
		"""\
if false:
	println(1)
else if true:
	println(2)
else if true:
	println(3)
else:
	println(4)
"""
	) == "2\n"

	assert output_from_code(
		"""\
if false:
	println(1)
else if false:
	println(2)
else if true:
	println(3)
else:
	println(4)
"""
	) == "3\n"

	assert output_from_code(
		"""\
if false:
	println(1)
else if false:
	println(2)
else:
	println(3)
"""
	) == "3\n"

	assert output_from_code(
		"""\
if false:
else if:
""",
		expected_return_code=1
	) == """\
Error (PARSER-1): The parser failed: unexpected token "else"

  1  │ if false:
  2  │ else if:
     │ ^^^^

"""

	assert output_from_code(
		"""\
if false:
else if true
""",
		expected_return_code=1
	) == """\
Error (PARSER-1): The parser failed: unexpected token "else"

  1  │ if false:
  2  │ else if true
     │ ^^^^

"""

def test_formatting() -> None:
	assert output_from_code(
		"""\
if
	true
:
	println("Hello, world!")
"""
	) == "Hello, world!\n"

	assert output_from_code(
		"""\
if false:
else
if
	true
:
	println("Hello, world!")
"""
	) == "Hello, world!\n"

	assert output_from_code(
		"""\
if false:
else if false:
else
:
	println("Hello, world!")
"""
	) == "Hello, world!\n"

def test_invalid_conditions() -> None:
	code1 = "if 0:\n"
	code2 = """\
if false:
else if 0:
"""

	assert \
		output_from_code(code1, expected_return_code=1) == \
		output_from_code(code2, expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

__if_else__ expected argument #1 to be of a different type.
"""

def test_if_expression_values() -> None:
	assert output_from_code(
		"""\
println(
	if true:
		"foo"
)
"""
	) == "foo\n"

	assert output_from_code("println(if true:)\n") == "(unit)\n"
	assert output_from_code(
		"""\
println(
	if false:
		"foo"
)
"""
	) == "(unit)\n"

	assert output_from_code(
		"""\
println(
	if true:
		"foo"
	else:
		"bar"
)
"""
	) == "foo\n"

	assert output_from_code(
		"""\
println(
	if false:
		"foo"
	else:
		"bar"
)
"""
	) == "bar\n"

	assert output_from_code(
		"""\
println(
	if false:
	else:
)
"""
	) == "(unit)\n"

	assert output_from_code(
		"""\
println(
	if true:
		"foo"
	else if true:
		"bar"
	else:
		"bizz"
)
"""
	) == "foo\n"

	assert output_from_code(
		"""\
println(
	if false:
		"foo"
	else if true:
		"bar"
	else if true:
		"bizz"
	else:
		"buzz"
)
"""
	) == "bar\n"

	assert output_from_code(
		"""\
println(
	if false:
		"foo"
	else if false:
		"bar"
	else if true:
		"bizz"
	else:
		"buzz"
)
"""
	) == "bizz\n"

	assert output_from_code(
		"""\
println(
	if false:
		"foo"
	else if false:
		"bar"
	else:
		"bizz"
)
"""
	) == "bizz\n"

def test_if_expression_value_formatting() -> None:
	assert output_from_code("println(if true:)\n") == "(unit)\n"
	assert output_from_code(
		"""\
println(if true:
else:
)
"""
	) == "(unit)\n"

	assert output_from_code(
		"""\
println(if true:
else if true:
else:
)
"""
	) == "(unit)\n"
