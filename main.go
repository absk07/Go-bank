package main

import (
	"context"
	"fmt"
	"log"

	"github.com/absk07/Go-Bank/api"
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := utils.LoadConfig()
	if err != nil {
		log.Fatal("Problem loading configs...")
	}
	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer connPool.Close()

	store := db.NewStore(connPool)

	var msg string
	err = connPool.QueryRow(context.Background(), "SELECT 'Database successfully connected'").Scan(&msg)
	if err != nil {
		log.Fatal("QueryRow failed:", err)
		// os.Exit(1)
	}
	fmt.Println(msg)

	server := api.NewServer(config, store)

	err = server.Start(config.Port)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
