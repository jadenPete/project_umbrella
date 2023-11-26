import os
import re
from tests import output_from_code, output_from_multiple_files

def test_imports() -> None:
	assert output_from_multiple_files(
		{
			"main.krait": """\
message = import("message")

println(message.message)
""",

			"message.krait": 'message = "Hello, world!"\n'
		},

		"main.krait"
	) == "Hello, world!\n"

	assert output_from_multiple_files(
		{
			"main.krait": """\
bar = import("foo.bar")

println(bar.foo)
""",

			os.path.join("foo", "bar.krait"): 'foo = "bar"\n'
		},

		"main.krait"
	) == "bar\n"

	assert output_from_multiple_files(
		{
			"main.krait": """\
bar = import("foo.bar")

println(bar.foo)
""",

			os.path.join("foo", "bar.krait"): 'foo = "bar"\n',
			os.path.join("foo.krait"): 'foo = "foo"'
		},

		"main.krait"
	) == "bar\n"

def test_importing_nonexistent_modules() -> None:
	assert output_from_code('import("foo")\n', expected_return_code=1) == \
		"Error (RUNTIME-13): The module \"foo\" wasn't found\n"

	assert output_from_multiple_files(
		{
			"main.krait": 'import("foo.foo")\n',
			os.path.join("foo", "bar.krait"): "",
		},

		"main.krait",
		expected_return_code=1
	) == "Error (RUNTIME-13): The module \"foo.foo\" wasn't found\n"

	assert output_from_multiple_files(
		{
			"main.krait": 'import("foo")\n',
			os.path.join("foo", "bar.krait"): "",
		},

		"main.krait",
		expected_return_code=1
	) == "Error (RUNTIME-13): The module \"foo\" wasn't found\n"

	assert output_from_multiple_files(
		{
			"main.krait": 'import("foo.bar")\n',
			"foo.krait": "",
		},

		"main.krait",
		expected_return_code=1
	) == "Error (RUNTIME-13): The module \"foo.bar\" wasn't found\n"

def test_import_cycles() -> None:
	assert re.match(
		 r"""Error \(RUNTIME-13\): Encountered an import cycle

".*/main\.krait" couldn't be imported\. See the following import stack\.

.*/main\.krait$""",

		output_from_multiple_files(
			{
				"main.krait": 'import("main")\n',
			},

			"main.krait",
			expected_return_code=1
		),

		re.MULTILINE
	) is not None

	assert re.match(
		r"""Error \(RUNTIME-13\): Encountered an import cycle

".*/bar\.krait" couldn't be imported\. See the following import stack\.

.*/main\.krait
↳ .*/foo\.krait
↳ .*/bar\.krait$""",

		output_from_multiple_files(
			{
				"main.krait": 'import("foo")\n',
				"foo.krait": 'import("bar")\n',
				"bar.krait": 'import("main")\n'
			},

			"main.krait",
			expected_return_code=1
		)
	)

def test_exported_values() -> None:
	assert output_from_multiple_files(
		{
			"main.krait": """\
greeting_printer = import("greeting_printer")
greeting_printer.print_greeting(greeting_printer.user)
""",

			"greeting_printer.krait": """\
user = "user"

fn print_greeting(subject):
	println("Hello, " + subject + "!")
"""
		},

		"main.krait"
	) == "Hello, user!\n"
