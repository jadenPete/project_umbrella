from tests import output_from_code

def test_empty() -> None:
	assert output_from_code("println(__tuple__())\n") == "(,)\n"

def test_one_element() -> None:
	assert output_from_code("println(__tuple__(0))\n") == "(0,)\n"

def test_multiple_elements() -> None:
	assert output_from_code("println(__tuple__(0, 1))\n") == "(0, 1)\n"
