package common

func Abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func Max(x int, y int) int {
	if x < y {
		return y
	}

	return x
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
