from tests import output_from_code

def _test_character_manipulation(character: str, expected_codepoint: int) -> None:
	assert output_from_code(
		f'''\
println(
	("{character}".codepoint() == {expected_codepoint}) &&
	(({expected_codepoint}).to_character() == "{character}")
)
'''
	) == "true\n"

def test_codepoint() -> None:
	_test_character_manipulation("A", 65)
	_test_character_manipulation("€", 8364)
	_test_character_manipulation("😀", 128512)

	assert output_from_code('println("".codepoint())\n', expected_return_code=1) == """\
Error (RUNTIME-18): `codepoint` was called on a non-character

`codepoint` was called on a string of length 0: ""
"""

	assert output_from_code('println("foo".codepoint())\n', expected_return_code=1) == """\
Error (RUNTIME-18): `codepoint` was called on a non-character

`codepoint` was called on a string of length 3: "foo"
"""

def test_get() -> None:
	assert output_from_code('println("abc".get(0))\n') == "a\n"
	assert output_from_code('println("abc".get(1))\n') == "b\n"
	assert output_from_code('println("abc".get(2))\n') == "c\n"
	assert output_from_code('println("abc".get(3))\n', expected_return_code=1) == """\
Error (RUNTIME-14): An out-of-bounds index was provided to string#get

Expected an index in the range [0, 3), but got 3.
"""

def test_length() -> None:
	assert output_from_code('println("".length)\n') == "0\n"
	assert output_from_code('println("0".length)\n') == "1\n"
	assert output_from_code('println("01".length)\n') == "2\n"

def test_plus() -> None:
	assert output_from_code(
		'println((("foo" + "") == "foo") && (("" + "foo") == "foo"))'
	) == "true\n"

	assert output_from_code(
		'println((("foo" + "bar") == "foobar") && (("bar" + "foo") == "barfoo"))'
	) == "true\n"

	assert output_from_code('"foo" + 0', expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

+ expected argument #1 to be of a different type.
"""

def test_slice() -> None:
	assert output_from_code('println("Hello".slice(0, 5))\n') == "Hello\n"
	assert output_from_code('println("Hello".slice(1, 4))\n') == "ell\n"
	assert output_from_code('println("Hello".slice(1, 1))\n') == "\n"
	assert output_from_code('println("Hello".slice(1, 0))\n') == "\n"
	assert output_from_code('println("Hello".slice(-1, 4))\n') == "Hell\n"
	assert output_from_code('println("Hello".slice(1, 6))\n') == 'ello\n'
	assert output_from_code('println("Hello".slice(0))\n', expected_return_code=1) == \
		"Error (RUNTIME-1): A function accepting 2 arguments was called with 1 arguments\n"

	assert output_from_code('println("Hello".slice("0", "3"))\n', expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

slice expected argument #1 to be of a different type.
"""

def test_split() -> None:
	assert output_from_code('println("".split("") == (,))\n') == "true\n"
	assert output_from_code('println("".split("foo") == ("",))\n') == "true\n"
	assert output_from_code('println("foo".split("") == ("f", "o", "o"))\n') == "true\n"
	assert output_from_code('println("foo".split("foo") == ("", ""))\n') == "true\n"
	assert output_from_code('println("foobar".split("bar") == ("foo", ""))\n') == "true\n"
	assert output_from_code('println("foobarbizz".split("bar") == ("foo", "bizz"))\n') == "true\n"
	assert output_from_code('"".split(0)\n', expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

split expected argument #1 to be of a different type.
"""

def test_strip() -> None:
	assert output_from_code('println(" Hello, world! ".strip(" "))\n') == "Hello, world!\n"
	assert output_from_code('println("Hello, world! 😀".strip("😀"))\n') == "Hello, world! \n"
	assert output_from_code('println("Hello, world!".strip(" "))\n') == "Hello, world!\n"
	assert output_from_code('println(" ".strip(" "))\n') == "\n"
	assert output_from_code('println("Hello, world!".strip(""))\n') == "Hello, world!\n"
	assert output_from_code('println("Hello, world!".strip("H!"))\n') == "ello, world\n"
