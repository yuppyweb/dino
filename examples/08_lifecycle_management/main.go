package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/yuppyweb/dino"
)

type Database struct {
	Name string
	mu   sync.Mutex
	open bool
}

func (d *Database) Init() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.open = true
	fmt.Printf("[%s] Initialized\n", d.Name)
	return nil
}

func (d *Database) Shutdown() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.open {
		d.open = false
		fmt.Printf("[%s] Shutdown\n", d.Name)
	}
	return nil
}

type Cache struct {
	Name string
	mu   sync.Mutex
	open bool
}

func (c *Cache) Init() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.open = true
	fmt.Printf("[%s] Initialized\n", c.Name)
	return nil
}

func (c *Cache) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.open {
		c.open = false
		fmt.Printf("[%s] Shutdown\n", c.Name)
	}
	return nil
}

type Application struct {
	DB    *Database
	Cache *Cache
}

// Example demonstrating lifecycle management with shutdown hooks
func main() {
	di := dino.New()

	// Initialize resources
	db := &Database{Name: "PostgreSQL"}
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	if err := di.Singleton(db); err != nil {
		log.Fatal(err)
	}

	cache := &Cache{Name: "Redis"}
	if err := cache.Init(); err != nil {
		log.Fatal(err)
	}
	if err := di.Singleton(cache); err != nil {
		log.Fatal(err)
	}

	// Create application
	app := &Application{}
	if err := di.Inject(app); err != nil {
		log.Fatal(err)
	}

	// Use application
	fmt.Println("=== Application Running ===")
	fmt.Printf("Database: %s (open: %v)\n", app.DB.Name, app.DB.open)
	fmt.Printf("Cache: %s (open: %v)\n", app.Cache.Name, app.Cache.open)

	// Shutdown resources (simulate graceful shutdown)
	fmt.Println("\n=== Shutting Down ===")
	if err := app.Cache.Shutdown(); err != nil {
		log.Fatal(err)
	}
	if err := app.DB.Shutdown(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Shutdown Complete ===")
}
