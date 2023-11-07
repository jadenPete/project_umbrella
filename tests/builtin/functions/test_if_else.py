from tests import output_from_code

def test_then() -> None:
	assert output_from_code(
		"""\
fn then():
	"foo"

fn else_():
	"bar"

println(__if_else__(true, then, else_))
"""
	) == "foo\n"

def test_else() -> None:
	assert output_from_code(
		"""\
fn then():
	"foo"

fn else_():
	"bar"

println(__if_else__(false, then, else_))
"""
	) == "bar\n"

def _test_invalid_argument(code: str, i: int) -> None:
	assert output_from_code(code, expected_return_code=1) == f"""\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

__if_else__ expected argument #{i} to be of a different type.
"""

def test_invalid_conditional() -> None:
	_test_invalid_argument(
		"""\
fn do_nothing():

__if_else__(1, do_nothing, do_nothing)
""",
		1
	)

def test_invalid_branch() -> None:
	_test_invalid_argument(
		"""\
fn do_nothing():

__if_else__(true, unit, do_nothing)
""",
		2
	)

	_test_invalid_argument(
		"""\
fn do_nothing():

__if_else__(true, do_nothing, unit)
""",
		3
	)

def _test_invalid_branch_arity_case(code: str) -> None:
	assert output_from_code(code, expected_return_code=1) == \
		"Error (RUNTIME-1): A function accepting 1 argument was called with 0 arguments\n"

def test_invalid_branch_arity() -> None:
	_test_invalid_branch_arity_case(
		"""
fn do_nothing():
fn identity(value):
	value

println(__if_else__(true, identity, do_nothing))
"""
	)

	_test_invalid_branch_arity_case(
		"""
fn do_nothing():
fn identity(value):
	value

println(__if_else__(false, do_nothing, identity))
"""
	)
