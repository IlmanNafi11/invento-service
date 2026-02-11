package main

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/app"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDatabase(cfg)

	server := app.NewServer(cfg, db)

	log.Printf("Server berjalan di port %s", cfg.App.Port)
	log.Fatal(server.Listen(":" + cfg.App.Port))
}
