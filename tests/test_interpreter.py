from tests import has_expected_lines, output_from_code, output_from_filename

def test_arithmetic():
	assert has_expected_lines(
		output_from_filename("arithmetic.krait"),
		["4", "10", "4.5", "5"] +
		["1", "1"] +
		["6", "8.75"] +
		["0", "3.2857142857142856"]
	)

	assert output_from_filename("arithmetic_precedence.krait") == "3.5\n"

def test_calls():
	assert output_from_filename("call.krait") == "Hello, world!\n"
	assert \
		output_from_filename("call_nested.krait") == \
		output_from_filename("call_return_value.krait") == "Hello, world!\n(unit)\n"

	assert set(output_from_filename("call_with_multiple_arguments.krait").split("\n")) == {
		"Hello, world!",
		""
	}

def test_empty():
	assert output_from_code("") == ""
	assert output_from_code("\n") == ""

def test_functions():
	assert \
		output_from_filename("function_basic.krait") == \
		output_from_filename("function_nested.krait") == "Hello, world!\n"

def test_select():
	assert has_expected_lines(output_from_filename("select.krait"), [
		"Hello, world!",
		"Hello, world!",
		"(built-in function)",
	])

def test_values():
	assert output_from_filename("value_alias.krait") == "Hello, world!\nHello, world!\n"
	assert output_from_filename("value_storing_function.krait") == "Hello, world!\n"
	assert has_expected_lines(output_from_filename("value.krait"), [
		"Hello, world!",
		"It's nice to see you."
	])
