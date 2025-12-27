package dino_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yuppyweb/dino"
)

func TestInjector_WithEmptyRegistry(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type ServiceWithDeps struct {
		DB *DatabaseConnection
	}

	dbVal := &DatabaseConnection{
		Host: "localhost1",
	}

	target1 := new(ServiceWithDeps)
	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(dbVal), reflect.ValueOf(dbVal)); err != nil {
		t.Fatalf("failed to bind database connection: %v", err)
	}

	if err := injector.Inject(reflect.ValueOf(target1)); err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	if target1.DB == nil {
		t.Fatalf("expected database to be injected")
	}

	if target1.DB.Host != "localhost1" {
		t.Fatalf("expected host to be 'localhost', got '%s'", target1.DB.Host)
	}

	target2 := new(ServiceWithDeps)
	registry := new(dino.Registry)
	injector = injector.WithRegistry(registry)

	if err := injector.Inject(reflect.ValueOf(target2)); err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	if target2.DB == nil {
		t.Fatalf("expected database to be injected")
	}

	if target2.DB.Host == "localhost1" {
		t.Fatalf("expected host to not be 'localhost1', got '%s'", target2.DB.Host)
	}
}

func TestInjector_BindRegisterError(t *testing.T) {
	t.Parallel()

	injector := dino.NewInjector()

	err := injector.Bind(nil, reflect.ValueOf(42))
	if !errors.Is(err, dino.ErrKeyTypeNil) {
		t.Fatalf("expected ErrKeyTypeNil, got %v", err)
	}

	if !contains(err.Error(), "bind value to registry") {
		t.Fatalf("expected error message to contain 'bind value to registry', got '%s'", err.Error())
	}
}

func TestInjector_InjectSimpleFields(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type TargetStruct struct {
		DB *DatabaseConnection
	}

	dbVal := &DatabaseConnection{
		Host: "localhost2",
	}

	target := new(TargetStruct)
	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(dbVal), reflect.ValueOf(dbVal)); err != nil {
		t.Fatalf("failed to bind database connection: %v", err)
	}

	if err := injector.Inject(reflect.ValueOf(target)); err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	if target.DB == nil {
		t.Fatalf("expected database to be injected")
	}

	if target.DB.Host != "localhost2" {
		t.Fatalf("expected host to be 'localhost2', got '%s'", target.DB.Host)
	}
}

func TestInjector_InjectMultipleDependencies(t *testing.T) {
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
		Host: "localhost3",
	}

	logVal := &Logger{
		Level: "info",
	}

	target := new(ServiceWithDeps)
	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(dbVal), reflect.ValueOf(dbVal)); err != nil {
		t.Fatalf("failed to bind database connection: %v", err)
	}

	if err := injector.Bind(reflect.TypeOf(logVal), reflect.ValueOf(logVal)); err != nil {
		t.Fatalf("failed to bind logger: %v", err)
	}

	if err := injector.Inject(reflect.ValueOf(target)); err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	if target.DB == nil {
		t.Fatalf("expected database to be injected")
	}

	if target.DB.Host != "localhost3" {
		t.Fatalf("expected host to be 'localhost3', got '%s'", target.DB.Host)
	}

	if target.Log == nil {
		t.Fatalf("expected logger to be injected")
	}

	if target.Log.Level != "info" {
		t.Fatalf("expected level to be 'info', got '%s'", target.Log.Level)
	}
}

func TestInjector_InjectWithTag(t *testing.T) {
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

	replicaDB := &DatabaseConnection{
		Host: "replica-host",
	}

	target := new(ServiceWithTaggedDeps)
	enjector := dino.NewInjector()

	if err := enjector.Bind(reflect.TypeOf(primaryDB), reflect.ValueOf(primaryDB), "primary"); err != nil {
		t.Fatalf("failed to bind primary database: %v", err)
	}

	if err := enjector.Bind(reflect.TypeOf(replicaDB), reflect.ValueOf(replicaDB), "replica"); err != nil {
		t.Fatalf("failed to bind replica database: %v", err)
	}

	if err := enjector.Inject(reflect.ValueOf(target)); err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	if target.Primary == nil || target.Primary.Host != "primary-host" {
		t.Fatalf("expected primary database to be injected")
	}

	if target.Replica == nil || target.Replica.Host != "replica-host" {
		t.Fatalf("expected replica database to be injected")
	}
}

func TestInjector_InjectSkipsPrivateFields(t *testing.T) {
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

	target := new(StructWithPrivateField)
	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(srv), reflect.ValueOf(srv)); err != nil {
		t.Fatalf("failed to bind service: %v", err)
	}

	if err := injector.Inject(reflect.ValueOf(target)); err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	if target.Service == nil {
		t.Fatalf("expected public field to be injected")
	}

	if target.Service.Value != "test" {
		t.Fatalf("expected public field value to be 'test', got '%s'", target.Service.Value)
	}

	if target.private != nil {
		t.Fatalf("expected private field to remain nil")
	}
}

func TestInjector_InjectErrorResolveFields(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	type TargetStruct struct {
		Service *SimpleService
	}

	expectedErr := errors.New("service factory failed")

	srv := &SimpleService{
		Value: "error-case",
	}

	factory := func() (*SimpleService, error) {
		return nil, expectedErr
	}

	target := new(TargetStruct)
	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(srv), reflect.ValueOf(factory)); err != nil {
		t.Fatalf("failed to bind service factory: %v", err)
	}

	err := injector.Inject(reflect.ValueOf(target))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected factory error, got %v", err)
	}

	errMsg := "resolve field Service: factory function for type *dino_test.SimpleService with tag '' returned error:"

	if !contains(err.Error(), errMsg) {
		t.Fatalf("expected error message to contain '%s', got '%s'", errMsg, err.Error())
	}
}

func TestInjector_InjectErrorNestedInject(t *testing.T) {
	t.Parallel()

	type NestedService struct {
		Value string
	}

	type SimpleService struct {
		Service *NestedService
	}

	type TargetStruct struct {
		Service *SimpleService
	}

	expectedErr := errors.New("service factory failed")

	srv := &NestedService{
		Value: "error-case",
	}

	factory := func() (*NestedService, error) {
		return nil, expectedErr
	}

	target := new(TargetStruct)
	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(srv), reflect.ValueOf(factory)); err != nil {
		t.Fatalf("failed to bind service factory: %v", err)
	}

	err := injector.Inject(reflect.ValueOf(target))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected factory error, got %v", err)
	}

	errMsg := "inject field Service: resolve field Service: factory function " +
		"for type *dino_test.NestedService with tag '' returned error: service factory failed"

	if !contains(err.Error(), errMsg) {
		t.Fatalf("expected error message to contain '%s', got '%s'", errMsg, err.Error())
	}
}

func TestInjector_InvokeSimpleFunction(t *testing.T) {
	t.Parallel()

	executed := false
	fn := func() {
		executed = true
	}

	injector := dino.NewInjector()

	if err := injector.Invoke(reflect.ValueOf(fn)); err != nil {
		t.Fatalf("failed to invoke function: %v", err)
	}

	if !executed {
		t.Fatalf("expected function to be executed")
	}
}

func TestInjector_InvokeFunctionWithDependencies(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	dbVal := &DatabaseConnection{
		Host: "localhost6",
	}

	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(dbVal), reflect.ValueOf(dbVal)); err != nil {
		t.Fatalf("failed to bind database connection: %v", err)
	}

	var capturedDB *DatabaseConnection

	fn := func(db *DatabaseConnection) {
		capturedDB = db
	}

	err := injector.Invoke(reflect.ValueOf(fn))
	if err != nil {
		t.Fatalf("failed to invoke function: %v", err)
	}

	if capturedDB == nil {
		t.Fatalf("expected database to be passed to function")
	}

	if capturedDB.Host != "localhost6" {
		t.Fatalf("expected host to be 'localhost6', got '%s'", capturedDB.Host)
	}
}

func TestInjector_InvokeFunctionReturningError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("execution failed")

	fn := func() error {
		return expectedErr
	}

	injector := dino.NewInjector()

	err := injector.Invoke(reflect.ValueOf(fn))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected execution error, got %v", err)
	}

	if !contains(err.Error(), "function execution returned error:") {
		t.Fatalf(
			"expected error message to contain 'function execution returned error:', got '%s'",
			err.Error(),
		)
	}
}

func TestInjector_InvokeNotFunction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input any
		typ   string
	}{
		{
			name:  "Integer",
			input: 42,
			typ:   "int",
		},
		{
			name:  "String",
			input: "test",
			typ:   "string",
		},
		{
			name:  "Struct",
			input: struct{}{},
			typ:   "struct",
		},
		{
			name:  "Slice",
			input: []int{1, 2, 3},
			typ:   "slice",
		},
		{
			name:  "Map",
			input: map[string]int{"a": 1},
			typ:   "map",
		},
		{
			name:  "Channel",
			input: make(chan int),
			typ:   "chan",
		},
		{
			name:  "Pointer",
			input: new(int),
			typ:   "ptr",
		},
		{
			name:  "Boolean",
			input: true,
			typ:   "bool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			injector := dino.NewInjector()

			err := injector.Invoke(reflect.ValueOf(tc.input))

			if !errors.Is(err, dino.ErrExpectedFunction) {
				t.Fatalf("expected ErrExpectedFunction, got %v", err)
			}

			expectedMsg := "expected function: got " + tc.typ
			if !contains(err.Error(), expectedMsg) {
				t.Fatalf("expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
			}
		})
	}
}

func TestInjector_InvokeFunctionReturningMultipleValues(t *testing.T) {
	t.Parallel()

	executed := false

	fn := func() (int, string, error) {
		executed = true

		return 42, "test", nil
	}

	injector := dino.NewInjector()

	err := injector.Invoke(reflect.ValueOf(fn))
	if err != nil {
		t.Fatalf("failed to invoke function: %v", err)
	}

	if !executed {
		t.Fatalf("expected function to be executed")
	}
}

func TestInjector_ResolveSimpleFactory(t *testing.T) {
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
		Tag:  "",
		Type: reflect.TypeOf(srv),
	}

	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(srv), reflect.ValueOf(factory)); err != nil {
		t.Fatalf("failed to bind factory: %v", err)
	}

	val, err := injector.Resolve(key)
	if err != nil {
		t.Fatalf("failed to resolve factory: %v", err)
	}

	if !val.CanInterface() {
		t.Fatalf("expected value to be interfaceable")
	}

	service, ok := val.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected *SimpleService, got %T", val.Interface())
	}

	if service.Value != "existing" {
		t.Fatalf("expected value to be 'existing', got '%s'", service.Value)
	}
}

func TestInjector_ResolveFactoryWithDependencies(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type SimpleService struct {
		Value string
	}

	dbVal := &DatabaseConnection{
		Host: "localhost4",
	}

	factory := func(db *DatabaseConnection) *SimpleService {
		return &SimpleService{
			Value: db.Host,
		}
	}

	srvKey := dino.Key{
		Tag:  "",
		Type: reflect.TypeOf(new(SimpleService)),
	}

	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(dbVal), reflect.ValueOf(dbVal)); err != nil {
		t.Fatalf("failed to bind database connection: %v", err)
	}

	if err := injector.Bind(srvKey.Type, reflect.ValueOf(factory)); err != nil {
		t.Fatalf("failed to bind factory: %v", err)
	}

	val, err := injector.Resolve(srvKey)
	if err != nil {
		t.Fatalf("failed to resolve factory: %v", err)
	}

	if !val.CanInterface() {
		t.Fatalf("expected value to be interfaceable")
	}

	service, ok := val.Interface().(*SimpleService)
	if !ok {
		t.Fatalf("expected *SimpleService, got %T", val.Interface())
	}

	if service.Value != "localhost4" {
		t.Fatalf("expected value to be 'localhost4', got '%s'", service.Value)
	}
}

func TestInjector_ResolveFactoryReturningError(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	expectedErr := errors.New("factory failed")

	factory := func() (*SimpleService, error) {
		return nil, expectedErr
	}

	key := dino.Key{
		Tag:  "",
		Type: reflect.TypeOf(new(SimpleService)),
	}

	injector := dino.NewInjector()

	if err := injector.Bind(key.Type, reflect.ValueOf(factory)); err != nil {
		t.Fatalf("failed to bind factory: %v", err)
	}

	val, err := injector.Resolve(key)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected factory error, got %v", err)
	}

	if !contains(
		err.Error(),
		"factory function for type *dino_test.SimpleService with tag '' returned error:",
	) {
		t.Fatalf(
			"expected error message to contain 'factory function for type *dino_test.SimpleService with tag "+
				"'' returned error:', got '%s'",
			err.Error(),
		)
	}

	if val != reflect.Zero(key.Type) {
		t.Fatalf("expected returned value to be zero, got %v", val)
	}
}

func TestInjector_ResolveUnboundKey(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	key := dino.Key{
		Tag:  "missing",
		Type: reflect.TypeOf(new(SimpleService)),
	}

	injector := dino.NewInjector()

	val, err := injector.Resolve(key)
	if !errors.Is(err, dino.ErrValueNotFound) {
		t.Fatalf("expected ErrValueNotFound, got %v", err)
	}

	if val != reflect.Zero(key.Type) {
		t.Fatalf("expected returned value to be zero, got %v", val)
	}
}

func TestInjector_ResolveInvalidStoredValue(t *testing.T) {
	t.Parallel()

	type SimpleService struct {
		Value string
	}

	key := dino.Key{
		Tag:  "invalid",
		Type: reflect.TypeOf(new(SimpleService)),
	}

	registry := &dino.Registry{}
	registry.MockRegister(key, "this is not a reflect.Value")

	injector := dino.NewInjector()
	injector = injector.WithRegistry(registry)

	val, err := injector.Resolve(key)
	if !errors.Is(err, dino.ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}

	if val != reflect.Zero(key.Type) {
		t.Fatalf("expected returned value to be zero, got %v", val)
	}
}

func TestInjector_ResolveCircularDependency(t *testing.T) {
	t.Parallel()

	type ServiceB struct {
		Name string
	}

	type ServiceA struct {
		B *ServiceB
	}

	factoryA := func(b *ServiceB) *ServiceA {
		return &ServiceA{
			B: b,
		}
	}

	factoryB := func(a *ServiceA) *ServiceB {
		return &ServiceB{
			Name: "B",
		}
	}

	keyA := dino.Key{
		Tag:  "",
		Type: reflect.TypeOf(new(ServiceA)),
	}

	injector := dino.NewInjector()

	if err := injector.Bind(keyA.Type, reflect.ValueOf(factoryA)); err != nil {
		t.Fatalf("failed to bind factoryA: %v", err)
	}

	if err := injector.Bind(reflect.TypeOf(new(ServiceB)), reflect.ValueOf(factoryB)); err != nil {
		t.Fatalf("failed to bind factoryB: %v", err)
	}

	val, err := injector.Resolve(keyA)
	if !errors.Is(err, dino.ErrCircularDependency) {
		t.Fatalf("expected ErrCircularDependency, got %v", err)
	}

	if val != reflect.Zero(keyA.Type) {
		t.Fatalf("expected returned value to be zero, got %v", val)
	}
}

func TestInjector_PrepareArguments(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	type Logger struct {
		Level string
	}

	dbVal := &DatabaseConnection{
		Host: "localhost5",
	}

	logVal := &Logger{
		Level: "info",
	}

	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(dbVal), reflect.ValueOf(dbVal)); err != nil {
		t.Fatalf("failed to bind database connection: %v", err)
	}

	if err := injector.Bind(reflect.TypeOf(logVal), reflect.ValueOf(logVal)); err != nil {
		t.Fatalf("failed to bind logger: %v", err)
	}

	fn := func(*DatabaseConnection, *Logger) {}

	args, err := injector.Prepare(reflect.TypeOf(fn))
	if err != nil {
		t.Fatalf("failed to prepare arguments: %v", err)
	}

	if len(args) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(args))
	}

	if !args[0].CanInterface() {
		t.Fatalf("expected first arg to be interfaceable")
	}

	db, ok := args[0].Interface().(*DatabaseConnection)
	if !ok {
		t.Fatalf("expected first arg to be *DatabaseConnection")
	}

	if db.Host != "localhost5" {
		t.Fatalf("expected host to be 'localhost5', got '%s'", db.Host)
	}

	if !args[1].CanInterface() {
		t.Fatalf("expected second arg to be interfaceable")
	}

	log, ok := args[1].Interface().(*Logger)
	if !ok {
		t.Fatalf("expected second arg to be *Logger")
	}

	if log.Level != "info" {
		t.Fatalf("expected level to be 'info', got '%s'", log.Level)
	}
}

func TestInjector_PrepareArgumentsUnboundDependency(t *testing.T) {
	t.Parallel()

	type DatabaseConnection struct {
		Host string
	}

	fn := func(*DatabaseConnection) {}

	injector := dino.NewInjector()

	args, err := injector.Prepare(reflect.TypeOf(fn))
	if err != nil {
		t.Fatalf("failed to prepare arguments: %v", err)
	}

	if len(args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(args))
	}

	if args[0].IsNil() {
		t.Fatalf("expected argument to be non-nil")
	}
}

func TestInjector_PrepareNotFunction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input any
		typ   string
	}{
		{
			name:  "Integer",
			input: 42,
			typ:   "int",
		},
		{
			name:  "String",
			input: "test",
			typ:   "string",
		},
		{
			name:  "Struct",
			input: struct{}{},
			typ:   "struct",
		},
		{
			name:  "Slice",
			input: []int{1, 2, 3},
			typ:   "slice",
		},
		{
			name:  "Map",
			input: map[string]int{"a": 1},
			typ:   "map",
		},
		{
			name:  "Channel",
			input: make(chan int),
			typ:   "chan",
		},
		{
			name:  "Pointer",
			input: new(int),
			typ:   "ptr",
		},
		{
			name:  "Boolean",
			input: true,
			typ:   "bool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			injector := dino.NewInjector()

			_, err := injector.Prepare(reflect.TypeOf(tc.input))
			if !errors.Is(err, dino.ErrExpectedFunction) {
				t.Fatalf("expected ErrExpectedFunction, got %v", err)
			}

			expectedMsg := "expected function: got " + tc.typ
			if !contains(err.Error(), expectedMsg) {
				t.Fatalf("expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
			}
		})
	}
}

func TestInjector_PrepareArgumentsInjectError(t *testing.T) {
	t.Parallel()

	type NestedService struct {
		Value string
	}

	type SimpleService struct {
		Service *NestedService
	}

	fn := func(*SimpleService) {}

	expectedErr := errors.New("service factory failed")

	srv := &NestedService{
		Value: "error-case",
	}

	factory := func() (*NestedService, error) {
		return nil, expectedErr
	}

	injector := dino.NewInjector()

	if err := injector.Bind(reflect.TypeOf(srv), reflect.ValueOf(factory)); err != nil {
		t.Fatalf("failed to bind service factory: %v", err)
	}

	args, err := injector.Prepare(reflect.TypeOf(fn))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected factory error, got %v", err)
	}

	errMsg := "inject argument of type *dino_test.SimpleService: resolve field Service: " +
		"factory function for type *dino_test.NestedService with tag '' returned error: service factory failed"

	if !contains(err.Error(), errMsg) {
		t.Fatalf("expected error message to contain '%s', got '%s'", errMsg, err.Error())
	}

	if len(args) != 0 {
		t.Fatalf("expected 0 arguments on error, got %d", len(args))
	}
}

func TestInjector_CreateSlice(t *testing.T) {
	t.Parallel()

	injector := dino.NewInjector()

	rv := injector.Create(reflect.SliceOf(reflect.TypeOf(0)))

	if rv.Kind() != reflect.Slice {
		t.Fatalf("expected kind Slice, got %s", rv.Kind())
	}

	if rv.Len() != 0 {
		t.Fatalf("expected length 0, got %d", rv.Len())
	}

	if rv.Cap() != 0 {
		t.Fatalf("expected capacity 0, got %d", rv.Cap())
	}
}

func TestInjector_CreateMap(t *testing.T) {
	t.Parallel()

	injector := dino.NewInjector()

	rv := injector.Create(reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(0)))

	if rv.Kind() != reflect.Map {
		t.Fatalf("expected kind Map, got %s", rv.Kind())
	}

	if rv.Len() != 0 {
		t.Fatalf("expected length 0, got %d", rv.Len())
	}
}

func TestInjector_CreateChannel(t *testing.T) {
	t.Parallel()

	injector := dino.NewInjector()

	rv := injector.Create(reflect.ChanOf(reflect.BothDir, reflect.TypeOf(0)))

	if rv.Kind() != reflect.Chan {
		t.Fatalf("expected kind Chan, got %s", rv.Kind())
	}

	if rv.Cap() != 0 {
		t.Fatalf("expected capacity 0, got %d", rv.Cap())
	}
}

func TestInjector_CreatePointer(t *testing.T) {
	t.Parallel()

	injector := dino.NewInjector()

	rv := injector.Create(reflect.PointerTo(reflect.TypeOf(0)))

	if rv.Kind() != reflect.Pointer {
		t.Fatalf("expected kind Ptr, got %s", rv.Kind())
	}

	if rv.IsNil() {
		t.Fatalf("expected pointer to be non-nil")
	}
}

func TestInjector_CreateFunction(t *testing.T) {
	t.Parallel()

	injector := dino.NewInjector()

	rv := injector.Create(reflect.FuncOf(nil, nil, false))

	if rv.Kind() != reflect.Func {
		t.Fatalf("expected kind Func, got %s", rv.Kind())
	}
}

func TestInjector_CreateStruct(t *testing.T) {
	t.Parallel()

	injector := dino.NewInjector()

	rv := injector.Create(reflect.TypeOf(struct {
		Value string
	}{}))

	if rv.Kind() != reflect.Struct {
		t.Fatalf("expected kind Struct, got %s", rv.Kind())
	}

	field := rv.FieldByName("Value")
	if !field.IsValid() {
		t.Fatalf("expected field 'Value' to be valid")
	}

	if field.String() != "" {
		t.Fatalf("expected field 'Value' to be empty, got '%s'", field.String())
	}
}

func TestInjector_CreateAnyType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		typ  reflect.Type
	}{
		{
			name: "Integer",
			typ:  reflect.TypeOf(42),
		},
		{
			name: "String",
			typ:  reflect.TypeOf("test"),
		},
		{
			name: "Boolean",
			typ:  reflect.TypeOf(true),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			injector := dino.NewInjector()

			rv := injector.Create(tc.typ)

			if rv.Kind() != tc.typ.Kind() {
				t.Fatalf("expected kind %s, got %s", tc.typ.Kind(), rv.Kind())
			}
		})
	}
}

func contains(str, substr string) bool {
	for i := range len(str) - len(substr) + 1 {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
