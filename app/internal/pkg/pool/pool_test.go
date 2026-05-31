package pool

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testObj struct {
	Value int
	Data  []byte
}

func (t *testObj) Reset() {
	t.Value = 0
	t.Data = t.Data[:0]
}

func TestPool_GetReturnsNewObject(t *testing.T) {
	p := New(func() *testObj { return &testObj{} })

	obj := p.Get()
	require.NotNil(t, obj)
	assert.Equal(t, 0, obj.Value)
}

func TestPool_PutResetsBeforeReuse(t *testing.T) {
	p := New(func() *testObj { return &testObj{} })

	obj := p.Get()
	obj.Value = 42
	obj.Data = append(obj.Data, 1, 2, 3)
	p.Put(obj)

	reused := p.Get()
	assert.Equal(t, 0, reused.Value)
	assert.Empty(t, reused.Data)
}

func TestPool_ConcurrentAccess(t *testing.T) {
	p := New(func() *testObj { return &testObj{} })

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := p.Get()
			obj.Value = 1
			obj.Data = append(obj.Data, 0xFF)
			p.Put(obj)
		}()
	}
	wg.Wait()
}
