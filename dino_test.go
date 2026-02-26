package dino_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
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

	if !strings.Contains(err.Error(), "factory function cannot be nil") {
		t.Fatalf(
			"expected error message to contain 'factory function cannot be nil', got %s",
			err.Error(),
		)
	}
}

func TestDino_FactoryNotFunction(t *testing.T) {
	t.Parallel()

	di := dino.New()

	err := di.Factory(42)
	if !errors.Is(err, dino.ErrInvalidInputValue) {
		t.Fatalf("expected ErrInvalidInputValue, got %v", err)
	}

	if !strings.Contains(err.Error(), "factory expected a function") {
		t.Fatalf(
			"expected error message to contain 'factory expected a function', got %s",
			err.Error(),
		)
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

	factory := func() error {
		return nil
	}

	registry := NewMockRegistry()

	di := dino.New()
	di = di.WithRegistry(registry)

	err := di.Factory(factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(registry.RegisterOn) != 0 {
		t.Fatalf("expected no registrations in registry, got %d", len(registry.RegisterOn))
	}
}

func TestDino_FactoryWithOnlyError(t *testing.T) {
	t.Parallel()

	factory := func() error {
		return errors.New("some error")
	}

	registry := NewMockRegistry()

	di := dino.New()
	di = di.WithRegistry(registry)

	err := di.Factory(factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(registry.RegisterOn) != 0 {
		t.Fatalf("expected no registrations in registry, got %d", len(registry.RegisterOn))
	}
}

func TestDino_FactoryWithErrorAndOtherOutputs(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	srv := &SimpleService{
		Value: "test",
	}

	factory := func() (*SimpleService, error) {
		return srv, nil
	}

	registry := NewMockRegistry()
	registry.RegisterOut = append(registry.RegisterOut, nil)

	di := dino.New()
	di = di.WithRegistry(registry)

	err := di.Factory(factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(registry.RegisterOn) != 1 {
		t.Fatalf("expected 1 registration in registry, got %d", len(registry.RegisterOn))
	}

	registeredValue := registry.RegisterOn[0].Value

	resultFunc, ok := registeredValue.Interface().(func() (*SimpleService, error))
	if !ok {
		t.Fatalf("expected registered value to be of type func() (*SimpleService, error)")
	}

	result, err := resultFunc()
	if err != nil {
		t.Fatalf("unexpected error from result function: %v", err)
	}

	if result != srv {
		t.Fatalf("expected service to be %v, got %v", srv, result)
	}
}

func TestDino_FactoryBindError(t *testing.T) {
	t.Parallel()

	type SimpleService struct{}

	factory := func() *SimpleService {
		return &SimpleService{}
	}

	registry := NewMockRegistry()
	registry.RegisterOut = append(registry.RegisterOut, errors.New("some bind error"))

	di := dino.New()
	di = di.WithRegistry(registry)

	err := di.Factory(factory)
	if err == nil {
		t.Fatalf("expected error from Factory, got nil")
	}

	if !strings.Contains(err.Error(), "some bind error") {
		t.Fatalf("expected error message to contain 'some bind error', got %s", err.Error())
	}

	if !strings.Contains(err.Error(), "failed to bind factory function output") {
		t.Fatalf(
			"expected error message to contain 'failed to bind factory function output', got %s",
			err.Error(),
		)
	}
}

func TestDino_FactoryIntegerWithoutTag(t *testing.T) {
	t.Parallel()

	value := 42
	di := dino.New()

	err := di.Factory(func() int {
		return value
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(func() int)
	if !ok {
		t.Fatalf("expected value to be of type func() int")
	}

	res := result()
	if res != value {
		t.Fatalf("expected result to be %d, got %d", value, res)
	}
}

func TestDino_FactoryIntegerWithTag(t *testing.T) {
	t.Parallel()

	value := 100
	di := dino.New()

	err := di.Factory(func() int {
		return value
	}, "intTag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key1 := dino.RegistryKey{
		Tag:  "intTag",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key1)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(func() int)
	if !ok {
		t.Fatalf("expected value to be of type func() int")
	}

	res := result()
	if res != value {
		t.Fatalf("expected result to be %d, got %d", value, res)
	}

	key2 := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	_, err = registry.Find(key2)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound for empty tag, got %v", err)
	}
}

func TestDino_FactoryStringWithoutTag(t *testing.T) {
	t.Parallel()

	value := "hello dino"
	di := dino.New()

	err := di.Factory(func() string {
		return value
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(func() string)
	if !ok {
		t.Fatalf("expected value to be of type func() string")
	}

	res := result()
	if res != value {
		t.Fatalf("expected result to be '%s', got '%s'", value, res)
	}
}

func TestDino_FactoryStringWithTag(t *testing.T) {
	t.Parallel()

	value := "tagged string"
	di := dino.New()

	err := di.Factory(func() string {
		return value
	}, "stringTag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key1 := dino.RegistryKey{
		Tag:  "stringTag",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key1)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(func() string)
	if !ok {
		t.Fatalf("expected value to be of type func() string")
	}

	res := result()
	if res != value {
		t.Fatalf("expected result to be '%s', got '%s'", value, res)
	}

	key2 := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	_, err = registry.Find(key2)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound for empty tag, got %v", err)
	}
}

func TestDino_FactoryConcurrentAccess(t *testing.T) {
	t.Parallel()

	di := dino.New()
	wg := sync.WaitGroup{}

	for i := range 100 {
		wg.Go(func() {
			err := di.Factory(func() int {
				return i
			}, fmt.Sprintf("concurrentTag%d", i))
			if err != nil {
				t.Fatalf("unexpected error during factory registration: %v", err)
			}
		})
	}

	wg.Wait()

	for idx := range 100 {
		keyNum := dino.RegistryKey{
			Tag:  fmt.Sprintf("concurrentTag%d", idx),
			Type: reflect.TypeOf(0),
		}

		registry := di.MockRegistry()

		val, err := registry.Find(keyNum)
		if err != nil {
			t.Fatalf("expected key %v to be found", keyNum)
		}

		result, ok := val.Interface().(func() int)
		if !ok {
			t.Fatalf("expected value to be of type func() int")
		}

		res := result()
		if res != idx {
			t.Fatalf("expected result to be %d, got %d", idx, res)
		}
	}
}

func TestDino_SingletonNilValue(t *testing.T) {
	t.Parallel()

	di := dino.New()

	err := di.Singleton(nil)
	if !errors.Is(err, dino.ErrInvalidInputValue) {
		t.Fatalf("expected ErrInvalidInputValue, got %v", err)
	}

	if !strings.Contains(err.Error(), "singleton value cannot be nil") {
		t.Fatalf(
			"expected error message to contain 'singleton value cannot be nil', got %s",
			err.Error(),
		)
	}
}

func TestDino_SingletonWithoutTags(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	srv := &SimpleService{
		Value: "singleton",
	}

	di := dino.New()

	err := di.Singleton(srv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(srv),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected value to be of type *SimpleService")
	}

	if result != srv {
		t.Fatalf("expected service to be %v, got %v", srv, result)
	}
}

func TestDino_SingletonWithTags(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	srv := &SimpleService{
		Value: "tagged singleton",
	}

	di := dino.New()

	err := di.Singleton(srv, "tagA", "tagB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()

	keyA := dino.RegistryKey{
		Tag:  "tagA",
		Type: reflect.TypeOf(srv),
	}

	keyB := dino.RegistryKey{
		Tag:  "tagB",
		Type: reflect.TypeOf(srv),
	}

	emptyKey := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(srv),
	}

	valA, err := registry.Find(keyA)
	if err != nil {
		t.Fatalf("expected keyA to be found")
	}

	valB, err := registry.Find(keyB)
	if err != nil {
		t.Fatalf("expected keyB to be found")
	}

	_, err = registry.Find(emptyKey)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound for empty tag, got %v", err)
	}

	resultA, ok := valA.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected valueA to be of type *SimpleService")
	}

	resultB, ok := valB.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected valueB to be of type *SimpleService")
	}

	if resultA != srv {
		t.Fatalf("expected serviceA to be %v, got %v", srv, resultA)
	}

	if resultB != srv {
		t.Fatalf("expected serviceB to be %v, got %v", srv, resultB)
	}
}

func TestDino_SingletonFunctionValue(t *testing.T) {
	t.Parallel()

	di := dino.New()

	err := di.Singleton(func() int {
		return 7
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(func() int { return 0 }),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(func() int)
	if !ok {
		t.Fatalf("expected value to be of type func() int")
	}

	res := result()
	if res != 7 {
		t.Fatalf("expected result to be 7, got %d", res)
	}
}

func TestDino_SingletonIntegerWithoutTag(t *testing.T) {
	t.Parallel()

	value := 256
	di := dino.New()

	err := di.Singleton(value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(int)
	if !ok {
		t.Fatalf("expected value to be of type int")
	}

	if result != value {
		t.Fatalf("expected result to be %d, got %d", value, result)
	}
}

func TestDino_SingletonIntegerWithTag(t *testing.T) {
	t.Parallel()

	value := 512
	di := dino.New()

	err := di.Singleton(value, "intSingletonTag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key1 := dino.RegistryKey{
		Tag:  "intSingletonTag",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key1)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(int)
	if !ok {
		t.Fatalf("expected value to be of type int")
	}

	if result != value {
		t.Fatalf("expected result to be %d, got %d", value, result)
	}

	key2 := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	_, err = registry.Find(key2)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound for empty tag, got %v", err)
	}
}

func TestDino_SingletonStringWithoutTag(t *testing.T) {
	t.Parallel()

	value := "singleton string"
	di := dino.New()

	err := di.Singleton(value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(string)
	if !ok {
		t.Fatalf("expected value to be of type string")
	}

	if result != value {
		t.Fatalf("expected result to be '%s', got '%s'", value, result)
	}
}

func TestDino_SingletonStringWithTag(t *testing.T) {
	t.Parallel()

	value := "tagged singleton string"
	di := dino.New()

	err := di.Singleton(value, "stringSingletonTag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key1 := dino.RegistryKey{
		Tag:  "stringSingletonTag",
		Type: reflect.TypeOf(value),
	}

	val, err := registry.Find(key1)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(string)
	if !ok {
		t.Fatalf("expected value to be of type string")
	}

	if result != value {
		t.Fatalf("expected result to be '%s', got '%s'", value, result)
	}

	key2 := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(value),
	}

	_, err = registry.Find(key2)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound for empty tag, got %v", err)
	}
}

func TestDino_SingletonErrorValue(t *testing.T) {
	t.Parallel()

	di := dino.New()
	expectedErr := errors.New("singleton error")

	err := di.Singleton(expectedErr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	registry := di.MockRegistry()
	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeOf(expectedErr),
	}

	val, err := registry.Find(key)
	if err != nil {
		t.Fatalf("expected key to be found")
	}

	result, ok := val.Interface().(error)
	if !ok {
		t.Fatalf("expected value to be of type error")
	}

	if !errors.Is(result, expectedErr) {
		t.Fatalf("expected result to be %v, got %v", expectedErr, result)
	}
}

func TestDino_SingletonBindError(t *testing.T) {
	t.Parallel()

	registry := NewMockRegistry()
	expectedErr := errors.New("some bind error")

	di := dino.New()
	di = di.WithRegistry(registry)

	registry.RegisterOut = append(registry.RegisterOut, expectedErr)

	err := di.Singleton(5)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected bind error to be %v, got %v", expectedErr, err)
	}

	if len(registry.RegisterOn) != 1 {
		t.Fatalf("expected 1 registration in registry, got %d", len(registry.RegisterOn))
	}

	if !errors.Is(registry.RegisterOut[0], expectedErr) {
		t.Fatalf("expected bind error to be %v, got %v", expectedErr, registry.RegisterOut[0])
	}
}

func TestDino_SingletonConcurrentAccess(t *testing.T) {
	t.Parallel()

	di := dino.New()
	wg := sync.WaitGroup{}

	for i := range 100 {
		wg.Go(func() {
			err := di.Singleton(i, fmt.Sprintf("singletonTag%d", i))
			if err != nil {
				t.Fatalf("unexpected error during singleton registration: %v", err)
			}
		})
	}

	wg.Wait()

	for idx := range 100 {
		keyNum := dino.RegistryKey{
			Tag:  fmt.Sprintf("singletonTag%d", idx),
			Type: reflect.TypeFor[int](),
		}

		registry := di.MockRegistry()

		val, err := registry.Find(keyNum)
		if err != nil {
			t.Fatalf("expected key %v to be found", keyNum)
		}

		result, ok := val.Interface().(int)
		if !ok {
			t.Fatalf("expected value to be of type int")
		}

		if result != idx {
			t.Fatalf("expected result to be %d, got %d", idx, result)
		}
	}
}

func TestDino_InjectNilTarget(t *testing.T) {
	t.Parallel()

	di := dino.New()

	err := di.Inject(nil)
	if !errors.Is(err, dino.ErrInvalidInputValue) {
		t.Fatalf("expected ErrInvalidInputValue, got %v", err)
	}

	if !strings.Contains(err.Error(), "inject target cannot be nil") {
		t.Fatalf(
			"expected error message to contain 'inject target cannot be nil', got %s",
			err.Error(),
		)
	}
}

func TestDino_InjectNotStruct(t *testing.T) {
	t.Parallel()

	di := dino.New()

	err := di.Inject(42)
	if !errors.Is(err, dino.ErrExpectedStruct) {
		t.Fatalf("expected ErrExpectedStruct, got %v", err)
	}

	if !strings.Contains(err.Error(), "failed to inject dependencies:") {
		t.Fatalf(
			"expected error message to contain 'failed to inject dependencies:', got %s",
			err.Error(),
		)
	}
}

func TestDino_InjectUnregisteredSingleDependency(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type Consumer struct {
		A *ServiceA
	}

	di := dino.New()
	consumer := new(Consumer)

	err := di.Inject(consumer)
	if err != nil {
		t.Fatalf("unexpected error during injection: %v", err)
	}

	if consumer.A == nil {
		t.Fatalf("expected ServiceA to be injected, got nil")
	}

	if consumer.A.Value != "" {
		t.Fatalf("expected ServiceA.Value to be empty, got '%s'", consumer.A.Value)
	}
}

func TestDino_InjectUnregisteredSingleDependencyWithTag(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type Consumer struct {
		A *ServiceA `inject:"tagged"`
	}

	di := dino.New()
	consumer := new(Consumer)

	err := di.Inject(consumer)
	if err != nil {
		t.Fatalf("unexpected error during injection: %v", err)
	}

	if consumer.A == nil {
		t.Fatalf("expected ServiceA to be injected, got nil")
	}

	if consumer.A.Value != "" {
		t.Fatalf("expected ServiceA.Value to be empty, got '%s'", consumer.A.Value)
	}
}

func TestDino_InjectUnregisteredMultipleDependency(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type ServiceB struct {
		Number int
	}

	type Consumer struct {
		A *ServiceA
		B *ServiceB
	}

	di := dino.New()
	consumer := new(Consumer)

	err := di.Inject(consumer)
	if err != nil {
		t.Fatalf("unexpected error during injection: %v", err)
	}

	if consumer.A == nil {
		t.Fatalf("expected ServiceA to be injected, got nil")
	}

	if consumer.A.Value != "" {
		t.Fatalf("expected ServiceA.Value to be empty, got '%s'", consumer.A.Value)
	}

	if consumer.B == nil {
		t.Fatalf("expected ServiceB to be injected, got nil")
	}

	if consumer.B.Number != 0 {
		t.Fatalf("expected ServiceB.Number to be 0, got %d", consumer.B.Number)
	}
}

func TestDino_InjectUnregisteredMultipleDependencyWithTags(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type ServiceB struct {
		Number int
	}

	type Consumer struct {
		A *ServiceA `inject:"serviceB"`
		B *ServiceB `inject:"serviceA"`
	}

	di := dino.New()
	consumer := new(Consumer)

	err := di.Inject(consumer)
	if err != nil {
		t.Fatalf("unexpected error during injection: %v", err)
	}

	if consumer.A == nil {
		t.Fatalf("expected ServiceA to be injected, got nil")
	}

	if consumer.A.Value != "" {
		t.Fatalf("expected ServiceA.Value to be empty, got '%s'", consumer.A.Value)
	}

	if consumer.B == nil {
		t.Fatalf("expected ServiceB to be injected, got nil")
	}

	if consumer.B.Number != 0 {
		t.Fatalf("expected ServiceB.Number to be 0, got %d", consumer.B.Number)
	}
}

func TestDino_InjectRegisteredSingleDependency(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type Consumer struct {
		A *ServiceA
	}

	di := dino.New()

	srvA := &ServiceA{
		Value: "injected value",
	}

	err := di.Singleton(srvA)
	if err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	consumer := new(Consumer)

	err = di.Inject(consumer)
	if err != nil {
		t.Fatalf("unexpected error during injection: %v", err)
	}

	if consumer.A == nil {
		t.Fatalf("expected ServiceA to be injected, got nil")
	}

	if consumer.A != srvA {
		t.Fatalf("expected ServiceA to be %v, got %v", srvA, consumer.A)
	}

	if consumer.A.Value != "injected value" {
		t.Fatalf("expected ServiceA.Value to be 'injected value', got '%s'", consumer.A.Value)
	}
}

func TestDino_InjectRegisteredSingleDependencyWithTag(t *testing.T) {
	t.Parallel()

	type Service struct {
		Value string
	}

	type Consumer struct {
		Srv *Service `inject:"tagged"`
	}

	di := dino.New()

	srvA := &Service{
		Value: "tagged injected value",
	}

	srvB := &Service{
		Value: "untagged injected value",
	}

	if err := di.Singleton(srvA, "tagged"); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	if err := di.Singleton(srvB); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	consumer := new(Consumer)

	if err := di.Inject(consumer); err != nil {
		t.Fatalf("unexpected error during injection: %v", err)
	}

	if consumer.Srv == nil {
		t.Fatalf("expected Service to be injected, got nil")
	}

	if consumer.Srv != srvA {
		t.Fatalf("expected Service to be %v, got %v", srvA, consumer.Srv)
	}

	if consumer.Srv.Value != "tagged injected value" {
		t.Fatalf(
			"expected Service.Value to be 'tagged injected value', got '%s'",
			consumer.Srv.Value,
		)
	}
}

func TestDino_InjectNestedDependency(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type ServiceB struct {
		A *ServiceA
	}

	type Consumer struct {
		B *ServiceB
	}

	di := dino.New()

	srvA := &ServiceA{
		Value: "nested injected value",
	}

	if err := di.Singleton(srvA); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	consumer := new(Consumer)

	if err := di.Inject(consumer); err != nil {
		t.Fatalf("unexpected error during injection: %v", err)
	}

	if consumer.B == nil {
		t.Fatalf("expected ServiceB to be injected, got nil")
	}

	if consumer.B.A == nil {
		t.Fatalf("expected ServiceA to be injected into ServiceB, got nil")
	}

	if consumer.B.A != srvA {
		t.Fatalf("expected ServiceA to be %v, got %v", srvA, consumer.B.A)
	}

	if consumer.B.A.Value != "nested injected value" {
		t.Fatalf(
			"expected ServiceA.Value to be 'nested injected value', got '%s'",
			consumer.B.A.Value,
		)
	}
}

func TestDino_InjectConcurrentAccess(t *testing.T) {
	t.Parallel()

	type Service struct {
		Number int
	}

	type Consumer struct {
		Srv *Service
	}

	di := dino.New()

	srv := &Service{
		Number: 999,
	}

	if err := di.Singleton(srv); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	consumerList := make([]*Consumer, 100)
	wg := sync.WaitGroup{}

	for idx := range 100 {
		wg.Go(func() {
			consumer := new(Consumer)

			if err := di.Inject(consumer); err != nil {
				t.Fatalf("unexpected error during injection: %v", err)
			}

			consumerList[idx] = consumer
		})
	}

	wg.Wait()

	for idx, consumer := range consumerList {
		if consumer.Srv == nil {
			t.Fatalf("expected Service to be injected in consumer %d, got nil", idx)
		}

		if consumer.Srv != srv {
			t.Fatalf("expected Service to be %v in consumer %d, got %v", srv, idx, consumer.Srv)
		}

		if consumer.Srv.Number != 999 {
			t.Fatalf(
				"expected Service.Number to be 999 in consumer %d, got %d",
				idx,
				consumer.Srv.Number,
			)
		}
	}
}

func TestDino_InvokeNilFunction(t *testing.T) {
	t.Parallel()

	di := dino.New()

	results, err := di.Invoke(nil)
	if !errors.Is(err, dino.ErrInvalidInputValue) {
		t.Fatalf("expected ErrInvalidInputValue, got %v", err)
	}

	if !strings.Contains(err.Error(), "function to invoke cannot be nil") {
		t.Fatalf(
			"expected error message to contain 'function to invoke cannot be nil', got %s",
			err.Error(),
		)
	}

	if results != nil {
		t.Fatalf("expected results to be nil, got %v", results)
	}
}

func TestDino_InvokeNotFunction(t *testing.T) {
	t.Parallel()

	di := dino.New()

	results, err := di.Invoke(42)
	if !errors.Is(err, dino.ErrInvalidInputValue) {
		t.Fatalf("expected ErrInvalidInputValue, got %v", err)
	}

	if !strings.Contains(err.Error(), "invoke expected a function") {
		t.Fatalf(
			"expected error message to contain 'invoke expected a function', got %s",
			err.Error(),
		)
	}

	if results != nil {
		t.Fatalf("expected results to be nil, got %v", results)
	}
}

func TestDino_InvokeError(t *testing.T) {
	t.Parallel()

	key := dino.RegistryKey{
		Tag:  "",
		Type: reflect.TypeFor[int](),
	}

	registry := new(dino.SyncMapRegistry)
	registry.MockRegister(key, "this is not a reflect.Value")

	di := dino.New().WithRegistry(registry)

	results, err := di.Invoke(func(i int) int { return i })
	if err == nil {
		t.Fatalf("expected error from Invoke, got nil")
	}

	if !strings.Contains(err.Error(), "failed to invoke function:") {
		t.Fatalf(
			"expected error message to contain 'failed to invoke function:', got %s",
			err.Error(),
		)
	}

	if results != nil {
		t.Fatalf("expected results to be nil, got %v", results)
	}
}

func TestDino_InvokeWithoutFunctionArgument(t *testing.T) {
	t.Parallel()

	di := dino.New()

	results, err := di.Invoke(func() int { return 51 })
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from Invoke, got %d", len(results))
	}

	result, ok := results[0].(int)
	if !ok {
		t.Fatalf("expected result to be of type int, got %T", results[0])
	}

	if result != 51 {
		t.Fatalf("expected result to be 51, got %d", result)
	}
}

func TestDino_InvokeWithSingleFunctionArgument(t *testing.T) {
	t.Parallel()

	type Service struct {
		Value string
	}

	srv := &Service{
		Value: "service value",
	}

	di := dino.New()

	if err := di.Singleton(srv); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	results, err := di.Invoke(func(s *Service) string {
		return s.Value
	})
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from Invoke, got %d", len(results))
	}

	result, ok := results[0].(string)
	if !ok {
		t.Fatalf("expected result to be of type string, got %T", results[0])
	}

	if result != "service value" {
		t.Fatalf("expected result to be 'service value', got '%s'", result)
	}
}

func TestDino_InvokeWithMultipleFunctionArguments(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type ServiceB struct {
		Number int
	}

	srvA := &ServiceA{
		Value: "service A value",
	}

	srvB := &ServiceB{
		Number: 123,
	}

	di := dino.New()

	if err := di.Singleton(srvA); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	if err := di.Singleton(srvB); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	results, err := di.Invoke(func(a *ServiceA, b *ServiceB) (string, int) {
		return a.Value, b.Number
	})
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results from Invoke, got %d", len(results))
	}

	resultA, ok := results[0].(string)
	if !ok {
		t.Fatalf("expected first result to be of type string, got %T", results[0])
	}

	resultB, ok := results[1].(int)
	if !ok {
		t.Fatalf("expected second result to be of type int, got %T", results[1])
	}

	if resultA != "service A value" {
		t.Fatalf("expected first result to be 'service A value', got '%s'", resultA)
	}

	if resultB != 123 {
		t.Fatalf("expected second result to be 123, got %d", resultB)
	}
}

func TestDino_InvokeWithFunctionReturnError(t *testing.T) {
	t.Parallel()

	type Service struct {
		Value string
	}

	srv := &Service{
		Value: "service value",
	}

	expectedErr := errors.New("function error")

	di := dino.New()

	if err := di.Singleton(srv); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	results, err := di.Invoke(func(s *Service) (string, error) {
		return s.Value, expectedErr
	})
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results from Invoke, got %d", len(results))
	}

	resultStr, ok := results[0].(string)
	if !ok {
		t.Fatalf("expected first result to be of type string, got %T", results[0])
	}

	resultErr, ok := results[1].(error)
	if !ok {
		t.Fatalf("expected second result to be of type error, got %T", results[1])
	}

	if resultStr != "service value" {
		t.Fatalf("expected first result to be 'service value', got '%s'", resultStr)
	}

	if !errors.Is(resultErr, expectedErr) {
		t.Fatalf("expected second result to be '%v', got '%v'", expectedErr, resultErr)
	}
}

func TestDino_InvokeWithFunctionReturnNil(t *testing.T) {
	t.Parallel()

	type Service struct {
		Value string
	}

	srv := &Service{
		Value: "service value",
	}

	di := dino.New()

	if err := di.Singleton(srv); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	results, err := di.Invoke(func(*Service) *string {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from Invoke, got %d", len(results))
	}

	result, ok := results[0].(*string)
	if !ok {
		t.Fatalf("expected result to be of type *string, got %T", results[0])
	}

	if result != nil {
		t.Fatalf("expected result to be nil, got %v", result)
	}
}

func TestDino_InvokeWithUnexportedDependency(t *testing.T) {
	t.Parallel()

	type privateType struct {
		value string
	}

	type Consumer struct {
		Dep *privateType
	}

	srv := &privateType{
		value: "unexported dependency value",
	}

	di := dino.New()

	if err := di.Singleton(srv); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	results, err := di.Invoke(func(c *Consumer) string {
		if c.Dep == nil {
			return "dependency is nil"
		}

		return c.Dep.value
	})
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from Invoke, got %d", len(results))
	}

	result, ok := results[0].(string)
	if !ok {
		t.Fatalf("expected result to be of type string, got %T", results[0])
	}

	if result != "unexported dependency value" {
		t.Fatalf("expected result to be 'unexported dependency value', got '%s'", result)
	}
}

func TestDino_InvokeWithUnregisteredDependency(t *testing.T) {
	t.Parallel()

	di := dino.New()

	results, err := di.Invoke(func(s string) string {
		return s
	})
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from Invoke, got %d", len(results))
	}

	result, ok := results[0].(string)
	if !ok {
		t.Fatalf("expected result to be of type string, got %T", results[0])
	}

	if result != "" {
		t.Fatalf("expected result to be empty string, got '%s'", result)
	}
}

func TestDino_InvokeWithNestedFunctionDependencies(t *testing.T) {
	t.Parallel()

	type ServiceA struct {
		Value string
	}

	type ServiceB struct {
		A *ServiceA
	}

	srvA := &ServiceA{
		Value: "nested service value",
	}

	di := dino.New()

	if err := di.Singleton(srvA); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	results, err := di.Invoke(func(b *ServiceB) string {
		if b.A == nil {
			return "ServiceA is nil"
		}

		return b.A.Value
	})
	if err != nil {
		t.Fatalf("unexpected error from Invoke: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from Invoke, got %d", len(results))
	}

	result, ok := results[0].(string)
	if !ok {
		t.Fatalf("expected result to be of type string, got %T", results[0])
	}

	if result != "nested service value" {
		t.Fatalf("expected result to be 'nested service value', got '%s'", result)
	}
}

func TestDino_InvokeConcurrentAccess(t *testing.T) {
	t.Parallel()

	type Service struct {
		Value string
	}

	srv := &Service{
		Value: "test",
	}

	di := dino.New()

	if err := di.Singleton(srv); err != nil {
		t.Fatalf("unexpected error during singleton registration: %v", err)
	}

	resultsList := make([][]any, 100)
	wg := sync.WaitGroup{}

	for idx := range 100 {
		wg.Go(func() {
			results, err := di.Invoke(func(s *Service) string {
				return fmt.Sprintf("%s_%d", s.Value, idx)
			})
			if err != nil {
				t.Fatalf("unexpected error from Invoke: %v", err)
			}

			resultsList[idx] = results
		})
	}

	wg.Wait()

	for idx, results := range resultsList {
		if len(results) != 1 {
			t.Fatalf("expected 1 result from Invoke in goroutine %d, got %d", idx, len(results))
		}

		result, ok := results[0].(string)
		if !ok {
			t.Fatalf(
				"expected result to be of type string in goroutine %d, got %T",
				idx,
				results[0],
			)
		}

		if result != fmt.Sprintf("test_%d", idx) {
			t.Fatalf("expected result to be 'test_%d' in goroutine %d, got '%s'", idx, idx, result)
		}
	}
}
