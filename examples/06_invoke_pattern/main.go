package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

type Config struct {
	Port int
}

type Database struct {
	URL string
}

// Example demonstrating Invoke for function execution
func main() {
	di := dino.New()

	// Register dependencies
	config := &Config{Port: 8080}
	if err := di.Singleton(config); err != nil {
		log.Fatal(err)
	}

	db := &Database{URL: "postgresql://localhost:5432"}
	if err := di.Singleton(db); err != nil {
		log.Fatal(err)
	}

	// Define functions that use dependencies
	startServer := func(cfg *Config) string {
		return fmt.Sprintf("Server started on port %d", cfg.Port)
	}

	connectDB := func(db *Database) string {
		return fmt.Sprintf("Connected to %s", db.URL)
	}

	getStatus := func(cfg *Config, db *Database) string {
		return fmt.Sprintf("Server running on port %d with database %s", cfg.Port, db.URL)
	}

	// Invoke functions with automatic dependency resolution
	results1, err := di.Invoke(startServer)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result 1: %v\n", results1[0])

	results2, err := di.Invoke(connectDB)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result 2: %v\n", results2[0])

	results3, err := di.Invoke(getStatus)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result 3: %v\n", results3[0])
}
