package parser

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

/*
 * Convert an expression to bytecode, but use a previously compiled bytecode file ("*.krc" file)
 * from the cache if possible.
 *
 * Compiled files are stored in `$XDG_CACHE_HOME/projectumbrella` or `$HOME/.cache/projectumbrella`,
 * whichever is resolvable, and are named according to their SHA256 hash.
 *
 * They're stored as a literal MessagePack pickling of the resulting `Bytecode` object.
 */
func ExpressionToBytecodeFromCache(expression Expression, fileContent string) *Bytecode {
	var appDirectory string

	if cacheDirectory, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		appDirectory = fmt.Sprintf("%s/projectumbrella", cacheDirectory)
	} else if homeDirectory, ok := os.LookupEnv("HOME"); ok {
		appDirectory = fmt.Sprintf("%s/.cache/projectumbrella", homeDirectory)
	} else {
		panic("Parser error: The HOME environment variable is undefined.")
	}

	if os.MkdirAll(appDirectory, 0755) != nil {
		panic(fmt.Sprintf("Couldn't create the directory %s", appDirectory))
	}

	checksum := sourceChecksum(fileContent)

	bytecodePath := fmt.Sprintf("%s/%s.krc", appDirectory, hex.EncodeToString(checksum[:16]))

	file, err := os.Open(bytecodePath)

	if err == nil {
		defer func() {
			if err := file.Close(); err != nil {
				panic(err)
			}
		}()

		if encoded, err := io.ReadAll(file); err == nil {
			if bytecode := DecodeBytecode(encoded); bytecode.sourceChecksum == checksum {
				return bytecode
			}
		}
	}

	bytecode := NewBytecodeTranslator().ExpressionToBytecode(expression, fileContent)

	os.WriteFile(bytecodePath, bytecode.Encode(), 0644)

	return bytecode
}
