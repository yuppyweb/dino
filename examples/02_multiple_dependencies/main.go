package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

// Define types
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

type Repository struct {
	DB     *Database
	Logger *Logger
}

func (r *Repository) Find(id string) string {
	r.Logger.Info(fmt.Sprintf("Finding record: %s", id))
	return fmt.Sprintf("Record from %s", r.DB.URL)
}

type Service struct {
	Repo   *Repository
	Logger *Logger
}

func (s *Service) GetData(id string) string {
	s.Logger.Info("Fetching data")
	return s.Repo.Find(id)
}

// Example demonstrating multiple dependencies and dependency chains
func main() {
	di := dino.New()

	// Register dependencies
	config := &Config{DatabaseURL: "postgresql://localhost:5432/mydb", Port: 8080}
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

	// Register Repository factory with automatic dependency resolution
	if err := di.Factory(func(db *Database, log *Logger) *Repository {
		return &Repository{DB: db, Logger: log}
	}); err != nil {
		log.Fatal(err)
	}

	// Register Service factory with automatic dependency resolution
	if err := di.Factory(func(repo *Repository, log *Logger) *Service {
		return &Service{Repo: repo, Logger: log}
	}); err != nil {
		log.Fatal(err)
	}

	// Get the service instance
	service, err := di.Invoke(func(svc *Service) *Service {
		return svc
	})
	if err != nil {
		log.Fatal(err)
	}

	// Use the service
	svcInstance := service[0].(*Service)
	result := svcInstance.GetData("123")
	fmt.Printf("Result: %s\n", result)
}
