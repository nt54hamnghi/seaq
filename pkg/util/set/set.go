package set

import "iter"

type Set[T comparable] map[T]struct{}

func New[T comparable](elements ...T) Set[T] {
	s := make(Set[T])
	for _, e := range elements {
		s.Add(e)
	}
	return s
}

func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for k := range s {
			if !yield(k) {
				return
			}
		}
	}
}
