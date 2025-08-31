package main

import (
	"fmt"
	"log"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
)

func debugCaldavTaskService() {
	// Load test environment
	err := db.LoadFixtures()
	if err != nil {
		log.Fatalf("Failed to load fixtures: %v", err)
	}

	// Initialize database session
	s := db.NewSession()
	defer s.Close()

	// Get testuser15 (the user used in CalDAV tests)
	testUser := &user.User{}
	has, err := s.Where("username = ?", "user15").Get(testUser)
	if err != nil {
		log.Fatalf("Failed to get test user: %v", err)
	}
	if !has {
		log.Fatalf("Test user15 not found")
	}

	fmt.Printf("Test User: ID=%d, Username=%s\n", testUser.ID, testUser.Username)

	// Test TaskService.GetAllByProject for project 36
	taskService := services.NewTaskService(db.GetEngine())
	tasks, count, total, err := taskService.GetAllByProject(s, 36, testUser, 1, -1, "")
	if err != nil {
		log.Fatalf("Failed to get tasks for project 36: %v", err)
	}

	fmt.Printf("\n=== TaskService.GetAllByProject Results for Project 36 ===\n")
	fmt.Printf("Count: %d, Total: %d\n", count, total)
	fmt.Printf("Tasks returned: %d\n", len(tasks))

	for i, task := range tasks {
		fmt.Printf("Task %d: ID=%d, ProjectID=%d, Title=%s, UID=%s\n",
			i+1, task.ID, task.ProjectID, task.Title, task.UID)
	}

	// Also test for project 1 to compare
	tasks1, count1, total1, err := taskService.GetAllByProject(s, 1, testUser, 1, -1, "")
	if err != nil {
		log.Fatalf("Failed to get tasks for project 1: %v", err)
	}

	fmt.Printf("\n=== TaskService.GetAllByProject Results for Project 1 (for comparison) ===\n")
	fmt.Printf("Count: %d, Total: %d\n", count1, total1)
	fmt.Printf("Tasks returned: %d\n", len(tasks1))

	if len(tasks1) > 5 {
		fmt.Printf("(Showing first 5 tasks only)\n")
		for i := 0; i < 5; i++ {
			task := tasks1[i]
			fmt.Printf("Task %d: ID=%d, ProjectID=%d, Title=%s, UID=%s\n",
				i+1, task.ID, task.ProjectID, task.Title, task.UID)
		}
	} else {
		for i, task := range tasks1 {
			fmt.Printf("Task %d: ID=%d, ProjectID=%d, Title=%s, UID=%s\n",
				i+1, task.ID, task.ProjectID, task.Title, task.UID)
		}
	}

	s.Commit()
}

func main() {
	debugCaldavTaskService()
}
