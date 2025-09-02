package main

import (
	"fmt"
	"log"
	"os"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
)

func testFilter() {
	// Set up minimal environment
	config.InitDefaultConfig()
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))

	// Initialize database and services
	err := models.SetEngine()
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// Initialize services
	services.InitTaskService()

	// Create TaskCollection with the exact problematic parameters
	collection := &models.TaskCollection{
		ProjectID:          0,
		SortByArr:          []string{"due_date", "id"},
		OrderByArr:         []string{"asc", "desc"},
		Filter:             "done = false",
		FilterIncludeNulls: false,
		FilterTimezone:     "GMT",
	}

	fmt.Printf("Testing filter parsing: %s\n", collection.Filter)

	// Create a test session and user
	s := db.NewSession()
	defer s.Close()
	u := &user.User{ID: 1}

	fmt.Println("Testing TaskCollection.ReadAll() with problematic parameters...")

	// This should reproduce the actual bug
	result, resultCount, totalItems, err := collection.ReadAll(s, u, "", 1, 50)

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		fmt.Printf("Error type: %T\n", err)
		if models.IsErrInvalidTaskFilterValue(err) {
			fmt.Printf("This is a filter validation error!\n")
		}
		os.Exit(1)
	} else {
		fmt.Printf("SUCCESS: Got %d results, %d total items\n", resultCount, totalItems)
		fmt.Printf("Result type: %T\n", result)
	}
}
