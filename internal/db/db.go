package db

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"k3sdemo/internal/config"

	"github.com/jackc/pgx/v5"
)

func NewPGPool(cfg *config.Config) {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		cfg.DB.User, cfg.DB.Password, net.JoinHostPort(cfg.DB.Host, cfg.DB.Port), cfg.DB.Database,
	)

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
}
