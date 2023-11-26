package common

import "os"

func Abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
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

func Max(x int, y int) int {
	if x < y {
		return y
	}

	return x
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
