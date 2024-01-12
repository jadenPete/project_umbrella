import contextlib
import os
import subprocess
import tempfile

REPOSITORY_DIRECTORY = os.environ["BUILD_WORKING_DIRECTORY"]
STANDARD_LIBRARY_DIRECTORY = os.path.join("src", "standard_library", "standard_library")
STARTUP_FILE_PATH = os.path.join(REPOSITORY_DIRECTORY, "src", "startup_file.krait")

def output_from_code(
	code: str,
	expected_return_code=0,
	krait_path_directories: list[str] = []
) -> str:
	return output_from_multiple_files(
		{
			"main.krait": code
		},

		"main.krait",
		expected_return_code=expected_return_code,
		krait_path_directories=krait_path_directories
	)

def output_from_multiple_files(
	files: dict[str, str],
	entry_point: str,
	expected_return_code=0,
	krait_path_directories: list[str] = []
) -> str:
	krait_path_prefix = "".join(f"{directory}:" for directory in krait_path_directories)

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
				"KRAIT_PATH": f"{krait_path_prefix}{directory}:{STANDARD_LIBRARY_DIRECTORY}",
				"KRAIT_STARTUP": STARTUP_FILE_PATH,
				"KRAIT_STARTUP_EXCLUDE": STANDARD_LIBRARY_DIRECTORY
			},

			text=True
		)

		if process.returncode != expected_return_code:
			print(process.stdout, end="")

			raise AssertionError(
				f"Expected a return code of {expected_return_code}; got {process.returncode}"
			)

		return process.stdout
