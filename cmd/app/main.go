package main

import (
	"invento-service/config"
	"invento-service/internal/app"
	"log"
)

// @title Invento Service API
// @version 1.0
// @description REST API service for managing projects, modules, users, and file uploads with JWT authentication and RBAC.
// @host localhost:3000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
// @description Access token stored in HttpOnly cookie (set automatically on login/register/refresh)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDatabase(cfg)

	server := app.NewServer(cfg, db)

	log.Printf("Server berjalan di port %s", cfg.App.Port)
	log.Fatal(server.Listen(":" + cfg.App.Port))
}
