from tests import output_from_code

def test_functions() -> None:
	assert output_from_code(
		"""\
fn say_hello():
	println("Hello, world!")

say_hello()
"""
	) == "Hello, world!\n"

	assert output_from_code(
		"""\
fn say_hello(name):
	println("Hello, " + name + "!")

say_hello("Jaden")
"""
	) == "Hello, Jaden!\n"

	assert output_from_code(
		"""\
fn say_hello(first_name, last_name):
	println("Hello, " + first_name + " " + last_name + "!")

say_hello("Jaden", "Peterson")
"""
	) == "Hello, Jaden Peterson!\n"

	assert output_from_code(
		"""\
fn do_nothing():

do_nothing()
"""
	) == ""

def test_formatting() -> None:
	assert output_from_code(
		"""\
fn identity1(value):
	value

fn identity2(
	value
):
	value

fn identity3
(
	value
):
	value

println((identity1(unit) == unit) && (identity2(unit) == unit) && (identity3(unit) == unit))
"""
	) == "true\n"

def test_invalid_functions() -> None:
	assert output_from_code("fn do_nothing:\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "fn"

  1  │ fn do_nothing:
     │ ^^

"""

	assert output_from_code("fn do_nothing()", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "fn"

  1  │ fn do_nothing()
     │ ^^

"""

def test_call_validation() -> None:
	assert output_from_code(
		"""
fn do_nothing(dummy):

do_nothing()
""",
		expected_return_code=1
	) == "Error (RUNTIME-1): A function accepting 1 argument was called with 0 arguments\n"

	assert output_from_code(
		"""
fn do_nothing():

do_nothing(unit)
""",
		expected_return_code=1
	) == "Error (RUNTIME-1): A function accepting 0 arguments was called with 1 arguments\n"

def test_function_return_values() -> None:
	assert output_from_code(
		"""\
fn greeting(name):
	"Hello, " + name + "!"

println(greeting("Jaden"))
"""
	) == "Hello, Jaden!\n"

	assert output_from_code(
		"""\
fn black_hole(value):

println(black_hole("space ship"))
"""
	) == "(unit)\n"

def test_function_scope_evasion() -> None:
	assert output_from_code(
		"""\
fn greeting_factory(verb, excited):
	suffix = if excited:
		"!"
	else:
		"."

	fn greeting(name):
		verb + ", " + name + suffix

	greeting

println(
	(greeting_factory("Hello", false)("Jaden") == "Hello, Jaden.") &&
		(greeting_factory("Hey", true)("Jaden") == "Hey, Jaden!")
)
"""
	) == "true\n"

def test_recursion() -> None:
	assert output_from_code(
		"""\
fn triangular_number(n):
	if n <= 0:
		0
	else:
		n + triangular_number(n - 1)

println(triangular_number(3))
"""
	) == "6\n"

def test_nested_functions() -> None:
	assert output_from_code(
		"""\
fn greeter():
	fn greet():
		println("Hello, world!")

	greet

greeter()()
"""
	) == "Hello, world!\n"

def test_first_class_functions() -> None:
	assert output_from_code(
		"""\
puts = print
puts("Hello, world!")
"""
	) == "Hello, world!"

def test_higher_order_functions() -> None:
	assert output_from_code(
		"""\
fn and_then(inner, outer):
	fn result(argument):
		outer(inner(argument))

	result

fn negate(number):
	-number

decrement = and_then((1).-, negate)

println((decrement(1) == 0) && (decrement(0) == -1) && (decrement(-1) == -2))
"""
	) == "true\n"

def test_function_shadowing() -> None:
	assert output_from_code(
		"""\
fn do_nothing():
fn do_nothing():
""",
		expected_return_code=1
	) == """\
Error (PARSER-5): Reassigning to an already declared value is impossible

  1  │ fn do_nothing():
  2  │ fn do_nothing():
     │    ^^^^^^^^^^

Consider assigning to a new value.
"""
