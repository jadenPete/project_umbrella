package environment_variables

import (
	"os"
	"runtime"
)

func getEnvironmentVariable(name string, defaultLinuxValue string) string {
	if result, ok := os.LookupEnv(name); ok {
		return result
	}

	if runtime.GOOS == "linux" {
		return defaultLinuxValue
	}

	return ""
}

var (
	KRAIT_PATH            = getEnvironmentVariable("KRAIT_PATH", "/usr/lib/standard_library")
	KRAIT_STARTUP         = getEnvironmentVariable("KRAIT_STARTUP", "/usr/lib/startup_file.krait")
	KRAIT_STARTUP_EXCLUDE = getEnvironmentVariable(
		"KRAIT_STARTUP_EXCLUDE",
		"/usr/lib/standard_library",
	)
)
