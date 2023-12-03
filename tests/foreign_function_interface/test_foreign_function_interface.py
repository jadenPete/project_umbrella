import os
from tests import output_from_code

TEST_LIBRARY_DIRECTORY = os.path.join("tests", "foreign_function_interface", "test_libraries")

def _output_from_code_loading_library(code: str, expected_return_code=0) -> str:
	return output_from_code(
		code,
		expected_return_code=expected_return_code,
		krait_path_directories=
			[TEST_LIBRARY_DIRECTORY, os.path.join(TEST_LIBRARY_DIRECTORY, "test_library_valid_")]
	)

def test_loading_value() -> None:
	assert _output_from_code_loading_library(
		'println(import_library("test_library_valid").get("MeaningOfLife"))\n'
	) == "42\n"

def test_loading_function() -> None:
	assert _output_from_code_loading_library(
		"""\
square = import_library("test_library_valid").get("Square")

println(square(5))
"""
	) == "25\n"

def test_loading_nonexistent_library() -> None:
	assert output_from_code(
		'import_library("test_library_nonexistent")\n',
		expected_return_code=1
	) == "Error (RUNTIME-15): The library \"test_library_nonexistent\" wasn't found\n"

def test_loading_invalid_library() -> None:
	_output_from_code_loading_library(
		'import_library("test_library_invalid")\n',
		expected_return_code=2
	)

def test_loading_nonexistent_symbol() -> None:
	assert _output_from_code_loading_library(
		'println(import_library("test_library_valid").get("NonexistentSymbol"))\n',
		expected_return_code=1
	) == """\
Error (RUNTIME-17): Couldn't fetch the symbol "NonexistentSymbol" from the library at "tests/foreign_function_interface/test_libraries/test_library_valid_/test_library_valid.so"

"NonexistentSymbol" doesn't exist.
"""

def test_loading_invalid_symbol() -> None:
	assert _output_from_code_loading_library(
		'println(import_library("test_library_valid").get("InvalidSymbol"))\n',
		expected_return_code=1
	) == """\
Error (RUNTIME-16): Couldn't fetch the symbol "InvalidSymbol" from the library at "tests/foreign_function_interface/test_libraries/test_library_valid_/test_library_valid.so"

"InvalidSymbol" isn't a value.
"""

def test_import_library_caching() -> None:
	assert _output_from_code_loading_library(
		"""\
fn random_integer(): import_library("test_library_valid").get("RandomInteger")

println(random_integer() == random_integer())
"""
	) == "true\n"

def test_import_library_strong_typing() -> None:
	assert output_from_code("import_library(0)\n", expected_return_code=1) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

import_library expected argument #1 to be of a different type.
"""

def test_library_to_str() -> None:
	assert _output_from_code_loading_library('println(import_library("test_library_valid"))\n') == \
		"(library)\n"

def test_library_strong_typing() -> None:
	assert _output_from_code_loading_library(
		'import_library("test_library_valid").get(0)\n',
		expected_return_code=1
	) == """\
Error (RUNTIME-2): A built-in function was called with an argument of incorrect type

get expected argument #1 to be of a different type.
"""
