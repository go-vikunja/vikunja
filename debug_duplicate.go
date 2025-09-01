package main

import (
	"fmt"
	"os"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
)

func main() {
	// Initialize test environment like the tests do
	files.InitTestFileFixtures(nil)
	db.LoadAndAssertFixtures(nil)

	s := db.NewSession()
	defer s.Close()

	service := services.NewProjectDuplicateService(db.GetEngine())
	user1 := &user.User{ID: 1}

	fmt.Printf("User: %+v\n", user1)
	fmt.Println("About to call Duplicate with projectID=1, parentProjectID=0")

	duplicatedProject, err := service.Duplicate(s, 1, 0, user1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! Duplicated project: %+v\n", duplicatedProject)
}
