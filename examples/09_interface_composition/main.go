package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

// Define interfaces.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	Save(key, value string) error
	Load(key string) (string, error)
}

// Implement interfaces.
type ConsoleLogger struct{}

func (l *ConsoleLogger) Info(msg string) {
	fmt.Printf("[INFO] %s\n", msg)
}

func (l *ConsoleLogger) Error(msg string) {
	fmt.Printf("[ERROR] %s\n", msg)
}

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string]string)}
}

func (s *MemoryStorage) Save(key, value string) error {
	s.data[key] = value

	return nil
}

func (s *MemoryStorage) Load(key string) (string, error) {
	if val, ok := s.data[key]; ok {
		return val, nil
	}

	return "", fmt.Errorf("key not found: %s", key)
}

// Define service.
type UserService struct {
	Logger  Logger
	Storage Storage
}

func (s *UserService) SaveUser(id, name string) error {
	s.Logger.Info("Saving user: " + name)

	return s.Storage.Save(id, name)
}

func (s *UserService) GetUser(id string) (string, error) {
	s.Logger.Info("Loading user: " + id)

	return s.Storage.Load(id)
}

// Example demonstrating interface-based composition for loose coupling.
func main() {
	di := dino.New()

	// Register implementations
	if err := di.Factory(func() Logger {
		return &ConsoleLogger{}
	}); err != nil {
		log.Fatal(err)
	}

	if err := di.Factory(func() Storage {
		return NewMemoryStorage()
	}); err != nil {
		log.Fatal(err)
	}

	// Register service factory with interface dependencies
	if err := di.Factory(func(logger Logger, storage Storage) *UserService {
		return &UserService{
			Logger:  logger,
			Storage: storage,
		}
	}); err != nil {
		log.Fatal(err)
	}

	// Use via Invoke
	_, err := di.Invoke(func(service *UserService) {
		fmt.Println("=== User Service Demo ===")

		if err := service.SaveUser("user-1", "Alice"); err != nil {
			log.Fatal(err)
		}

		if err := service.SaveUser("user-2", "Bob"); err != nil {
			log.Fatal(err)
		}

		fmt.Println("\n=== Loading Users ===")

		name, err := service.GetUser("user-1")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("User 1: %s\n", name)

		name, err = service.GetUser("user-2")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("User 2: %s\n", name)
	})
	if err != nil {
		log.Fatal(err)
	}
}
