package tuple

type T2[A any, B any] struct {
	A A
	B B
}

func NewT2[A, B any](a A, b B) T2[A, B] {
	return T2[A, B]{
		A: a, B: b,
	}
}
