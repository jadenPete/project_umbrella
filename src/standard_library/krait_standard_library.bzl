"""
Defines a rule generating the Krait standard library from its components written in Krait
(located in `src/standard_library/krait`) and its native library components
(located in `src/standard_library/native`).
"""

def _krait_standard_library_impl(ctx):
    output_directory = ctx.actions.declare_directory("standard_library")
    arguments = []

    for target, directory in ctx.attr.deps.items():
        for file in target.files.to_list():
            arguments.append(directory)
            arguments.append(file.path)

    ctx.actions.run(
        outputs = [output_directory],
        inputs = [file for target in ctx.attr.deps for file in target.files.to_list()],
        executable = ctx.file._compiler,
        arguments = [output_directory.path] + arguments,
    )

    return DefaultInfo(files = depset([output_directory]))

krait_standard_library = rule(
    implementation = _krait_standard_library_impl,
    attrs = {
        "deps": attr.label_keyed_string_dict(),
        "_compiler": attr.label(
            default = "//src/standard_library:compiler.sh",
            executable = True,
            allow_single_file = True,
            cfg = "exec",
        )
    },
)
