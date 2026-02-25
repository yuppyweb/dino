package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

type Database struct {
	Name string
}

type App struct {
	PrimaryDB *Database `inject:"primary"`
	ReplicaDB *Database `inject:"replica"`
}

// Example demonstrating tagged dependencies for multiple implementations.
func main() {
	di := dino.New()

	// Register multiple implementations with tags
	if err := di.Factory(func() *Database {
		return &Database{Name: "primary-db"}
	}, "primary"); err != nil {
		log.Fatal(err)
	}

	if err := di.Factory(func() *Database {
		return &Database{Name: "replica-db"}
	}, "replica"); err != nil {
		log.Fatal(err)
	}

	// Create and inject
	app := &App{}
	if err := di.Inject(app); err != nil {
		log.Fatal(err)
	}

	// Use injected dependencies
	fmt.Printf("Primary Database: %s\n", app.PrimaryDB.Name)
	fmt.Printf("Replica Database: %s\n", app.ReplicaDB.Name)
}
