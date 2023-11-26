import contextlib
import os
import subprocess
import tempfile

def output_from_code(code: str, expected_return_code=0) -> str:
	return output_from_multiple_files(
		{
			"main.krait": code
		},

		"main.krait",
		expected_return_code=expected_return_code
	)

def output_from_multiple_files(
	files: dict[str, str],
	entry_point: str,
	expected_return_code=0
) -> str:
	with tempfile.TemporaryDirectory() as directory, contextlib.ExitStack() as exit_stack:
		for path, code in files.items():
			full_path = os.path.join(directory, path)

			os.makedirs(os.path.dirname(full_path), exist_ok=True)

			file = exit_stack.enter_context(open(full_path, mode="w+"))
			file.write(code)
			file.flush()

		process = subprocess.run(
			[
				os.path.join("src", "interpreter", "interpreter_", "interpreter"),
				os.path.join(directory, entry_point)
			],

			stdout=subprocess.PIPE,
			stderr=subprocess.STDOUT,
			env={
				**os.environ,
				"KRAIT_PATH": directory
			},

			text=True
		)

		if process.returncode != expected_return_code:
			print(process.stdout, end="")

			raise AssertionError(
				f"Expected a return code of {expected_return_code}; got {process.returncode}"
			)

		return process.stdout
