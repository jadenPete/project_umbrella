from tests import output_from_code

def test_values() -> None:
	assert output_from_code(
		"""\
message = "Hello, world!"

println(message)
"""
	) == "Hello, world!\n"

def test_formatting() -> None:
	assert output_from_code(
		"""\
message1 =
	"Hello, world!"

message2
	= "Hello, world!"

message3
	=
	"Hello, world!"

println((message1 == "Hello, world!") && (message1 == message2) && (message1 == message3))
"""
	) == "true\n"

def test_invalid_values() -> None:
	assert output_from_code("value =\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "="

  1  │ value =
     │       ^

"""

	assert output_from_code("= value\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "="

  1  │ = value
     │ ^

"""

def test_value_cycles() -> None:
	assert output_from_code("value = value\n", expected_return_code=1) == """\
Error (PARSER-6): Unknown value: `value`

  1  │ value = value
     │         ^^^^^

"""

def test_value_aliases() -> None:
	assert output_from_code(
		"""\
message1 = message2 = "Hello, world!"

println(message1 == message2)
"""
	) == "true\n"

	assert output_from_code(
		"""\
message = message = "Hello, world!"

println(message)
"""
	) == "Hello, world!\n"

def test_value_alias_formatting() -> None:
	assert output_from_code(
		"""\
message1 =
	message2
		= message3
			=
				message4 = "Hello, world!"

println((message1 == message2) && (message1 == message3) && (message1 == message4))
"""
	) == "true\n"

def test_value_shadowing() -> None:
	assert output_from_code(
		"""\
message = "Hello, world!"
message = "Hey, world!"

println(message)
""",
	expected_return_code=1
	) == """\
Error (PARSER-5): Reassigning to an already declared value is impossible

  1  │ message = "Hello, world!"
  2  │ message = "Hey, world!"
     │ ^^^^^^^^^^^^^^^^^^^^^^^

Consider assigning to a new value.
"""

	assert output_from_code(
		"""\
println = print

println("Hello, world!")
"""
	) == "Hello, world!"

def test_values_containing_keywords() -> None:
	assert output_from_code("if_ = unit") == ""
	assert output_from_code("else_ = unit") == ""
	assert output_from_code("fn_ = unit") == ""
