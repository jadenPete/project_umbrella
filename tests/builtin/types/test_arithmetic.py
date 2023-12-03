from tests import output_from_code

def test_int_arithmetic() -> None:
	assert output_from_code("println(1 + 2 + 3)\n") == "6\n"
	assert output_from_code("println(1 - 2 - 3)\n") == "-4\n"
	assert output_from_code("println(-(2 + 2))\n") == "-4\n"
	assert output_from_code("println(1 * 2 * 3)\n") == "6\n"
	assert output_from_code("println(4 / 2 / 1)\n") == "2\n"
	assert output_from_code("println(81 % 12)\n") == "9\n"
	assert output_from_code("1 / 0\n", expected_return_code=1) == """\
Error (RUNTIME-7): Cannot divide by zero

Expected the right-hand side of int#/ to be nonzero.
"""

	assert output_from_code("1 % 0\n", expected_return_code=1) == """\
Error (RUNTIME-7): Cannot divide by zero

Expected the right-hand side of int#% to be nonzero.
"""

def test_float_arithmetic() -> None:
	assert output_from_code("println(1.1 + 2.2 + 3.3)\n") == "6.6\n"
	assert output_from_code("println(1.1 - 2.2 - 3.3)\n") == "-4.4\n"
	assert output_from_code("println(-(2.2 + 2.2))\n") == "-4.4\n"
	assert output_from_code("println(1.1 * 2.2 * 3.3)\n") == "7.986000000000001\n"
	assert output_from_code("println(4.4 / 2.2 / 1.1)\n") == "1.8181818181818181\n"
	assert output_from_code("println(81.81 % 12.12)\n") == "9.090000000000007\n"
	assert output_from_code("1.0 / 0.0\n", expected_return_code=1) == """\
Error (RUNTIME-7): Cannot divide by zero

Expected the right-hand side of float#/ to be nonzero.
"""

	assert output_from_code("1.0 % 0.0\n", expected_return_code=1) == """\
Error (RUNTIME-7): Cannot divide by zero

Expected the right-hand side of float#% to be nonzero.
"""

def test_float_formatting() -> None:
	assert output_from_code("println(1.)\n") == "1\n"
	assert output_from_code("println(.1)\n") == "0.1\n"
	assert output_from_code(".\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "."

  1  │ .
     │ ^

"""

def test_arithmetic_precedence() -> None:
	# Ensure the additive operators have equal precedence
	assert output_from_code("println(1 + 1 - 1)\n") == "1\n"
	assert output_from_code("println(1 - 1 + 1)\n") == "1\n"

	# Ensure prefix operations precede every other operation
	# (except the multiplicative ones because negation distributes over them)
	assert output_from_code("println(-1 + 1)\n") == "0\n"
	assert output_from_code("println(-1 - 1)\n") == "-2\n"

	# Ensure the multiplicative operators precede the additive operaotrs
	assert output_from_code("println(1 + 2 * 2)\n") == "5\n"

	# Ensure the multiplicative operators have equal precedence
	assert output_from_code("println(4 / 2 * 3)\n") == "6\n"
	assert output_from_code("println(4 * 2 / 4)\n") == "2\n"
	assert output_from_code("println(4 * 2 % 4)\n") == "0\n"
	assert output_from_code("println(4 % 2 * 4)\n") == "0\n"

def _test_strong_typing(*args: str) -> None:
	for operator in args:
		assert \
			output_from_code(f"1 {operator} 1.0\n", expected_return_code=1) == \
			output_from_code(f"1.0 {operator} 1\n", expected_return_code=1) == f"""\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

{operator} expected argument #1 to be of a different type.
"""

def test_arithmetic_strong_typing() -> None:
	_test_strong_typing("+", "-", "*", "/", "%")

def _test_comparison_case(
	left_hand_side: int,
	operator: str,
	right_hand_side: int,
	expected_value: str
) -> None:
	assert output_from_code(f"println({left_hand_side} {operator} {right_hand_side})\n") == \
		expected_value

	assert output_from_code(
		f"println({float(left_hand_side)} {operator} {float(right_hand_side)})\n"
	) == expected_value

def test_int_comparison() -> None:
	_test_comparison_case(1, "<", 2, "true\n")
	_test_comparison_case(1, "<", 1, "false\n")
	_test_comparison_case(1, "<=", 1, "true\n")
	_test_comparison_case(1, "<=", 0, "false\n")
	_test_comparison_case(1, ">", 0, "true\n")
	_test_comparison_case(1, ">", 1, "false\n")
	_test_comparison_case(1, ">=", 1, "true\n")
	_test_comparison_case(1, ">=", 2, "false\n")
	_test_comparison_case(1, "==", 1, "true\n")
	_test_comparison_case(1, "==", 2, "false\n")

def test_comparison_strong_typing() -> None:
	_test_strong_typing("<", "<=", ">", ">=")
