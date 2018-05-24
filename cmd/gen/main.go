package main

import (
	"log"
	"experimental/project-generator/pkg/configuration"
	"experimental/project-generator/pkg/services"
	"fmt"
)

func main() () {
	// Read config file.
	config, err := configuration.LoadConfig()
	if err != nil {
		log.Fatalf("FAILURE: %v\n", err.Error())
	}

	// Generate the project.
	err = services.GenerateProject(config)
	if err != nil {
		log.Fatalf("FAILURE: %v\n", err.Error())
	}

	fmt.Printf("SUCCESS: Project created at %v.\n", config.OutputDirectoryLocation)
}
