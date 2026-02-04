# 🦕 Dino

A lightweight and simple dependency injection container for Go. Dino provides a clean API for managing dependencies with support for factories, singletons, tagged dependencies, and automatic injection using struct tags.

## ✨ Features

- 🎯 **Simple API**: Easy-to-use methods for registering and resolving dependencies
- 🏭 **Factory Functions**: Register factory functions that create dependencies on demand
- 📦 **Singletons**: Register singleton instances that are created once and reused
- 🏷️ **Tagged Dependencies**: Support for multiple implementations of the same type using tags
- 🔗 **Automatic Injection**: Inject dependencies into structs using the `inject` tag
- 🪆 **Nested Injection**: Automatically resolves nested struct dependencies
- ⚠️ **Error Handling**: Propagates errors from factory functions with context
- 🔒 **Type-Safe**: Leverages Go's type system for compile-time safety
- 🔄 **Circular Dependency Detection**: Catches circular dependencies before they cause issues
- ⚡ **Thread-Safe**: Safe concurrent access to the container

## Installation

```bash
go get github.com/yuppyweb/dino
```

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "github.com/yuppyweb/dino"
)

type Database struct {
    ConnectionString string
}

type UserService struct {
    DB *Database
}

func main() {
    // Create a new container
    di := dino.New()
    
    // Register a factory
    di.Factory(func() *Database {
        return &Database{
            ConnectionString: "localhost:5432",
        }
    })
    
    // Create and inject dependencies
    service := &UserService{}
    if err := di.Inject(service); err != nil {
        panic(err)
    }
    
    fmt.Println(service.DB.ConnectionString) // Output: localhost:5432
}
```

## 📚 Usage

### Registering Dependencies

#### Singleton - Register Once, Use Many Times 📌

Register an object instance that will be reused throughout the application:

```go
di := dino.New()

// Register a singleton instance
config := &Config{DatabaseURL: "localhost:5432"}
di.Singleton(config)

// Register another singleton
logger := &Logger{}
di.Singleton(logger)

// Singleton with factory function (Factory pattern)
di.Factory(func(cfg *Config) *Database {
    return &Database{URL: cfg.DatabaseURL}
})
```

#### Factory - Create New Instances on Demand 🏭

Factory caches the result after the first execution, so it's suitable for dependencies that need to be initialized once:

```go
di := dino.New()

// Factory caches the result after injection
di.Factory(func() *ConnectionPool {
    return &ConnectionPool{MaxConnections: 10}
})
```

#### Tagged Dependencies - Multiple Implementations 🏷️

Use tags to register multiple implementations of the same type:

```go
di := dino.New()

// Register different database connections
di.Factory(func() *Database {
    return &Database{ConnectionString: "primary:5432"}
}, "primary")

di.Factory(func() *Database {
    return &Database{ConnectionString: "replica:5432"}
}, "replica")

// Inject with tags
type App struct {
    PrimaryDB *Database `inject:"primary"`
    ReplicaDB *Database `inject:"replica"`
}

app := &App{}
di.Inject(app)
```

### Injecting Dependencies 💉

Use the `inject` struct tag to mark fields for automatic injection:

```go
type App struct {
    DB      *Database    
    Cache   *Cache       
    Primary *Database    `inject:"primary"`
    Replica *Database    `inject:"replica"`
}

app := &App{}
if err := di.Inject(app); err != nil {
    panic(err)
}
```

### Dependency Resolution 🔗

Dino automatically resolves dependencies for factory functions:

```go
di := dino.New()

// Register dependencies as instances
config := &Config{Port: 8080}
di.Singleton(config)

db := &Database{}
di.Singleton(db)

// Factory with automatic dependency resolution
di.Factory(func(cfg *Config, db *Database) *Server {
    return &Server{
        Port: cfg.Port,
        DB:   db,
    }
})
```

### Nested Structs 🪆

Dino automatically injects dependencies into nested structs:

```go
type Repository struct {
    DB *Database
}

type Service struct {
    Repo *Repository // Nested struct
}

type Handler struct {
    Svc *Service // Deeply nested
}

di := dino.New()

// Register database instance
db := &Database{}
di.Singleton(db)

// Register services
di.Singleton(&Repository{})
di.Singleton(&Service{})

handler := &Handler{}
di.Inject(handler)
// handler.Svc.Repo.DB is now injected
```

### Function Invocation 🎯

Automatically resolve and invoke functions with their dependencies:

```go
di := dino.New()

di.Factory(func() *Config {
    return &Config{Port: 8080}
})

// Invoke a function with automatic dependency resolution
results, err := di.Invoke(func(cfg *Config) string {
    return fmt.Sprintf("Server running on port %d", cfg.Port)
})

if err != nil {
    panic(err)
}

fmt.Println(results[0]) // Output: Server running on port 8080
```

## 📖 Complete Example

Here's a real-world example with multiple services:

```go
package main

import (
    "fmt"
    "log"
    "github.com/yuppyweb/dino"
)

// Models
type Config struct {
    DatabaseURL string
    Port        int
}

type Database struct {
    URL string
}

type Logger struct{}

func (l *Logger) Info(msg string) {
    log.Println("[INFO]", msg)
}

type UserRepository struct {
    DB     *Database
    Logger *Logger
}

func (ur *UserRepository) GetUser(id string) string {
    ur.Logger.Info("Fetching user: " + id)
    return "User from " + ur.DB.URL
}

type UserService struct {
    Repo   *UserRepository
    Logger *Logger
}

func (us *UserService) FetchUser(id string) string {
    us.Logger.Info("UserService.FetchUser called")
    return us.Repo.GetUser(id)
}

type Handler struct {
    Service *UserService
    Logger  *Logger
}

func (h *Handler) HandleRequest(userID string) {
    h.Logger.Info("Handling request for user: " + userID)
    result := h.Service.FetchUser(userID)
    fmt.Println(result)
}

func main() {
    // Create container
    di := dino.New()
    
    // Register dependencies
    config := &Config{
        DatabaseURL: "postgres://localhost:5432/mydb",
        Port:        8080,
    }
    if err := di.Singleton(config); err != nil {
        log.Fatal(err)
    }
    
    database := &Database{URL: config.DatabaseURL}
    if err := di.Singleton(database); err != nil {
        log.Fatal(err)
    }
    
    logger := &Logger{}
    if err := di.Singleton(logger); err != nil {
        log.Fatal(err)
    }
    
    userRepo := &UserRepository{}
    if err := di.Singleton(userRepo); err != nil {
        log.Fatal(err)
    }
    
    userService := &UserService{}
    if err := di.Singleton(userService); err != nil {
        log.Fatal(err)
    }
    
    // Create and inject handler
    handler := &Handler{}
    if err := di.Inject(handler); err != nil {
        log.Fatal(err)
    }
    
    // Use the handler
    handler.HandleRequest("user-123")
}
```

## 🔍 API Reference

### `New() *Dino`

Creates a new dependency injection container.

```go
di := dino.New()
```

### `Singleton(val any, tags ...string) error`

Registers an object instance as a singleton dependency. The instance will be reused throughout the application.

**Parameters:**
- `val`: An object instance to register
- `tags`: Optional tags to identify this dependency

**Returns:**
- `error`: An error if the provided argument is invalid

**Example:**
```go
config := &Config{DatabaseURL: "localhost"}
di.Singleton(config, "mysql")

logger := &Logger{}
di.Singleton(logger)
```

### `Factory(fn any, tags ...string) error`

Registers a factory function with optional tags. Allows multiple implementations of the same type.

**Parameters:**
- `fn`: A factory function
- `tags`: Optional tags to identify this factory

**Returns:**
- `error`: An error if the provided argument is not a function

**Example:**
```go
di.Factory(func() *Database {
    return &Database{conn: "read-replica"}
}, "read")

di.Factory(func() *Database {
    return &Database{conn: "primary"}
}, "write")
```

### `Inject(target any) error`

Injects dependencies into the target struct. Scans all fields and resolves their dependencies.

**Parameters:**
- `target`: A pointer to a struct

**Returns:**
- `error`: An error if dependency resolution fails or if a factory returns an error

**Example:**
```go
type Service struct {
    DB *Database
}

svc := &Service{}
if err := di.Inject(svc); err != nil {
    log.Fatal(err)
}
```

### `Invoke(fn any) ([]any, error)`

Automatically resolves and invokes a function with its dependencies.

**Parameters:**
- `fn`: A function to invoke

**Returns:**
- `[]any`: Slice containing the function's return values
- `error`: An error if dependency resolution fails

**Example:**
```go
results, err := di.Invoke(func(db *Database) string {
    return "Connected to " + db.URL
})
```

### `WithRegistry(registry Registry) *Dino`

Sets a custom registry implementation (advanced usage).

**Parameters:**
- `registry`: Custom Registry implementation

**Returns:**
- `*Dino`: The container instance for chaining

## ⚠️ Error Handling from Factories

When a factory function returns an error, that error is immediately returned by the Resolve method. This ensures:

1. **Fail-Fast**: Errors are caught immediately, not silently ignored
2. **Consistency**: No partial initialization (all-or-nothing semantics)
3. **Transparency**: Original error is preserved and wrapped with context

**Example:**
```go
func NewDatabase(cfg *Config) (*Database, error) {
    if cfg == nil {
        return nil, fmt.Errorf("config required")
    }
    return &Database{...}, nil
}

di.Factory(NewDatabase)

// If NewDatabase returns an error, Inject will propagate it:
svc := &Service{}
if err := di.Inject(svc); err != nil {
    // err will contain the factory's error with context
    log.Fatal(err)
}
```

## 🔄 Circular Dependency Detection

Dino detects and prevents circular dependencies:

```go
// This will be caught and return an error
di.Factory(func(svc *ServiceA) *ServiceB {
    return &ServiceB{SvcA: svc}
})

di.Factory(func(svc *ServiceB) *ServiceA {
    return &ServiceA{SvcB: svc}
})

// Error: circular dependency detected
```

## 💡 Best Practices

1. **Use Factory for initialization**: Use `Factory()` to register functions that initialize and return instances
2. **Use Singleton for instances**: Use `Singleton()` to register already created object instances
3. **Handle errors**: Always check and handle errors from `Inject()`
4. **Tag strategically**: Use tags for multiple implementations of the same interface
5. **Keep it simple**: Don't over-complicate your dependency graph
6. **Export fields**: Only exported struct fields can be injected
7. **Avoid circular dependencies**: Design your dependency graph as a DAG (Directed Acyclic Graph)
8. **Use interfaces**: Define dependencies as interfaces for better testability

**Example with Factory and Interface:**
```go
type Logger interface {
    Info(msg string)
    Error(msg string)
}

type ConsoleLogger struct{}

func (cl *ConsoleLogger) Info(msg string) {
    log.Println("[INFO]", msg)
}

func (cl *ConsoleLogger) Error(msg string) {
    log.Println("[ERROR]", msg)
}

type Service struct {
    Logger Logger
}

di := dino.New()

// Register logger factory
di.Factory(func() Logger {
    return &ConsoleLogger{}
})

// Register service factory
di.Factory(func(logger Logger) *Service {
    return &Service{Logger: logger}
})
```

## 🔗 Examples

Comprehensive examples are available in the [examples directory](./dino/examples/). To run examples:

```bash
go run dino/examples/01_basic_usage.go
go run dino/examples/02_multiple_dependencies.go
go run dino/examples/03_tagged_dependencies.go
```

Available examples:
- **01_basic_usage.go** - Basic DI setup
- **02_multiple_dependencies.go** - Multiple services and dependency chains
- **03_tagged_dependencies.go** - Using tags for multiple instances
- **04_dependency_chain.go** - Complex dependency resolution
- **05_factory_vs_set.go** - Factory vs Singleton patterns
- **06_execute_pattern.go** - Function invocation with `Invoke()`
- **07_real_world_app.go** - Real-world application example
- **08_lifecycle_management.go** - Managing component lifecycle
- **09_interface_composition.go** - Using interfaces for loose coupling

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

For more information about the MIT License, visit [opensource.org/licenses/MIT](https://opensource.org/licenses/MIT).
