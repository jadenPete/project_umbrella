from tests import output_from_code

def test_not() -> None:
	assert output_from_code("println(!true)\n") == "false\n"
	assert output_from_code("println(!false)\n") == "true\n"

def test_and() -> None:
	assert output_from_code("println(true && true)\n") == "true\n"
	assert output_from_code("println(true && false)\n") == "false\n"
	assert output_from_code("println(false && true)\n") == "false\n"
	assert output_from_code("println(false && false)\n") == "false\n"

def test_or() -> None:
	assert output_from_code("println(true || true)\n") == "true\n"
	assert output_from_code("println(true || false)\n") == "true\n"
	assert output_from_code("println(false || true)\n") == "true\n"
	assert output_from_code("println(false || false)\n") == "false\n"

def test_strong_typing() -> None:
	for operator in ["&&", "||"]:
		assert output_from_code(f"true {operator} 0\n", expected_return_code=1) == f"""\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

{operator} expected argument #1 to be of a different type.
"""
