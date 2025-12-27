package dino_test

import (
	"errors"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/yuppyweb/dino"
)

func TestRegistry_EmptyTag(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag:  "",
		Type: reflect.TypeOf(0),
	}

	registry := new(dino.Registry)

	if err := registry.Register(key, reflect.ValueOf(42)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	if val.Int() != 42 {
		t.Fatalf("expected value to be 42, got %d", val.Int())
	}
}

func TestRegistry_FilledTag(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag:  "test",
		Type: reflect.TypeOf(""),
	}

	registry := new(dino.Registry)

	if err := registry.Register(key, reflect.ValueOf("hello")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	if val.String() != "hello" {
		t.Fatalf("expected value to be 'hello', got %s", val.String())
	}
}

func TestRegistry_DifferentTagsSomeTypes(t *testing.T) {
	t.Parallel()

	key1 := dino.Key{
		Tag:  "",
		Type: reflect.TypeOf(0),
	}

	key2 := dino.Key{
		Tag:  "special",
		Type: reflect.TypeOf(0),
	}

	key3 := dino.Key{
		Tag:  "another",
		Type: reflect.TypeOf(0),
	}

	registry := new(dino.Registry)

	if err := registry.Register(key1, reflect.ValueOf(1)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := registry.Register(key2, reflect.ValueOf(2)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := registry.Register(key3, reflect.ValueOf(3)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val1, err1 := registry.Find(key1)
	if err1 != nil {
		t.Fatalf("expected key1 to be found")
	}

	if val1.Int() != 1 {
		t.Fatalf("expected value for key1 to be 1, got %d", val1.Int())
	}

	val2, err2 := registry.Find(key2)
	if err2 != nil {
		t.Fatalf("expected key2 to be found")
	}

	if val2.Int() != 2 {
		t.Fatalf("expected value for key2 to be 2, got %d", val2.Int())
	}

	val3, err3 := registry.Find(key3)
	if err3 != nil {
		t.Fatalf("expected key3 to be found")
	}

	if val3.Int() != 3 {
		t.Fatalf("expected value for key3 to be 3, got %d", val3.Int())
	}
}

func TestRegistry_OverwriteWithSomeKeys(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag:  "duplicate",
		Type: reflect.TypeOf(0),
	}

	registry := new(dino.Registry)

	if err := registry.Register(key, reflect.ValueOf(100)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	if val.Int() != 100 {
		t.Fatalf("expected value to be 100, got %d", val.Int())
	}

	if err := registry.Register(key, reflect.ValueOf(200)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err = registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	if val.Int() != 200 {
		t.Fatalf("expected value to be 200, got %d", val.Int())
	}
}

func TestRegistry_KeyTypeNil(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag:  "niltype",
		Type: nil,
	}

	registry := new(dino.Registry)

	err := registry.Register(key, reflect.ValueOf(0))
	if !errors.Is(err, dino.ErrKeyTypeNil) {
		t.Fatalf("expected ErrKeyTypeNil, got %v", err)
	}
}

func TestRegistry_InvalidValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		val  reflect.Value
	}{
		{
			name: "Zero Value",
			val:  reflect.Value{},
		},
		{
			name: "Nil Value",
			val:  reflect.ValueOf(nil),
		},
		{
			name: "Non-Reflect Value",
			val:  reflect.ValueOf((error)(nil)),
		},
	}

	key := dino.Key{
		Tag:  "invalid",
		Type: reflect.TypeOf(0),
	}

	registry := new(dino.Registry)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := registry.Register(key, tc.val)
			if !errors.Is(err, dino.ErrInvalidValue) {
				t.Fatalf("expected ErrInvalidValue, got %v", err)
			}
		})
	}
}

func TestRegistry_ValueNotFound(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag:  "missing",
		Type: reflect.TypeOf(""),
	}

	registry := new(dino.Registry)

	val, err := registry.Find(key)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound, got %v", err)
	}

	if val != reflect.Zero(key.Type) {
		t.Fatalf("expected value to be zero value, got %v", val)
	}
}

func TestRegistry_InvalidValueStored(t *testing.T) {
	t.Parallel()

	key := dino.Key{
		Tag:  "invalid",
		Type: reflect.TypeOf(0),
	}

	registry := new(dino.Registry)
	registry.MockRegister(key, "this is not a reflect.Value")

	val, err := registry.Find(key)
	if !errors.Is(err, dino.ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}

	if val != reflect.Zero(key.Type) {
		t.Fatalf("expected value to be zero value, got %v", val)
	}
}

func TestRegistry_DifferentTypesSameTag(t *testing.T) {
	t.Parallel()

	tag := "shared"

	key1 := dino.Key{
		Tag:  tag,
		Type: reflect.TypeOf(0),
	}

	key2 := dino.Key{
		Tag:  tag,
		Type: reflect.TypeOf(""),
	}

	registry := new(dino.Registry)

	if err := registry.Register(key1, reflect.ValueOf(123)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := registry.Register(key2, reflect.ValueOf("abc")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val1, err1 := registry.Find(key1)
	if err1 != nil {
		t.Fatalf("expected key1 to be found")
	}

	if val1.Int() != 123 {
		t.Fatalf("expected value for key1 to be 123, got %d", val1.Int())
	}

	val2, err2 := registry.Find(key2)
	if err2 != nil {
		t.Fatalf("expected key2 to be found")
	}

	if val2.String() != "abc" {
		t.Fatalf("expected value for key2 to be 'abc', got %s", val2.String())
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	registry := new(dino.Registry)
	keyChan := make(chan dino.Key, 100)

	for idx := range 100 {
		wg.Go(func() {
			key := dino.Key{
				Tag:  strconv.Itoa(idx),
				Type: reflect.TypeOf(idx),
			}

			if err := registry.Register(key, reflect.ValueOf(idx)); err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			keyChan <- key
		})

		wg.Go(func() {
			key := <-keyChan

			val, err := registry.Find(key)
			if err != nil {
				t.Errorf("expected key to be found for goroutine %d", idx)
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
