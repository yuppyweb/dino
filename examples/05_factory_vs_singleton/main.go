package main

import (
	"fmt"
	"log"

	"github.com/yuppyweb/dino"
)

type RequestID struct {
	ID string
}

type Logger struct {
	Name string
}

// Example demonstrating Factory vs Singleton patterns.
func main() {
	di := dino.New()

	// Use Factory for values that should be created fresh
	// (though Factory still caches after first call in Dino)
	if err := di.Factory(func() *RequestID {
		id := &RequestID{ID: "req-123"}
		fmt.Printf("Factory creating new RequestID: %s\n", id.ID)

		return id
	}); err != nil {
		log.Fatal(err)
	}

	// Use Singleton for shared instances
	logger := &Logger{Name: "MyApp"}
	fmt.Printf("Singleton registering Logger: %s\n", logger.Name)

	if err := di.Singleton(logger); err != nil {
		log.Fatal(err)
	}

	// Inject into multiple structs
	type Service1 struct {
		RequestID *RequestID
		Logger    *Logger
	}

	type Service2 struct {
		RequestID *RequestID
		Logger    *Logger
	}

	s1 := &Service1{}
	if err := di.Inject(s1); err != nil {
		log.Fatal(err)
	}

	s2 := &Service2{}
	if err := di.Inject(s2); err != nil {
		log.Fatal(err)
	}

	// Verify that Singleton returns same instance
	fmt.Printf("Service1 Logger: %p (%s)\n", s1.Logger, s1.Logger.Name)
	fmt.Printf("Service2 Logger: %p (%s)\n", s2.Logger, s2.Logger.Name)
	fmt.Printf("Same Logger instance: %v\n", s1.Logger == s2.Logger)

	// Factory caches the result after first call
	fmt.Printf("Service1 RequestID: %p (%s)\n", s1.RequestID, s1.RequestID.ID)
	fmt.Printf("Service2 RequestID: %p (%s)\n", s2.RequestID, s2.RequestID.ID)
	fmt.Printf("Same RequestID instance (cached): %v\n", s1.RequestID == s2.RequestID)
}
