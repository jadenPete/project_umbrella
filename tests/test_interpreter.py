from tests import output_from_code, output_from_filename

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

def test_select():
	assert set(output_from_filename("select.krait").splitlines()) == {
		"Hello, world!",
		"Hello, world!",
		"(built-in function)",
	}

def test_values():
	assert output_from_filename("value_alias.krait") == "Hello, world!\nHello, world!\n"
	assert output_from_filename("value_storing_function.krait") == "Hello, world!\n"
	assert set(output_from_filename("value.krait").splitlines()) == {
		"Hello, world!",
		"It's nice to see you."
	}
