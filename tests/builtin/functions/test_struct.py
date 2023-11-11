from tests import output_from_code

def test_basic():
	assert output_from_code(
		"""\
fn field_factory(_):
	(,)

println(__struct__(field_factory))
"""
	) == "(built-in function)\n"

	assert output_from_code(
		"""\
fn Color(red, green, blue):
	fn field_factory(_):
		(
			("red", red),
			("green", green),
			("blue", blue)
		)

	__struct__(field_factory)

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

	__struct__(field_factory)

println(Color(0, 0, 255)("brightness")())
"""
	) == "85\n"

def test_nonexistent_fields():
	assert output_from_code(
		"""\
fn field_factory(_):
	(,)

__struct__(field_factory)("")
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: ``\n"

	assert output_from_code(
		"""\
fn field_factory(_):
	(,)

__struct__(field_factory)("foo")
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `foo`\n"

def test_self():
	assert output_from_code(
		"""\
fn field_factory(_):
	(,)

__struct__(field_factory)("self")
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `self`\n"

	assert output_from_code(
		"""\
fn field_factory(self):
	(("self_alias", self),)

struct_ = __struct__(field_factory)

println(struct_("self_alias") == struct_)
"""
	) == "true\n"

	assert output_from_code(
		"""\
fn field_factory(self):
	println(self("foo"))

	(("foo", "bar"),)

__struct__(field_factory)
""",
		expected_return_code=1
	) == "Error (RUNTIME-9): Unknown field: `foo`\n"

def test_call_factory_once():
	assert output_from_code(
		"""\
fn field_factory(_):
	println("Called!")

	(("foo", "bar"),)

struct_ = __struct__(field_factory)
struct_("foo")
struct_("foo")
"""
	) == "Called!\n"

def test_strong_typing():
	maltyped_argument_error = """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

__struct__ expected argument #1 to be of a different type.
"""

	assert output_from_code('__struct__((("foo", "bar"),))\n', expected_return_code=1) == \
		maltyped_argument_error

	assert output_from_code(
		"""\
fn field_factory():

__struct__(field_factory)
""",
		expected_return_code=1
	) == "Error (RUNTIME-1): A function accepting 0 arguments was called with 1 arguments\n"

	code1 = """\
fn field_factory(_):

__struct__(field_factory)
"""

	code2 = """\
fn field_factory(_):
	("foo", "bar")

__struct__(field_factory)
"""

	code3 = """\
fn field_factory(_):
	((,),)

__struct__(field_factory)
"""

	code4 = """\
fn field_factory(_):
	(("foo",))

__struct__(field_factory)
"""

	code4 = """\
fn field_factory(_):
	((0, "bar"),)

__struct__(field_factory)
"""

	assert \
		output_from_code(code1, expected_return_code=1) == \
		output_from_code(code2, expected_return_code=1) == \
		output_from_code(code3, expected_return_code=1) == \
		output_from_code(code4, expected_return_code=1) == maltyped_argument_error

	assert output_from_code(
		"""\
fn field_factory(_):
	(("foo", "bar"),)

__struct__(field_factory)(0)
""",
		expected_return_code=1
	) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

(built-in function) expected argument #1 to be of a different type.
"""
