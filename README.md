# 🦕 Dino

A lightweight and simple dependency injection container for Go. Dino provides a clean API for managing dependencies with support for factories, singletons, tagged dependencies, and automatic injection using struct tags.

## Features

- **Simple API**: Easy-to-use methods for registering and resolving dependencies
- **Factory Functions**: Register factory functions that create dependencies on demand
- **Singletons**: Register singleton instances that are created once and reused
- **Tagged Dependencies**: Support for multiple implementations of the same type using tags
- **Automatic Injection**: Inject dependencies into structs using the `inject` tag
- **Nested Injection**: Automatically resolves nested struct dependencies
- **Error Handling**: Propagates errors from factory functions
- **Type-Safe**: Leverages Go's type system for compile-time safety

## Installation

```bash
go get github.com/yuppyweb/dino
```

## Quick Start

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
    DB *Database `inject:""`
}

func main() {
    // Create a new container
    di := dino.New()
    
    // Register a singleton factory
    di.Singleton(func() *Database {
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

## Usage

### Registering Dependencies

#### Singleton

Register a factory function that creates a dependency. The result will be cached after the first call:

```go
di := dino.New()

// Simple singleton
di.Singleton(func() string {
    return "Hello, World!"
})

// Singleton with dependencies
di.Singleton(func() *Database {
    return &Database{ConnectionString: "localhost:5432"}
})

// Factory with error handling
di.Singleton(func() (*Service, error) {
    service, err := NewService()
    if err != nil {
        return nil, err
    }
    return service, nil
})
```

#### Tagged Dependencies

Use tags to register multiple implementations of the same type:

```go
di := dino.New()

// Register different database connections
di.Factory("primary", func() *Database {
    return &Database{ConnectionString: "primary:5432"}
})

di.Factory("replica", func() *Database {
    return &Database{ConnectionString: "replica:5432"}
})
```

### Injecting Dependencies

Use the `inject` struct tag to mark fields for automatic injection:

```go
type App struct {
    DB      *Database    `inject:""`
    Cache   *Cache       `inject:""`
    Primary *Database    `inject:"primary"`
    Replica *Database    `inject:"replica"`
}

app := &App{}
if err := di.Inject(app); err != nil {
    panic(err)
}
```

### Dependency Resolution

Dino automatically resolves dependencies for factory functions:

```go
di := dino.New()

// Register dependencies
di.Singleton(func() *Config {
    return &Config{Port: 8080}
})

di.Singleton(func() *Database {
    return &Database{}
})

// Factory with automatic dependency resolution
di.Singleton(func(cfg *Config, db *Database) *Server {
    return &Server{
        Port: cfg.Port,
        DB:   db,
    }
})
```

### Nested Structs

Dino automatically injects dependencies into nested structs:

```go
type Repository struct {
    DB *Database `inject:""`
}

type Service struct {
    Repo Repository `inject:""` // Nested struct
}

type Handler struct {
    Svc Service `inject:""` // Deeply nested
}

di := dino.New()
di.Singleton(func() *Database {
    return &Database{}
})

handler := &Handler{}
di.Inject(handler)
// handler.Svc.Repo.DB is now injected
```

## Complete Example

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

type UserRepository struct {
    DB *Database `inject:""`
}

type UserService struct {
    Repo *UserRepository `inject:""`
}

type Handler struct {
    Service *UserService `inject:""`
}

// Constructor functions
func NewConfig() *Config {
    return &Config{
        DatabaseURL: "localhost:5432",
        Port:        8080,
    }
}

func NewDatabase(cfg *Config) *Database {
    return &Database{URL: cfg.DatabaseURL}
}

func main() {
    // Create container
    di := dino.New()
    
    // Register dependencies
    if err := di.Singleton(NewConfig); err != nil {
        log.Fatal(err)
    }
    
    if err := di.Singleton(NewDatabase); err != nil {
        log.Fatal(err)
    }
    
    if err := di.Singleton(func() *UserRepository {
        return &UserRepository{}
    }); err != nil {
        log.Fatal(err)
    }
    
    if err := di.Singleton(func() *UserService {
        return &UserService{}
    }); err != nil {
        log.Fatal(err)
    }
    
    // Create and inject
    handler := &Handler{}
    if err := di.Inject(handler); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Handler initialized with DB: %s\n", 
        handler.Service.Repo.DB.URL)
}
```

## API Reference

### `New() *Container`

Creates a new dependency injection container.

### `Singleton(fn any) error`

Registers a factory function that creates a singleton dependency. The factory is called once, and the result is cached.

**Parameters:**
- `fn`: A factory function that returns one or more values. Can return an error as the last return value.

**Returns:**
- `error`: An error if the provided argument is not a function.

### `Factory(tag string, fn any) error`

Registers a factory function with a specific tag. Allows multiple implementations of the same type.

**Parameters:**
- `tag`: A string tag to identify this factory
- `fn`: A factory function

**Returns:**
- `error`: An error if the provided argument is not a function.

### `Inject(target any) error`

Injects dependencies into the target struct. Scans all fields with the `inject` tag and resolves their dependencies.

**Parameters:**
- `target`: A pointer to a struct

**Returns:**
- `error`: An error if dependency resolution fails or if a factory returns an error.

## Error Handling

Factory functions can return errors, which will be propagated through the injection process:

```go
di := dino.New()

di.Singleton(func() (*Database, error) {
    db, err := sql.Open("postgres", "...")
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %w", err)
    }
    return &Database{conn: db}, nil
})

app := &App{}
if err := di.Inject(app); err != nil {
    // Handle the error from the factory
    log.Fatal(err)
}
```

## Best Practices

1. **Use constructor functions**: Create factory functions that return initialized instances
2. **Handle errors**: Always check and handle errors from `Inject()`
3. **Tag strategically**: Use tags for multiple implementations of the same interface
4. **Keep it simple**: Don't over-complicate your dependency graph
5. **Export fields**: Only exported struct fields can be injected

## Testing

Run tests with coverage:

```bash
go test -v -cover
```

Generate coverage report:

```bash
go test -coverprofile=cover.out
go tool cover -html=cover.out
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
