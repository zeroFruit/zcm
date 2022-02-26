package container

import (
	"errors"
	"fmt"
	"sync"
)

type Iterator interface {
	HasNext() bool
	Next() *Handle
}

type defaultIter struct {
	curr    int
	handles []*Handle
}

func newDefaultIter(handles []*Handle) *defaultIter {
	return &defaultIter{
		curr:    0,
		handles: handles,
	}
}

func (i *defaultIter) HasNext() bool {
	return i.curr < len(i.handles)-1
}

func (i *defaultIter) Next() *Handle {
	if i.curr == len(i.handles)-1 {
		return nil
	}
	result := i.handles[i.curr]
	i.curr += 1
	return result
}

type ReadOnlyStore interface {
	Get(id Id) (*Handle, error)
}

type Store interface {
	ReadOnlyStore
	Put(*Handle) error
	Iter() Iterator
}

type InMemStore struct {
	lock    *sync.RWMutex
	handles map[Id]*Handle
}

func NewInMemStore() *InMemStore {
	return &InMemStore{
		lock:    &sync.RWMutex{},
		handles: make(map[Id]*Handle),
	}
}

func (s *InMemStore) Put(handle *Handle) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if handle.id == "" {
		return errors.New("cannot put handle because id is not defined")
	}
	s.handles[handle.id] = handle
	return nil
}
func (s *InMemStore) Get(id Id) (*Handle, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	handle, ok := s.handles[id]
	if !ok {
		return nil, fmt.Errorf("cannot find handle. id [%s]", id)
	}
	return handle, nil
}

func (s *InMemStore) Iter() Iterator {
	items := make([]*Handle, len(s.handles))
	for _, h := range s.handles {
		items = append(items, h)
	}
	return newDefaultIter(items)
}
