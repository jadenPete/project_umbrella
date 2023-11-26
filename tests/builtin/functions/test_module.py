from tests import output_from_code

def test_basic() -> None:
	assert output_from_code(
		"""\
message = "Hello, world!"
module = __module__((("message", message),))

println(module.message)
"""
	) == "Hello, world!\n"

	assert output_from_code(
		"""\
message = "Hello, world!"

fn print_message():
	println(message)

module = __module__(
	(
		("message", message),
		("print_message", print_message)
	)
)

module.print_message()
"""
	) == "Hello, world!\n"

def test_nonexistent_fields() -> None:
	assert output_from_code("__module__((,)).foo\n", expected_return_code=1) == \
		"Error (RUNTIME-9): Unknown field: `foo`\n"

def test_strong_typing_argument() -> None:
	assert \
		output_from_code('__module__("foo")\n', expected_return_code=1) == \
		output_from_code('__module__(("foo", "bar"))\n', expected_return_code=1) == \
		output_from_code('__module__((("foo",),))\n', expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

__module__ expected argument #1 to be of a different type.
"""

def test_strong_typing_return_function() -> None:
	assert output_from_code(
		"""\
module = __module__((("foo", "bar"),))
module(0)
""",
		expected_return_code=1
	) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

(built-in function) expected argument #1 to be of a different type.
"""
