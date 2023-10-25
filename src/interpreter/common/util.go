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

func LinkedListToSlice[LinkedList any, Element any](
	linkedList *LinkedList,
	head func(*LinkedList) Element,
	tail func(*LinkedList) *LinkedList,
) ([]Element, *LinkedList) {
	result := []Element{}

	if linkedList == nil {
		return result, nil
	}

	current := linkedList

	for {
		result = append(result, head(current))
		next := tail(current)

		if next == nil {
			break
		}

		current = next
	}

	return result, current
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
