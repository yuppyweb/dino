package dino_test

import (
	"errors"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/yuppyweb/dino"
)

type MockRegistry struct {
	RegisterOn []struct {
		Key   dino.RegistryKey
		Value reflect.Value
	}
	RegisterOut []error
	FindOn      []dino.RegistryKey
	FindOut     []struct {
		Value reflect.Value
		Err   error
	}
	numRegOut  int
	numFindOut int
}

func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		RegisterOn: []struct {
			Key   dino.RegistryKey
			Value reflect.Value
		}{},
		RegisterOut: []error{},
		FindOn:      []dino.RegistryKey{},
		FindOut: []struct {
			Value reflect.Value
			Err   error
		}{},
		numRegOut:  0,
		numFindOut: 0,
	}
}

func (m *MockRegistry) Register(key dino.RegistryKey, value reflect.Value) error {
	m.RegisterOn = append(m.RegisterOn, struct {
		Key   dino.RegistryKey
		Value reflect.Value
	}{
		Key:   key,
		Value: value,
	})

	defer func() {
		m.numRegOut++
	}()

	return m.RegisterOut[m.numRegOut]
}

func (m *MockRegistry) Find(key dino.RegistryKey) (reflect.Value, error) {
	m.FindOn = append(m.FindOn, key)

	defer func() {
		m.numFindOut++
	}()

	return m.FindOut[m.numFindOut].Value, m.FindOut[m.numFindOut].Err
}

var _ dino.Registry = (*MockRegistry)(nil)

func TestRegistry_EmptyTag(t *testing.T) {
	t.Parallel()

	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeFor[int](),
	}

	registry := new(dino.SyncMapRegistry)

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

	key := dino.RegistryKey{
		Tag:  "test",
		Type: reflect.TypeFor[string](),
	}

	registry := new(dino.SyncMapRegistry)

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

	key1 := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeFor[int](),
	}

	key2 := dino.RegistryKey{
		Tag:  "special",
		Type: reflect.TypeFor[int](),
	}

	key3 := dino.RegistryKey{
		Tag:  "another",
		Type: reflect.TypeFor[int](),
	}

	registry := new(dino.SyncMapRegistry)

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

	key := dino.RegistryKey{
		Tag:  "duplicate",
		Type: reflect.TypeFor[int](),
	}

	registry := new(dino.SyncMapRegistry)

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

func TestRegistry_RegisterKeyTypeNil(t *testing.T) {
	t.Parallel()

	key := dino.RegistryKey{
		Tag:  "niltype",
		Type: nil,
	}

	registry := new(dino.SyncMapRegistry)

	err := registry.Register(key, reflect.ValueOf(0))
	if !errors.Is(err, dino.ErrKeyTypeNil) {
		t.Fatalf("expected ErrKeyTypeNil, got %v", err)
	}
}

func TestRegistry_FindKeyTypeNil(t *testing.T) {
	t.Parallel()

	key := dino.RegistryKey{
		Tag:  "niltype",
		Type: nil,
	}

	registry := new(dino.SyncMapRegistry)

	val, err := registry.Find(key)

	if !errors.Is(err, dino.ErrKeyTypeNil) {
		t.Fatalf("expected ErrKeyTypeNil, got %v", err)
	}

	if val != (reflect.Value{}) {
		t.Fatalf("expected zero reflect.Value, got %v", val)
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

	key := dino.RegistryKey{
		Tag:  "invalid",
		Type: reflect.TypeFor[int](),
	}

	registry := new(dino.SyncMapRegistry)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := registry.Register(key, tc.val)
			if !errors.Is(err, dino.ErrInvalidValue) {
				t.Fatalf("expected ErrInvalidValue, got %v", err)
			}
		})
	}
}

func TestRegistry_ValueNotFound(t *testing.T) {
	t.Parallel()

	key := dino.RegistryKey{
		Tag:  "missing",
		Type: reflect.TypeFor[string](),
	}

	registry := new(dino.SyncMapRegistry)

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

	key := dino.RegistryKey{
		Tag:  "invalid",
		Type: reflect.TypeFor[int](),
	}

	registry := new(dino.SyncMapRegistry)
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

	key1 := dino.RegistryKey{
		Tag:  tag,
		Type: reflect.TypeFor[int](),
	}

	key2 := dino.RegistryKey{
		Tag:  tag,
		Type: reflect.TypeFor[string](),
	}

	registry := new(dino.SyncMapRegistry)

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

	registry := new(dino.SyncMapRegistry)
	keyChan := make(chan dino.RegistryKey, 100)

	for idx := range 100 {
		wg.Go(func() {
			key := dino.RegistryKey{
				Tag:  strconv.Itoa(idx),
				Type: reflect.TypeFor[int](),
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
