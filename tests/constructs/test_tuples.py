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
