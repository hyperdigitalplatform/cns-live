package main

import (
	"fmt"
	"log"

	"milestone-service/internal/api"
	"milestone-service/internal/config"
	"milestone-service/internal/rest"
	"milestone-service/internal/service"
	"milestone-service/internal/soap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Milestone Service...")
	log.Printf("Milestone Server: %s", cfg.Milestone.BaseURL)
	log.Printf("Server Port: %d", cfg.Server.Port)

	// Create SOAP client (for recording control)
	soapClient := soap.NewClient(
		cfg.Milestone.BaseURL,
		cfg.Milestone.Username,
		cfg.Milestone.Password,
	)

	// Create REST client (for camera discovery)
	restClient := rest.NewClient(
		cfg.Milestone.BaseURL,
		cfg.Milestone.Username,
		cfg.Milestone.Password,
	)

	// Create discovery service
	discoveryService := service.NewDiscoveryService(restClient)

	// Create API handler
	handler := api.NewHandler(soapClient, restClient, discoveryService)

	// Setup router
	router := api.SetupRouter(handler)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
