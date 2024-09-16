package storage

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func create(ctx context.Context, conn *pgxpool.Pool) error {

	file, err := os.ReadFile("./scripts/init.sql")
	if err != nil {
		return err
	}

	requests := strings.Split(string(file), ";\n\n")

	for _, request := range requests {
		_, err := conn.Exec(ctx, request)
		if err != nil {
			log.Println(err)
		}
	}

	return nil

}
