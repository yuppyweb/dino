package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

// Define a simple service
type Database struct {
	ConnectionString string
}

type UserService struct {
	DB *Database
}

// Basic example demonstrating simple dependency registration and injection
func main() {
	// Create a new DI container
	di := dino.New()

	// Register a singleton instance
	db := &Database{ConnectionString: "localhost:5432"}
	if err := di.Singleton(db); err != nil {
		log.Fatal(err)
	}

	// Create and inject dependencies
	service := &UserService{}
	if err := di.Inject(service); err != nil {
		log.Fatal(err)
	}

	// Use the injected service
	fmt.Printf("Service connected to: %s\n", service.DB.ConnectionString)
}
