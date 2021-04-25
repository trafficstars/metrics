package syncmap

import (
	"sync"

	"github.com/xaionaro-go/atomicmap"
)

type Map struct {
	sync.Map
}

func (m *Map) GetByBytes(b []byte) (interface{}, error) {
	return m.Get(b)
}

func (m *Map) Unset(k interface{}) error {
	_, ok := m.Map.LoadAndDelete(string(k.([]byte)))
	if ok {
		return nil
	} else {
		return atomicmap.NotFound
	}
}

func (m *Map) Keys() []interface{} {
	var r []interface{}
	m.Map.Range(func(key, value interface{}) bool {
		r = append(r, []byte(key.(string)))
		return true
	})
	return r
}

func (m *Map) Get(k interface{}) (interface{}, error) {
	v, ok := m.Map.Load(string(k.([]byte)))
	if ok {
		return v, nil
	}
	return v, atomicmap.NotFound
}

func (m *Map) Set(k, v interface{}) error {
	m.Map.Store(string(k.([]byte)), v)
	return nil
}
