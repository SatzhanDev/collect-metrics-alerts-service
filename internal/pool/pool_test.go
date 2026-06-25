package pool

import (
	"testing"
)

type MyStruct struct {
	Name  string
	Count int
}

func (m *MyStruct) Reset() {
	m.Name = ""
	m.Count = 0
}

func TestPoolGetPut(t *testing.T) {
	p := New(func() *MyStruct {
		return &MyStruct{}
	})

	obj := p.Get()
	obj.Name = "test"
	obj.Count = 42
	p.Put(obj)

	obj2 := p.Get()
	if obj2.Name != "" {
		t.Errorf("Name не сброшен: %q", obj2.Name)
	}
	if obj2.Count != 0 {
		t.Errorf("Count не сброшен: %d", obj2.Count)
	}
}
