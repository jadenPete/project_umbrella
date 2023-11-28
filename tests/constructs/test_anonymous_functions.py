from tests import output_from_code

def test_basic():
	assert output_from_code(
		"""\
(():
	println("Hello, world!")
)()
"""
	) == "Hello, world!\n"

	assert output_from_code(
		"""\
((name):
	println("Hello, " + name + "!")
)("Jaden")
"""
	) == "Hello, Jaden!\n"

	assert output_from_code(
		"""\
((greeting, name):
	println(greeting + ", " + name + "!")
)("Hi", "Jaden")
"""
	) == "Hi, Jaden!\n"

	assert output_from_code(
		"""\
add = (x, y): x + y

println(add(1, 2))
"""
	) == "3\n"

	assert output_from_code("println((():)())") == "(unit)\n"

def test_formatting():
	assert output_from_code(
		"""\
multiply = (
	a,
	b
):
	a * b

println(multiply(2, 3))
"""
	) == "6\n"


	assert output_from_code(
		"""\
multiply = (
	a,
	b
):
	a * b

println(multiply(2, 3))
"""
	) == "6\n"

	assert output_from_code(
		"""\
multiple_statements = ():
	a = 1
	b = 2

	println(a.__to_str__() + " + " + b.__to_str__() + " = " + (a + b).__to_str__())

multiple_statements()
"""
	) == "1 + 2 = 3\n"

	assert output_from_code(
		"""\
plus = (x): (y): x + y

println(plus(1)(2))
"""
	) == "3\n"

def test_invalid_functions():
	assert output_from_code(': "No parameter list!"', expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token ":"

  1  │ : "No parameter list!"
     │ ^

"""

	assert output_from_code('(): message = "This is a statement."', expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "="

  1  │ (): message = "This is a statement."
     │             ^

"""
