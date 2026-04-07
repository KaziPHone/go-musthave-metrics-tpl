package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// testObject — тестовая структура для проверки Pool
type testObject struct {
	value int
	count int
}

func (t *testObject) Reset() {
	t.value = 0
	t.count = 0
}

func TestPool_GetAndPut(t *testing.T) {
	createCount := 0
	p := New(func() *testObject {
		createCount++
		return &testObject{}
	})

	// При первом Get должен быть создан новый объект
	obj1 := p.Get()
	assert.NotNil(t, obj1)
	assert.Equal(t, 1, createCount)

	// Возврат объекта в пул не должен создавать новый объект
	p.Put(obj1)
	assert.Equal(t, 1, createCount)

	// Следующий Get должен вернуть объект из пула (не создавать новый)
	obj2 := p.Get()
	assert.Equal(t, obj1, obj2)
	assert.Equal(t, 1, createCount)

	// Объект должен быть сброшен после Put
	assert.Equal(t, 0, obj2.value)
	assert.Equal(t, 0, obj2.count)
}

func TestPool_PutCallsReset(t *testing.T) {
	p := New(func() *testObject {
		return &testObject{value: 42, count: 100}
	})

	obj := p.Get()
	assert.Equal(t, 42, obj.value)
	assert.Equal(t, 100, obj.count)

	// После Put объект должен быть сброшен
	p.Put(obj)

	// Получаем тот же объект обратно
	obj2 := p.Get()
	assert.Equal(t, 0, obj2.value)
	assert.Equal(t, 0, obj2.count)
}

func TestPool_MultipleObjects(t *testing.T) {
	p := New(func() *testObject {
		return &testObject{}
	})

	// Получаем и возвращаем несколько объектов
	objs := make([]*testObject, 5)
	for i := 0; i < 5; i++ {
		objs[i] = p.Get()
		objs[i].value = i
		objs[i].count = i * 2
	}

	for i := 0; i < 5; i++ {
		p.Put(objs[i])
	}

	// Проверяем, что все объекты сброшены
	for i := 0; i < 5; i++ {
		obj := p.Get()
		assert.Equal(t, 0, obj.value)
		assert.Equal(t, 0, obj.count)
	}
}

func TestPool_FactoryCalledWhenPoolEmpty(t *testing.T) {
	createCount := 0
	p := New(func() *testObject {
		createCount++
		return &testObject{}
	})

	// Первый Get — создает новый объект
	obj1 := p.Get()
	assert.Equal(t, 1, createCount)

	// Возвращаем объект в пул
	p.Put(obj1)
	assert.Equal(t, 1, createCount)

	// Берем снова — из пула, создание не происходит
	obj2 := p.Get()
	assert.Equal(t, 1, createCount)
	assert.Equal(t, obj1, obj2)

	// Возвращаем и берем еще раз
	p.Put(obj2)
	obj3 := p.Get()
	assert.Equal(t, 1, createCount)
	assert.Equal(t, obj2, obj3)
}
