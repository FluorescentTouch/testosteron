package sync

import "sync"

type Map[T any] struct {
	m map[string]T
	l *sync.Mutex
}

func MakeSyncMap[T any]() Map[T] {
	return Map[T]{
		m: make(map[string]T),
		l: &sync.Mutex{},
	}
}

func (s Map[T]) Get(key string) (T, bool) {
	s.l.Lock()
	defer s.l.Unlock()

	v, ok := s.m[key]
	return v, ok
}

func (s Map[T]) Set(key string, val T) {
	s.l.Lock()
	defer s.l.Unlock()

	s.m[key] = val
}

func (s Map[T]) Delete(key string) {
	s.l.Lock()
	defer s.l.Unlock()

	delete(s.m, key)
}
