package common

import (
	"os"
	"path/filepath"
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func IsDirectoryAncestorOfFile(directory string, filePath string) bool {
	absolutizeAndEvaluate := func(path string) (string, error) {
		absolute, err := filepath.Abs(path)

		if err != nil {
			return "", err
		}

		evaluated, err := filepath.EvalSymlinks(absolute)

		if err != nil {
			return "", err
		}

		return evaluated, nil
	}

	evaluatedDirectory, err := absolutizeAndEvaluate(directory)

	if err != nil {
		return false
	}

	currentFilePath, err := absolutizeAndEvaluate(filePath)

	if err != nil {
		return false
	}

	for {
		newFilePath := filepath.Dir(currentFilePath)

		if newFilePath == currentFilePath {
			break
		}

		currentFilePath = newFilePath

		if currentFilePath == evaluatedDirectory {
			return true
		}
	}

	return currentFilePath == evaluatedDirectory
}

func IsDirectoryUnsafe(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		panic(err)
	}

	return info.IsDir()
}

func IsFileUnsafe(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		panic(err)
	}

	return info.Mode().IsRegular()
}

func Repeat[T any](element T, size int) []T {
	result := make([]T, 0, size)

	for i := 0; i < size; i++ {
		result = append(result, element)
	}

	return result
}

/*
 * Resize a slice to a given size.
 *
 * If `size > len(slice)`, the zero value of `T` is repeatedly appended to `slice`.
 */
func Resize[T any](slice []T, size int) []T {
	if size < len(slice) {
		return slice[:size]
	}

	return append(slice, make([]T, size-len(slice))...)
}
