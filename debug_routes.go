package main

import (
	"fmt"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/web"
)

func main() {
	config.InitDefaultConfig()
	log.InitLogger()

	// Setup routes to collect them
	web.Handler = routes.NewRoutes()

	// Get the collected routes
	collectedRoutes := models.GetAPITokenRoutes()

	fmt.Println("Collected API Token Routes:")
	for version, versionRoutes := range collectedRoutes {
		fmt.Printf("Version: %s\n", version)
		for group, groupRoutes := range versionRoutes {
			fmt.Printf("  Group: %s\n", group)
			for permission, details := range groupRoutes {
				fmt.Printf("    Permission: %s -> %s %s\n", permission, details.Method, details.Path)
			}
		}
		fmt.Println()
	}
}
