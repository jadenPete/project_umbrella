from tests import output_from_code

def test_basic() -> None:
	assert output_from_code(
		"""\
struct Box(self, value):

println(Box("foo").value)
"""
	) == "foo\n"

	assert output_from_code(
		"""\
struct Box(self, value):

println(Box("foo")("value"))
"""
	) == "foo\n"

	assert output_from_code(
		"""\
struct Incrementor(self, value):
	fn incremented():
		value + 1

println(Incrementor(1).incremented())
"""
	) == "2\n"

def test_self_parameter() -> None:
	assert output_from_code(
		"""\
struct Box(self, value):
	fn value_mirror():
		self.value

println(Box("foo").value_mirror())
"""
	) == "foo\n"

	assert output_from_code(
		"""\
struct Box(self, value):
	self.value

Box("foo")
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `value`\n"

	assert output_from_code(
		"""\
struct Struct(self):
	println(self)

Struct()
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `__to_str__`\n"

	assert output_from_code("struct Struct():\n", expected_return_code=1) == """\
Error (PARSER-1): The parser failed: unexpected token "struct"

  1  │ struct Struct():
     │ ^^^^^^

"""

def test_cannot_override_methods() -> None:
	assert output_from_code(
		"""\
struct Struct(self):
	fn __to_str__():
		""

println(Struct())
"""
	) == "Struct()\n"

	assert output_from_code(
		"""\
struct Struct(self):
	fn ==(_):
		false

println(Struct() == Struct())
"""
	) == "true\n"

	assert output_from_code(
		"""\
struct Struct(self):
	fn !=(_):
		true

println(Struct() != Struct())
"""
	) == "false\n"

def test_formatting() -> None:
	assert output_from_code(
		"""\
struct
	Struct(self):
		value = "foo"

println(Struct().value)
"""
	) == "foo\n"

	assert output_from_code(
		"""\
struct Struct
	(self):
		value = "foo"

println(Struct().value)
"""
	) == "foo\n"

	assert output_from_code(
		"""\
struct Struct(
	self
):
	value = "foo"

println(Struct().value)
"""
	) == "foo\n"

def _test_equals_case(definitions: str, instantiation1: str, instantiation2: str) -> None:
	print(
		f"""\
{definitions}

println(
	({instantiation1} == {instantiation1}) &&
	!({instantiation1} != {instantiation1}) &&
	!({instantiation1} == {instantiation2}) &&
	({instantiation1} != {instantiation2})
)
"""
	)
	assert output_from_code(
		f"""\
{definitions}

println(
	({instantiation1} == {instantiation1}) &&
	!({instantiation1} != {instantiation1}) &&
	!({instantiation1} == {instantiation2}) &&
	({instantiation1} != {instantiation2})
)
"""
	) == "true\n"

def test_equals() -> None:
	_test_equals_case(
		"""\
struct Struct1(self):
struct Struct2(self):""",
		"Struct1()",
		"Struct2()"
	)

	_test_equals_case("struct Box(self, value):", 'Box("foo")', 'Box("bar")')
	_test_equals_case("struct Box(self, value):", 'Box(Box("foo"))', 'Box(Box("bar"))')
	_test_equals_case(
		"""\
struct Box1(self, value):
struct Box2(self, value):""",
		'Box1("foo")',
		'Box2("foo")'
	)

def test_to_string() -> None:
	assert output_from_code(
		"""\
struct Struct(self):

println(Struct())
"""
	) == "Struct()\n"

	assert output_from_code(
		"""\
struct Box(self, value):

println(Box("foo"))
"""
	) == "Box(foo)\n"

	assert output_from_code(
		"""\
struct Pair(self, number1, number2):

println(Pair(1, 2))
"""
	) == "Pair(1, 2)\n"

	assert output_from_code(
		"""\
struct Box(self, value):

println(Box(Box("foo")))
"""
	) == "Box(Box(foo))\n"
