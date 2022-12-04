package slices

func Map[E, D any](s []E, f func(e E) D) []D {
	var x []D

	for _, ss := range s {
		x = append(x, f(ss))
	}

	return x
}

func ToMap[E any, K comparable](s []E, f func(e E) K) map[K]E {
	var m = make(map[K]E)

	for _, ss := range s {
		m[f(ss)] = ss
	}

	return m
}
