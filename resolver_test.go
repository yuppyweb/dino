package dino_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yuppyweb/dino"
)

// Helper types for testing
type SimpleService struct {
	Value string
}

type DatabaseConnection struct {
	Host string
}

type Logger struct {
	Level string
}

type ServiceWithDeps struct {
	DB  *DatabaseConnection
	Log *Logger
}

func TestResolver_InjectSimpleFields(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type TargetStruct struct {
		DB *DatabaseConnection
	}

	dbVal := &DatabaseConnection{
		Host: "localhost",
	}

	dbKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(dbVal),
	}

	target := new(TargetStruct)
	registry := new(dino.Registry)
	registry.Add(dbKey, reflect.ValueOf(dbVal))

	resolver := dino.NewResolver(registry)
	resolver.Inject(reflect.ValueOf(target))

	if target.DB == nil {
		t.Fatalf("expected database to be injected")
	}

	if target.DB.Host != "localhost" {
		t.Fatalf("expected host to be 'localhost', got %s", target.DB.Host)
	}

	if len(resolver.Unwrap()) != 0 {
		t.Fatalf("expected no errors, got %d", len(resolver.Unwrap()))
	}
}

func TestResolver_InjectMultipleDependencies(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type Logger struct {
		Level string
	}

	type ServiceWithDeps struct {
		DB  *DatabaseConnection
		Log *Logger
	}

	dbVal := &DatabaseConnection{
		Host: "localhost",
	}

	dbKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(dbVal),
	}

	logVal := &Logger{
		Level: "info",
	}

	logKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(logVal),
	}

	target := new(ServiceWithDeps)
	registry := new(dino.Registry)
	registry.Add(dbKey, reflect.ValueOf(dbVal))
	registry.Add(logKey, reflect.ValueOf(logVal))

	resolver := dino.NewResolver(registry)
	resolver.Inject(reflect.ValueOf(target))

	if target.DB == nil {
		t.Fatalf("expected database to be injected")
	}

	if target.DB.Host != "localhost" {
		t.Fatalf("expected host to be 'localhost', got %s", target.DB.Host)
	}

	if target.Log == nil {
		t.Fatalf("expected logger to be injected")
	}

	if target.Log.Level != "info" {
		t.Fatalf("expected level to be 'info', got %s", target.Log.Level)
	}

	if len(resolver.Unwrap()) != 0 {
		t.Fatalf("expected no errors, got %d", len(resolver.Unwrap()))
	}
}

func TestResolver_InjectWithTag(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type ServiceWithTaggedDeps struct {
		Primary *DatabaseConnection `inject:"primary"`
		Replica *DatabaseConnection `inject:"replica"`
	}

	primaryDB := &DatabaseConnection{
		Host: "primary-host",
	}

	key1 := dino.Key{
		Tag: "primary",
		Ref: reflect.TypeOf(primaryDB),
	}

	replicaDB := &DatabaseConnection{
		Host: "replica-host",
	}

	key2 := dino.Key{
		Tag: "replica",
		Ref: reflect.TypeOf(replicaDB),
	}

	target := new(ServiceWithTaggedDeps)
	registry := new(dino.Registry)
	registry.Add(key1, reflect.ValueOf(primaryDB))
	registry.Add(key2, reflect.ValueOf(replicaDB))

	resolver := dino.NewResolver(registry)
	resolver.Inject(reflect.ValueOf(target))

	if target.Primary == nil || target.Primary.Host != "primary-host" {
		t.Fatalf("expected primary database to be injected")
	}

	if target.Replica == nil || target.Replica.Host != "replica-host" {
		t.Fatalf("expected replica database to be injected")
	}

	if len(resolver.Unwrap()) != 0 {
		t.Fatalf("expected no errors, got %d", len(resolver.Unwrap()))
	}
}

func TestResolver_InjectSkipsPrivateFields(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	type StructWithPrivateField struct {
		Service *SimpleService
		private *SimpleService
	}

	srv := &SimpleService{
		Value: "test",
	}

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(srv),
	}

	target := new(StructWithPrivateField)
	registry := new(dino.Registry)
	registry.Add(key, reflect.ValueOf(srv))

	resolver := dino.NewResolver(registry)
	resolver.Inject(reflect.ValueOf(target))

	if target.Service == nil {
		t.Fatalf("expected public field to be injected")
	}

	if target.private != nil {
		t.Fatalf("expected private field to remain nil")
	}

	if len(resolver.Unwrap()) != 0 {
		t.Fatalf("expected no errors, got %d", len(resolver.Unwrap()))
	}
}

func TestResolver_InjectIntoNonStruct(t *testing.T) {
	t.Parallel()

	value1 := 10
	value2 := 20

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(value1),
	}

	registry := new(dino.Registry)
	registry.Add(key, reflect.ValueOf(value2))

	resolver := dino.NewResolver(registry)
	resolver.Inject(reflect.ValueOf(value1))

	if value1 != 10 {
		t.Fatalf("expected value to remain 10")
	}

	if len(resolver.Unwrap()) != 0 {
		t.Fatalf("expected no errors, got %d", len(resolver.Unwrap()))
	}
}

func TestResolver_ResolveSimpleFactory(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	srv := &SimpleService{
		Value: "existing",
	}

	factory := func() *SimpleService {
		return srv
	}

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(srv),
	}

	registry := new(dino.Registry)
	registry.Add(key, reflect.ValueOf(factory))

	resolver := dino.NewResolver(registry)

	val, ok := resolver.Resolve(key)
	if !ok {
		t.Fatalf("expected factory to be resolved")
	}

	if !val.CanInterface() {
		t.Fatalf("expected value to be interfaceable")
	}

	service, ok := val.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected *SimpleService, got %T", val.Interface())
	}

	if service.Value != "existing" {
		t.Fatalf("expected value to be 'existing', got %s", service.Value)
	}

	if len(resolver.Unwrap()) != 0 {
		t.Fatalf("expected no errors, got %d", len(resolver.Unwrap()))
	}
}

func TestResolver_ResolveFactoryWithDependencies(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type SimpleService struct {
		Value string
	}

	dbVal := &DatabaseConnection{
		Host: "localhost",
	}

	dbKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(dbVal),
	}

	factory := func(db *DatabaseConnection) *SimpleService {
		return &SimpleService{
			Value: db.Host,
		}
	}

	srvKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(new(SimpleService)),
	}

	registry := new(dino.Registry)
	registry.Add(dbKey, reflect.ValueOf(dbVal))
	registry.Add(srvKey, reflect.ValueOf(factory))

	resolver := dino.NewResolver(registry)

	val, ok := resolver.Resolve(srvKey)
	if !ok {
		t.Fatalf("expected factory to be resolved")
	}

	if !val.CanInterface() {
		t.Fatalf("expected value to be interfaceable")
	}

	service, ok := val.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected *SimpleService, got %T", val.Interface())
	}

	if service.Value != "localhost" {
		t.Fatalf("expected value to be 'localhost', got %s", service.Value)
	}

	if len(resolver.Unwrap()) != 0 {
		t.Fatalf("expected no errors, got %d", len(resolver.Unwrap()))
	}
}

func TestResolver_ResolveFactoryReturningError(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	expectedErr := errors.New("factory failed")

	factory := func() (*SimpleService, error) {
		return nil, expectedErr
	}

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(new(SimpleService)),
	}

	registry := new(dino.Registry)
	registry.Add(key, reflect.ValueOf(factory))

	resolver := dino.NewResolver(registry)

	val, ok := resolver.Resolve(key)
	if ok {
		t.Fatalf("expected factory to fail")
	}

	if val != reflect.ValueOf(factory) {
		t.Fatalf("expected returned value to be zero, got %v", val)
	}

	if len(resolver.Unwrap()) == 0 {
		t.Fatalf("expected error to be recorded")
	}

	if !errors.Is(resolver, expectedErr) {
		t.Fatalf("expected error message 'factory failed', got %s", expectedErr.Error())
	}
}

// TestResolver_PrepareArguments tests argument preparation for function calls
func TestResolver_PrepareArguments(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}

	dbKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&DatabaseConnection{}),
	}
	registry.Add(dbKey, reflect.ValueOf(&DatabaseConnection{Host: "localhost"}))

	logKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&Logger{}),
	}
	registry.Add(logKey, reflect.ValueOf(&Logger{Level: "info"}))

	resolver := dino.NewResolver(registry)

	fn := func(db *DatabaseConnection, log *Logger) {
		// dummy function
	}

	args := resolver.Prepare(reflect.TypeOf(fn))

	if len(args) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(args))
	}

	db, ok := args[0].Interface().(*DatabaseConnection)
	if !ok {
		t.Fatalf("expected first arg to be *DatabaseConnection")
	}

	if db.Host != "localhost" {
		t.Fatalf("expected host to be 'localhost', got %s", db.Host)
	}

	log, ok := args[1].Interface().(*Logger)
	if !ok {
		t.Fatalf("expected second arg to be *Logger")
	}

	if log.Level != "info" {
		t.Fatalf("expected level to be 'info', got %s", log.Level)
	}
}

// TestResolver_ExecuteSimpleFunction tests executing a simple function
func TestResolver_ExecuteSimpleFunction(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}
	resolver := dino.NewResolver(registry)

	executed := false
	fn := func() {
		executed = true
	}

	resolver.Execute(reflect.ValueOf(fn))

	if !executed {
		t.Fatalf("expected function to be executed")
	}
}

// TestResolver_ExecuteFunctionWithDependencies tests executing function with dependencies
func TestResolver_ExecuteFunctionWithDependencies(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}

	dbKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&DatabaseConnection{}),
	}
	registry.Add(dbKey, reflect.ValueOf(&DatabaseConnection{Host: "localhost"}))

	resolver := dino.NewResolver(registry)

	var capturedDB *DatabaseConnection
	fn := func(db *DatabaseConnection) {
		capturedDB = db
	}

	resolver.Execute(reflect.ValueOf(fn))

	if capturedDB == nil {
		t.Fatalf("expected database to be passed to function")
	}

	if capturedDB.Host != "localhost" {
		t.Fatalf("expected host to be 'localhost', got %s", capturedDB.Host)
	}
}

// TestResolver_ExecuteFunctionReturningError tests executing function that returns error
func TestResolver_ExecuteFunctionReturningError(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}
	resolver := dino.NewResolver(registry)

	fn := func() error {
		return errors.New("execution failed")
	}

	resolver.Execute(reflect.ValueOf(fn))

	if len(resolver.Unwrap()) == 0 {
		t.Fatalf("expected error to be recorded")
	}

	if resolver.Unwrap()[0].Error() != "execution failed" {
		t.Fatalf("expected error message 'execution failed'")
	}
}

// TestResolver_ExecuteFunctionReturningValue tests executing function that returns value
func TestResolver_ExecuteFunctionReturningValue(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}
	resolver := dino.NewResolver(registry)

	fn := func() *SimpleService {
		return &SimpleService{Value: "created"}
	}

	resolver.Execute(reflect.ValueOf(fn))

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&SimpleService{}),
	}

	val, ok := registry.Get(key)
	if !ok {
		t.Fatalf("expected result to be stored in registry")
	}

	service, ok := val.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected *SimpleService, got %T", val.Interface())
	}

	if service.Value != "created" {
		t.Fatalf("expected value to be 'created', got %s", service.Value)
	}
}

// TestResolver_ExecuteFunctionReturningMultipleValues tests function with multiple return values
func TestResolver_ExecuteFunctionReturningMultipleValues(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}
	resolver := dino.NewResolver(registry)

	fn := func() (*SimpleService, *DatabaseConnection, error) {
		return &SimpleService{Value: "service"}, &DatabaseConnection{Host: "db"}, nil
	}

	resolver.Execute(reflect.ValueOf(fn))

	// Check SimpleService was stored
	serviceKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&SimpleService{}),
	}
	serviceVal, ok := registry.Get(serviceKey)
	if !ok {
		t.Fatalf("expected SimpleService to be stored")
	}

	if service, ok := serviceVal.Interface().(*SimpleService); !ok || service.Value != "service" {
		t.Fatalf("expected correct SimpleService")
	}

	// Check DatabaseConnection was stored
	dbKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&DatabaseConnection{}),
	}
	dbVal, ok := registry.Get(dbKey)
	if !ok {
		t.Fatalf("expected DatabaseConnection to be stored")
	}

	if db, ok := dbVal.Interface().(*DatabaseConnection); !ok || db.Host != "db" {
		t.Fatalf("expected correct DatabaseConnection")
	}
}

// TestResolver_ResolveKeyNotFound tests resolving non-existent key
func TestResolver_ResolveKeyNotFound(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}
	resolver := dino.NewResolver(registry)

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&SimpleService{}),
	}

	_, ok := resolver.Resolve(key)

	if ok {
		t.Fatalf("expected resolve to return false for missing key")
	}
}

// TestResolver_InjectIntoNestedStructs tests injection into nested structures
func TestResolver_InjectIntoNestedStructs(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}

	logKey := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&Logger{}),
	}
	registry.Add(logKey, reflect.ValueOf(&Logger{Level: "debug"}))

	resolver := dino.NewResolver(registry)

	type InnerService struct {
		Log *Logger
	}

	type OuterService struct {
		Inner *InnerService
	}

	target := &OuterService{}
	resolver.Inject(reflect.ValueOf(target))

	if target.Inner == nil {
		t.Fatalf("expected inner service to be created")
	}

	if target.Inner.Log == nil {
		t.Fatalf("expected logger to be injected into inner service")
	}

	if target.Inner.Log.Level != "debug" {
		t.Fatalf("expected level to be 'debug', got %s", target.Inner.Log.Level)
	}
}

// TestResolver_ErrorCollection tests that multiple errors are collected
func TestResolver_ErrorCollection(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}

	factory1 := func() (*SimpleService, error) {
		return nil, errors.New("error 1")
	}

	factory2 := func() (*DatabaseConnection, error) {
		return nil, errors.New("error 2")
	}

	key1 := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&SimpleService{}),
	}
	registry.Add(key1, reflect.ValueOf(factory1))

	key2 := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&DatabaseConnection{}),
	}
	registry.Add(key2, reflect.ValueOf(factory2))

	resolver := dino.NewResolver(registry)

	resolver.Resolve(key1)
	resolver.Resolve(key2)

	errors := resolver.Unwrap()
	if len(errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errors))
	}

	if errors[0].Error() != "error 1" {
		t.Fatalf("expected first error to be 'error 1', got %s", errors[0].Error())
	}

	if errors[1].Error() != "error 2" {
		t.Fatalf("expected second error to be 'error 2', got %s", errors[1].Error())
	}
}

// TestResolver_ErrorMethod tests Error() method
func TestResolver_ErrorMethod(t *testing.T) {
	t.Parallel()

	registry := &dino.Registry{}

	factory := func() (*SimpleService, error) {
		return nil, errors.New("test error")
	}

	key := dino.Key{
		Tag: "",
		Ref: reflect.TypeOf(&SimpleService{}),
	}
	registry.Add(key, reflect.ValueOf(factory))

	resolver := dino.NewResolver(registry)
	resolver.Resolve(key)

	errorMsg := resolver.Error()
	if !contains(errorMsg, "test error") {
		t.Fatalf("expected error message to contain 'test error', got %s", errorMsg)
	}
}

// Helper function
func contains(str, substr string) bool {
	for i := 0; i < len(str)-len(substr)+1; i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
