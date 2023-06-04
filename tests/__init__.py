import os
import subprocess
import tempfile

def output_from_code(code: str) -> str:
	with tempfile.NamedTemporaryFile(mode="w+") as file:
		file.write(code)
		file.flush()

		return subprocess.check_output(
			[os.path.join("src", "interpreter", "interpreter_", "interpreter"), file.name],
			text=True
		)

def output_from_filename(filename: str) -> str:
	with open(os.path.join("tests", "data", filename)) as file:
		return output_from_code(file.read())
