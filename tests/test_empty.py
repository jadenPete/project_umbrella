from tests import output_from_code

def test_empty() -> None:
	assert output_from_code("") == ""
	assert output_from_code("\n") == ""
