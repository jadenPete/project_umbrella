from tests import output_from_code

def test_basic() -> None:
	assert output_from_code(
		"""\
fn Color(red, green, blue):
	fn field_factory(_):
		(
			("red", red),
			("green", green),
			("blue", blue)
		)

	__struct__("Color", Color, field_factory, (,))

println(Color(0, 0, 255)("blue"))
"""
	) == "255\n"

	assert output_from_code(
		"""\
fn Color(red, green, blue):
	fn field_factory(_):
		(,)

	__struct__(
		"Color",
		Color,
		field_factory,

		(
			("red", red),
			("green", green),
			("blue", blue)
		)
	)

println(Color(0, 0, 255)("blue"))
"""
	) == "255\n"

	assert output_from_code(
		"""\
fn Color(red, green, blue):
	fn brightness():
		(red + green + blue) / 3

	fn field_factory(_):
		(
			("red", red),
			("green", green),
			("blue", blue),
			("brightness", brightness)
		)

	__struct__("Color", Color, field_factory, (,))

println(Color(0, 0, 255)("brightness")())
"""
	) == "85\n"

def test_nonexistent_fields() -> None:
	assert output_from_code(
		"""\
fn Struct():
	fn field_factory(_):
		(,)

	__struct__("Struct", Struct, field_factory, (,))

Struct()("")
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: ``\n"

	assert output_from_code(
		"""\
fn Struct():
	fn field_factory(_):
		(,)

	__struct__("Struct", Struct, field_factory, (,))

Struct()("foo")
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `foo`\n"

def test_self() -> None:
	assert output_from_code(
		"""\
fn Struct():
	fn field_factory(_):
		(,)

	__struct__("Struct", Struct, field_factory, (,))

Struct()("self")
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `self`\n"

	assert output_from_code(
		"""\
fn Struct():
	fn field_factory(self):
		(("self_alias", self),)

	__struct__("Struct", Struct, field_factory, (,))

struct_ = Struct()

println(struct_("self_alias") == struct_)
"""
	) == "true\n"

	assert output_from_code(
		"""\
fn Struct():
	fn field_factory(self):
		println(self("foo"))

		(("foo", "bar"),)

	__struct__("Struct", Struct, field_factory, (,))

Struct()
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `foo`\n"

def test_call_factory_once() -> None:
	assert output_from_code(
		"""\
fn Struct():
	fn field_factory(_):
		println("Called!")

		(("foo", "bar"),)

	__struct__("Struct", Struct, field_factory, (,))

struct_ = Struct()
struct_("foo")
struct_("foo")
"""
	) == "Called!\n"

def _malformed_argument_error(i: int) -> str:
	return f"""\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

__struct__ expected argument #{i + 1} to be of a different type.
"""

def test_strong_typing_argument_0() -> None:
	assert output_from_code(
		"""\
fn Struct():
	__struct__(0, Struct, (,), (,))

Struct()
""",
		expected_return_code=1
	) == _malformed_argument_error(0)

def test_strong_typing_argument_1() -> None:
	assert output_from_code(
		"""\
fn Struct():
	__struct__("Struct", "Struct", (,), (,))

Struct()
""",
		expected_return_code=1
	) == _malformed_argument_error(1)

def test_strong_typing_argument_2() -> None:
	assert output_from_code(
		"""\
fn Struct():
	fn field_factory():

	__struct__("Struct", Struct, field_factory, (,))

Struct()
""",
		expected_return_code=1
	) == "Error (RUNTIME-1): A function accepting 0 arguments was called with 1 arguments\n"

	code1 = """\
fn Struct():
	__struct__("Struct", Struct, (("foo", "bar"),), (,))

Struct()
"""

	code2 = """\
fn Struct():
	fn field_factory(_):

	__struct__("Struct", Struct, field_factory, (,))

Struct()
"""

	code3 = """\
fn Struct():
	fn field_factory(_):
		("foo", "bar")

	__struct__("Struct", Struct, field_factory, (,))

Struct()
"""

	code4 = """\
fn Struct():
	fn field_factory(_):
		((,),)

	__struct__("Struct", Struct, field_factory, (,))

Struct()
"""

	code5 = """\
fn Struct():
	fn field_factory(_):
		(("foo",),)

	__struct__("Struct", Struct, field_factory, (,))

Struct()
"""

	code6 = """\
fn Struct():
	fn field_factory(_):
		((0, "bar"),)

	__struct__("Struct", Struct, field_factory, (,))

Struct()
"""

	assert \
		output_from_code(code1, expected_return_code=1) == \
		output_from_code(code2, expected_return_code=1) == \
		output_from_code(code3, expected_return_code=1) == \
		output_from_code(code4, expected_return_code=1) == \
		output_from_code(code5, expected_return_code=1) == \
		output_from_code(code6, expected_return_code=1) == _malformed_argument_error(2)

def test_strong_typing_argument_3() -> None:
	code1 = """\
fn Struct():
	fn field_factory(_):
		(,)

	__struct__("Struct", Struct, field_factory, "foo")

Struct()
"""

	code2 = """\
fn Struct():
	fn field_factory(_):
		(,)

	__struct__("Struct", Struct, field_factory, ("foo", "bar"))

Struct()
"""

	code3 = """\
fn Struct():
	fn field_factory(_):
		(,)

	__struct__("Struct", Struct, field_factory, (("foo",),))

Struct()
"""

	assert \
		output_from_code(code1, expected_return_code=1) == \
		output_from_code(code2, expected_return_code=1) == \
		output_from_code(code3, expected_return_code=1) == _malformed_argument_error(3)

def test_strong_typing_returned_function() -> None:
	assert output_from_code(
		"""\
fn Struct():
	fn field_factory(_):
		(("foo", "bar"),)

	__struct__("Struct", Struct, field_factory, (,))

Struct()(0)
""",
		expected_return_code=1
	) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

(built-in function) expected argument #1 to be of a different type.
"""
