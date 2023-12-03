from tests import output_from_code

def test_length() -> None:
	assert output_from_code('println("".length)\n') == "0\n"
	assert output_from_code('println("0".length)\n') == "1\n"
	assert output_from_code('println("01".length)\n') == "2\n"

def test_plus() -> None:
	assert output_from_code(
		'println((("foo" + "") == "foo") && (("" + "foo") == "foo"))'
	) == "true\n"

	assert output_from_code(
		'println((("foo" + "bar") == "foobar") && (("bar" + "foo") == "barfoo"))'
	) == "true\n"

	assert output_from_code('"foo" + 0', expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

+ expected argument #1 to be of a different type.
"""
