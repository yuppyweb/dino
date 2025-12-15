package dino_test

import (
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/yuppyweb/dino"
)

func TestRegistry_EmptyTag(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(0),
	}

	registry := &dino.Registry{}
	registry.Add(key, reflect.ValueOf(42))

	val, ok := registry.Get(key)
	if !ok {
		t.Fatalf("expected key to be found")
	}

	if val.Int() != 42 {
		t.Fatalf("expected value to be 42, got %d", val.Int())
	}
}

func TestRegistry_FilledTag(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag: "test",
		Ref: reflect.TypeOf(""),
	}

	registry := &dino.Registry{}
	registry.Add(key, reflect.ValueOf("hello"))

	val, ok := registry.Get(key)
	if !ok {
		t.Fatalf("expected key to be found")
	}

	if val.String() != "hello" {
		t.Fatalf("expected value to be 'hello', got %s", val.String())
	}
}

func TestRegistry_DifferentTagsSomeTypes(t *testing.T) {
	t.Parallel()

	key1 := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(0),
	}

	key2 := dino.Key{
		Tag: "special",
		Ref: reflect.TypeOf(0),
	}

	key3 := dino.Key{
		Tag: "another",
		Ref: reflect.TypeOf(0),
	}

	registry := &dino.Registry{}
	registry.Add(key1, reflect.ValueOf(1))
	registry.Add(key2, reflect.ValueOf(2))
	registry.Add(key3, reflect.ValueOf(3))

	val1, ok1 := registry.Get(key1)
	if !ok1 {
		t.Fatalf("expected key1 to be found")
	}

	if val1.Int() != 1 {
		t.Fatalf("expected value for key1 to be 1, got %d", val1.Int())
	}

	val2, ok2 := registry.Get(key2)
	if !ok2 {
		t.Fatalf("expected key2 to be found")
	}

	if val2.Int() != 2 {
		t.Fatalf("expected value for key2 to be 2, got %d", val2.Int())
	}

	val3, ok3 := registry.Get(key3)
	if !ok3 {
		t.Fatalf("expected key3 to be found")
	}

	if val3.Int() != 3 {
		t.Fatalf("expected value for key3 to be 3, got %d", val3.Int())
	}
}

func TestRegistry_OverwriteWithSomeKeys(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag: "duplicate",
		Ref: reflect.TypeOf(0),
	}

	registry := &dino.Registry{}
	registry.Add(key, reflect.ValueOf(100))

	val, ok := registry.Get(key)
	if !ok {
		t.Fatalf("expected key to be found")
	}

	if val.Int() != 100 {
		t.Fatalf("expected value to be 100, got %d", val.Int())
	}

	registry.Add(key, reflect.ValueOf(200))

	val, ok = registry.Get(key)
	if !ok {
		t.Fatalf("expected key to be found")
	}

	if val.Int() != 200 {
		t.Fatalf("expected value to be 200, got %d", val.Int())
	}
}

func TestRegistry_KeyNotFound(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag: "missing",
		Ref: reflect.TypeOf(""),
	}

	registry := &dino.Registry{}

	val, ok := registry.Get(key)
	if ok {
		t.Fatalf("expected key to not be found")
	}

	if val != reflect.Zero(key.Ref) {
		t.Fatalf("expected value to be zero value, got %v", val)
	}
}

func TestRegistry_DifferentTypesSameTag(t *testing.T) {
	t.Parallel()

	tag := "shared"

	key1 := dino.Key{
		Tag: tag,
		Ref: reflect.TypeOf(0),
	}

	key2 := dino.Key{
		Tag: tag,
		Ref: reflect.TypeOf(""),
	}

	registry := &dino.Registry{}
	registry.Add(key1, reflect.ValueOf(123))
	registry.Add(key2, reflect.ValueOf("abc"))

	val1, ok1 := registry.Get(key1)
	if !ok1 {
		t.Fatalf("expected key1 to be found")
	}

	if val1.Int() != 123 {
		t.Fatalf("expected value for key1 to be 123, got %d", val1.Int())
	}

	val2, ok2 := registry.Get(key2)
	if !ok2 {
		t.Fatalf("expected key2 to be found")
	}

	if val2.String() != "abc" {
		t.Fatalf("expected value for key2 to be 'abc', got %s", val2.String())
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	registry := &dino.Registry{}
	keyChan := make(chan dino.Key, 100)

	for i := range 100 {
		wg.Go(func() {
			key := dino.Key{
				Tag: strconv.Itoa(i),
				Ref: reflect.TypeOf(i),
			}

			registry.Add(key, reflect.ValueOf(i))
			keyChan <- key
		})

		wg.Go(func() {
			key := <-keyChan

			val, ok := registry.Get(key)
			if !ok {
				t.Errorf("expected key to be found for goroutine %d", i)
			}

			num, _ := strconv.Atoi(key.Tag)

			if val.Int() != int64(num) {
				t.Errorf("expected value to be %d for goroutine %d, got %d", num, num, val.Int())
			}
		})
	}

	wg.Wait()
	close(keyChan)
}
