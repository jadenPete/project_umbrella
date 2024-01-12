from tests import output_from_code

def test_empty() -> None:
	assert output_from_code("println((,))\n") == "(,)\n"

def test_one_element() -> None:
	assert output_from_code("println((0,))\n") == "(0,)\n"

def test_multiple_elements() -> None:
	assert output_from_code("println((0, 1))\n") == "(0, 1)\n"

def test_formatting() -> None:
	assert output_from_code(
		"""\
println(
	(
		,
	)
)
"""
	) == "(,)\n"

	assert output_from_code(
		"""\
println(
	(
		0
			,
	)
)
"""
	) == "(0,)\n"

	assert output_from_code(
		"""\
println(
	(
		0
			,
		1
	)
)
"""
	) == "(0, 1)\n"

def test_invalid_tuples() -> None:
	assert output_from_code("()\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "("

  1  │ ()
     │ ^

"""

	assert output_from_code("println((0))\n") == "0\n"
	assert output_from_code("println((0) == 0)\n") == "true\n"

def test_get() -> None:
	assert output_from_code('println(("foo",).get(0))\n') == "foo\n"
	assert output_from_code('println(("foo", "bar").get(1))\n') == "bar\n"
	assert output_from_code('println(("foo", "bar").get(2))\n', expected_return_code=1) == """\
Error (RUNTIME-14): An out-of-bounds index was provided to tuple#get

Expected an index in the range [0, 2), but got 2.
"""

	assert output_from_code('println(("foo", "bar").get(-1))\n', expected_return_code=1) == """\
Error (RUNTIME-14): An out-of-bounds index was provided to tuple#get

Expected an index in the range [0, 2), but got -1.
"""

def test_length() -> None:
	assert output_from_code("println((,).length)\n") == "0\n"
	assert output_from_code("println((0,).length)\n") == "1\n"
	assert output_from_code("println((0, 1).length)\n") == "2\n"

def test_plus() -> None:
	assert output_from_code("println((((0,) + (,)) == (0,)) && (((,) + (0,)) == (0,)))\n") == \
		"true\n"

	assert output_from_code("println((((0,) + (1,)) == (0, 1)) && (((1,) + (0,)) == (1, 0)))") == \
		"true\n"

	assert output_from_code(
		"println((((0,) + (1, 2)) == (0, 1, 2)) && (((1, 2) + (0,)) == (1, 2, 0)))"
	) == "true\n"

	assert output_from_code("(,) + 0", expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

+ expected argument #1 to be of a different type.
"""

def test_slice() -> None:
	assert output_from_code("println((1, 2, 3).slice(0, 3))\n") == "(1, 2, 3)\n"
	assert output_from_code("println((1, 2, 3).slice(1, 2))\n") == "(2,)\n"
	assert output_from_code("println((1, 2, 3).slice(1, 1))\n") == "(,)\n"
	assert output_from_code("println((1, 2, 3).slice(1, 0))\n") == "(,)\n"
	assert output_from_code("println((1, 2, 3).slice(-1, 2))\n") == "(1, 2)\n"
	assert output_from_code("println((1, 2, 3).slice(1, 4))\n") == "(2, 3)\n"
	assert output_from_code("println((1, 2, 3).slice(0))\n", expected_return_code=1) == \
		"Error (RUNTIME-1): A function accepting 2 arguments was called with 1 arguments\n"

	assert output_from_code('println((1, 2, 3).slice("0", "3"))\n', expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

slice expected argument #1 to be of a different type.
"""

def test_times() -> None:
	assert output_from_code('println((1,) * 4)\n') == "(1, 1, 1, 1)\n"
	assert output_from_code('println((1, 2, 3) * 3)\n') == "(1, 2, 3, 1, 2, 3, 1, 2, 3)\n"
	assert output_from_code('println((,) * 4)\n') == "(,)\n"
	assert output_from_code('println((1,) * -1)\n') == "(,)\n"
	assert output_from_code('println((1,) * "0")\n', expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

* expected argument #1 to be of a different type.
"""
