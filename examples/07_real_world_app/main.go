package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

// Define domain models
type Config struct {
	DatabaseURL string
	Port        int
}

type Database struct {
	URL string
}

type Logger struct{}

func (l *Logger) Info(msg string) {
	fmt.Printf("[INFO] %s\n", msg)
}

type User struct {
	ID   string
	Name string
}

type UserRepository struct {
	DB     *Database
	Logger *Logger
}

func (r *UserRepository) GetUserByID(id string) *User {
	r.Logger.Info(fmt.Sprintf("Fetching user %s from %s", id, r.DB.URL))
	return &User{ID: id, Name: fmt.Sprintf("User-%s", id)}
}

type UserService struct {
	Repo   *UserRepository
	Logger *Logger
}

func (s *UserService) GetUser(id string) *User {
	s.Logger.Info(fmt.Sprintf("Service: GetUser(%s)", id))
	return s.Repo.GetUserByID(id)
}

type UserHandler struct {
	Service *UserService
	Logger  *Logger
}

func (h *UserHandler) HandleGetUser(id string) {
	h.Logger.Info(fmt.Sprintf("Handler: GetUser(%s)", id))
	user := h.Service.GetUser(id)
	fmt.Printf("Response: User{ID: %s, Name: %s}\n", user.ID, user.Name)
}

// Real-world example: building a web API with DI
func main() {
	di := dino.New()

	// Register dependencies
	config := &Config{
		DatabaseURL: "postgresql://localhost:5432/users",
		Port:        8080,
	}
	if err := di.Singleton(config); err != nil {
		log.Fatal(err)
	}

	db := &Database{URL: config.DatabaseURL}
	if err := di.Singleton(db); err != nil {
		log.Fatal(err)
	}

	logger := &Logger{}
	if err := di.Singleton(logger); err != nil {
		log.Fatal(err)
	}

	// Register UserRepository factory with automatic dependency resolution
	if err := di.Factory(func(db *Database, log *Logger) *UserRepository {
		return &UserRepository{DB: db, Logger: log}
	}); err != nil {
		log.Fatal(err)
	}

	// Register UserService factory with automatic dependency resolution
	if err := di.Factory(func(repo *UserRepository, log *Logger) *UserService {
		return &UserService{Repo: repo, Logger: log}
	}); err != nil {
		log.Fatal(err)
	}

	// Register UserHandler factory with automatic dependency resolution
	if err := di.Factory(func(svc *UserService, log *Logger) *UserHandler {
		return &UserHandler{Service: svc, Logger: log}
	}); err != nil {
		log.Fatal(err)
	}

	// Get the handler instance
	results, err := di.Invoke(func(h *UserHandler) *UserHandler {
		return h
	})
	if err != nil {
		log.Fatal(err)
	}

	handler := results[0].(*UserHandler)

	// Simulate API request
	fmt.Println("=== Simulating API Request ===")
	handler.HandleGetUser("user-123")
}
