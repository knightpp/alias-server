package fp

func Map[T any, U any](slice []T, f func(t T) U) []U {
	if len(slice) == 0 {
		return nil
	}

	mapped := make([]U, len(slice))
	for i, el := range slice {
		mapped[i] = f(el)
	}

	return mapped
}

func Keys[K comparable, V any](m map[K]V) []K {
	if len(m) == 0 {
		return nil
	}

	slice := make([]K, len(m))
	i := 0
	for k := range m {
		slice[i] = k
		i++
	}
	return slice
}

func Values[K comparable, V any](m map[K]V) []V {
	if len(m) == 0 {
		return nil
	}

	slice := make([]V, len(m))
	i := 0
	for _, v := range m {
		slice[i] = v
		i++
	}
	return slice
}

func FilterInPlace[T any](slice []T, pred func(T) bool) []T {
	newSlice := slice[:0]

	for _, el := range slice {
		if pred(el) {
			newSlice = append(newSlice, el)
		}
	}

	var empty T
	for i := len(newSlice); i < len(slice); i++ {
		slice[i] = empty
	}

	return newSlice
}
