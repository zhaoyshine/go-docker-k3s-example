package main

import (
	"k3sdemo/internal/api"
	"k3sdemo/internal/config"
	"k3sdemo/internal/db"
)

func main() {
	cfg := config.LoadYamlConfig()
	db.NewPGPool(cfg)
	api.StartHttp(cfg)
}
