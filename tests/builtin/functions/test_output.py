from tests import output_from_code

# NOTE: That `println` and `print` use the `__to_str__` method is tested in
# `tests.test_selects.test_to_str`.

def _test_output_function(code_template: str, expected_value: str) -> None:
	assert output_from_code(code_template.format("println")) == f"{expected_value}\n"
	assert output_from_code(code_template.format("print")) == expected_value

def test_basic() -> None:
	_test_output_function('{}("Hello, world!")\n', "Hello, world!")

def test_no_parameters() -> None:
	_test_output_function("{}()\n", "")

def test_multiple_parameters() -> None:
	_test_output_function('{}("Hello,", "how", "is", "your", "day?")\n', "Hello, how is your day?")
