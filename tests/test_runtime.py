from tests import output_from_code

def test_value_cycle_detection() -> None:
	code1 = """\
result = foo(0)

fn foo(number):
	if number >= 10:
		result
	else:
		bar(number + 1)

fn bar(number):
	foo(number + 1)
"""

	code2 = """\
result = foo(0)

fn foo(number):
	if number >= 10:
		result
	else:
		bar(number + 1)

fn bar(number):
	bizz(number + 1)

fn bizz(number):
	foo(number + 1)
"""

	code3 = """\
foo = bar_getter()
bar = foo

fn bar_getter():
	bar
"""

	assert \
		output_from_code(code1, expected_return_code=1) == \
		output_from_code(code2, expected_return_code=1) == \
		output_from_code(code3, expected_return_code=1) == \
			"Error (RUNTIME-5): Encountered a cycle between values\n"
