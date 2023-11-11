from tests import output_from_code

def test_value_cycle_detection() -> None:
	code1 = """\
fn foo():
	bar()

fn bar():
	foo()
"""

	code2 = """\
fn foo():
	bar()

fn bar():
	bizz()

fn bizz():
	foo()
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
