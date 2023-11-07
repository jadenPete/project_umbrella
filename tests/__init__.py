import os
import subprocess
import tempfile

def output_from_code(code: str, expected_return_code=0) -> str:
	with tempfile.NamedTemporaryFile(mode="w+") as file:
		file.write(code)
		file.flush()

		process = subprocess.run(
			[
				os.path.join("src", "interpreter", "interpreter_", "interpreter"),
				os.path.join("tests", "data", file.name)
			],

			stdout=subprocess.PIPE,
			stderr=subprocess.STDOUT,
			text=True
		)

		assert process.returncode == expected_return_code

		return process.stdout
