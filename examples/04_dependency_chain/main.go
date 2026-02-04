package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

type Config struct {
	MaxConnections int
}

type ConnectionPool struct {
	MaxConnections int
}

type Database struct {
	Pool *ConnectionPool
}

type Cache struct {
	Database *Database
}

type Repository struct {
	Cache *Cache
}

type Service struct {
	Repository *Repository
}

type Handler struct {
	Service *Service
}

// Example demonstrating complex dependency chain resolution
func main() {
	di := dino.New()

	// Register dependencies
	config := &Config{MaxConnections: 10}
	if err := di.Singleton(config); err != nil {
		log.Fatal(err)
	}

	// Register factory that uses dependencies
	if err := di.Factory(func(cfg *Config) *ConnectionPool {
		fmt.Printf("Creating ConnectionPool with max connections: %d\n", cfg.MaxConnections)
		return &ConnectionPool{MaxConnections: cfg.MaxConnections}
	}); err != nil {
		log.Fatal(err)
	}

	// Database factory with automatic dependency resolution
	if err := di.Factory(func(pool *ConnectionPool) *Database {
		fmt.Println("Creating Database with ConnectionPool")
		return &Database{Pool: pool}
	}); err != nil {
		log.Fatal(err)
	}

	// Cache factory with automatic dependency resolution
	if err := di.Factory(func(db *Database) *Cache {
		fmt.Println("Creating Cache with Database")
		return &Cache{Database: db}
	}); err != nil {
		log.Fatal(err)
	}

	// Repository factory with automatic dependency resolution
	if err := di.Factory(func(cache *Cache) *Repository {
		fmt.Println("Creating Repository with Cache")
		return &Repository{Cache: cache}
	}); err != nil {
		log.Fatal(err)
	}

	// Service factory with automatic dependency resolution
	if err := di.Factory(func(repo *Repository) *Service {
		fmt.Println("Creating Service with Repository")
		return &Service{Repository: repo}
	}); err != nil {
		log.Fatal(err)
	}

	// Handler factory with automatic dependency resolution
	if err := di.Factory(func(svc *Service) *Handler {
		fmt.Println("Creating Handler with Service")
		return &Handler{Service: svc}
	}); err != nil {
		log.Fatal(err)
	}

	// Get the handler instance
	results, err := di.Invoke(func(h *Handler) *Handler {
		return h
	})
	if err != nil {
		log.Fatal(err)
	}

	handler := results[0].(*Handler)

	// Verify injection chain
	fmt.Println("Dependency chain successfully resolved!")
	fmt.Printf("Handler -> Service -> Repository -> Cache -> Database -> ConnectionPool (max: %d)\n",
		handler.Service.Repository.Cache.Database.Pool.MaxConnections)
}
