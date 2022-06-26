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
