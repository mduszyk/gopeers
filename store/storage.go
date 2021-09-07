package store

import (
	"errors"
	"sync"
)

type Storage interface {
	Set(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
}

type MemStorage struct {
	mapping map[string]string
	mutex sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mapping: make(map[string]string),
	}
}

func (s *MemStorage) Set(key []byte, value []byte) error {
	k := string(key)
	v := string(value)
	s.mutex.Lock()
	s.mapping[k] = v
	s.mutex.Unlock()
	return nil
}

func (s *MemStorage) Get(key []byte) ([]byte, error) {
	k := string(key)
	s.mutex.RLock()
	v, ok := s.mapping[k]
	s.mutex.RUnlock()
	if !ok {
		return nil, errors.New("key not found")
	}
	return []byte(v), nil
}
