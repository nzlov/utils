package syncmap

import "sync"

type Map[K any, T any] struct {
	m *sync.Map
}

func New[K any, T any]() *Map[K, T] {
	return &Map[K, T]{
		m: new(sync.Map),
	}
}

func (m *Map[K, T]) Load(key K) (value T, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return *new(T), ok
	}
	return v.(T), ok
}

func (m *Map[K, T]) Store(key K, value T) {
	m.m.Store(key, value)
}

func (m *Map[K, T]) Clear() {
	m.m.Clear()
}

func (m *Map[K, T]) LoadOrStore(key K, value T) (actual T, loaded bool) {
	v, ok := m.m.LoadOrStore(key, value)
	if !ok {
		return *new(T), ok
	}
	return v.(T), ok
}

func (m *Map[K, T]) LoadAndDelete(key K) (value T, loaded bool) {
	v, ok := m.m.LoadAndDelete(key)
	if !ok {
		return *new(T), ok
	}
	return v.(T), ok
}

func (m *Map[K, T]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, T]) Swap(key K, value T) (previous T, loaded bool) {
	v, ok := m.m.Swap(key, value)
	if !ok {
		return *new(T), ok
	}
	return v.(T), ok
}

func (m *Map[K, T]) CompareAndSwap(key K, old, new T) (swapped bool) {
	return m.m.CompareAndSwap(key, old, new)
}

func (m *Map[K, T]) CompareAndDelete(key K, old T) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

func (m *Map[K, T]) Range(f func(key K, value T) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(T))
	})
}
