package dino_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yuppyweb/dino"
)

//go:generate go tool cp ./dino_mock.tmpl ./dino_mock.go

func TestDino_WithDefaulRegistry(t *testing.T) {
	t.Parallel()

	di := dino.New()
	registry := di.MockRegistry()

	if _, ok := registry.(*dino.SyncMapRegistry); !ok {
		t.Fatalf("expected default registry to be of type SyncMapRegistry")
	}
}

func TestDino_WithCustomRegistry(t *testing.T) {
	t.Parallel()

	di := dino.New()
	di.WithRegistry(NewMockRegistry())
	registry := di.MockRegistry()

	if _, ok := registry.(*MockRegistry); !ok {
		t.Fatalf("expected custom registry to be of type MockRegistry")
	}
}

func TestDino_FactoryNilFunction(t *testing.T) {
	t.Parallel()

	di := dino.New()

	err := di.Factory(nil)
	if !errors.Is(err, dino.ErrInvalidInputValue) {
		t.Fatalf("expected ErrInvalidInputValue, got %v", err)
	}

	if !contains(err.Error(), "factory function cannot be nil") {
		t.Fatalf("expected error message to contain 'factory function cannot be nil', got %s", err.Error())
	}
}

func TestDino_FactoryNotFunction(t *testing.T) {
	t.Parallel()

	di := dino.New()

	err := di.Factory(42)
	if !errors.Is(err, dino.ErrInvalidInputValue) {
		t.Fatalf("expected ErrInvalidInputValue, got %v", err)
	}

	if !contains(err.Error(), "factory expected a function") {
		t.Fatalf("expected error message to contain 'factory expected a function', got %s", err.Error())
	}
}

func TestDino_FactorySingleOutputWithoutTags(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	service := &SimpleService{
		Value: "test",
	}

	di := dino.New()

	factory := func() *SimpleService {
		return service
	}

	err := di.Factory(factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(service),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(func() *SimpleService)
	if !ok {
		t.Fatalf("expected value to be of type *SimpleService")
	}

	res := result()
	if res != service {
		t.Fatalf("expected service to be %v, got %v", service, res)
	}
}

func TestDino_FactorySingleOutputWithTags(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	service := &SimpleService{
		Value: "tagged",
	}

	di := dino.New()

	factory := func() *SimpleService {
		return service
	}

	err := di.Factory(factory, "tag1", "tag2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()

	key1 := dino.RegistryKey{
		Tag:  "tag1",
		Type: reflect.TypeOf(service),
	}

	key2 := dino.RegistryKey{
		Tag:  "tag2",
		Type: reflect.TypeOf(service),
	}

	emptyKey := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(service),
	}

	val1, err := registry.Find(key1)
	if err != nil {
		t.Fatalf("expected key1 to be found")
	}

	val2, err := registry.Find(key2)
	if err != nil {
		t.Fatalf("expected key2 to be found")
	}

	_, err = registry.Find(emptyKey)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound for empty tag, got %v", err)
	}

	result1, ok := val1.Interface().(func() *SimpleService)
	if !ok {
		t.Fatalf("expected value1 to be of type *SimpleService")
	}

	result2, ok := val2.Interface().(func() *SimpleService)
	if !ok {
		t.Fatalf("expected value2 to be of type *SimpleService")
	}

	res1 := result1()
	if res1 != service {
		t.Fatalf("expected service to be %v, got %v", service, res1)
	}

	res2 := result2()
	if res2 != service {
		t.Fatalf("expected service to be %v, got %v", service, res2)
	}
}

func TestDino_FactoryMultipleOutputs(t *testing.T) {
	t.Parallel()

	type ServiceA struct{}
	type ServiceB struct{}

	srvA := &ServiceA{}
	srvB := &ServiceB{}

	di := dino.New()

	factory := func() (*ServiceA, *ServiceB) {
		return srvA, srvB
	}

	err := di.Factory(factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()

	keyA := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(srvA),
	}

	keyB := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(srvB),
	}

	valA, err := registry.Find(keyA)
	if err != nil {
		t.Fatalf("expected keyA to be found")
	}

	valB, err := registry.Find(keyB)
	if err != nil {
		t.Fatalf("expected keyB to be found")
	}

	resultA, ok := valA.Interface().(func() (*ServiceA, *ServiceB))
	if !ok {
		t.Fatalf("expected valueA to be of type *ServiceA")
	}

	resultB, ok := valB.Interface().(func() (*ServiceA, *ServiceB))
	if !ok {
		t.Fatalf("expected valueB to be of type *ServiceB")
	}

	resA1, resB1 := resultA()
	if resA1 != srvA {
		t.Fatalf("expected serviceA to be %v, got %v", srvA, resA1)
	}

	if resB1 != srvB {
		t.Fatalf("expected serviceB to be %v, got %v", srvB, resB1)
	}

	resA2, resB2 := resultB()
	if resA2 != srvA {
		t.Fatalf("expected serviceA to be %v, got %v", srvA, resA2)
	}

	if resB2 != srvB {
		t.Fatalf("expected serviceB to be %v, got %v", srvB, resB2)
	}
}

func TestDino_FactoryWithNilError(t *testing.T) {
	t.Parallel()
}

func contains(str, substr string) bool {
	for i := range len(str) - len(substr) + 1 {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
