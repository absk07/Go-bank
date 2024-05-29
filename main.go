package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	connPool, err := pgxpool.New(context.Background(), config.DBUri)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer connPool.Close()

	store := db.NewStore(connPool)

	var msg string
	err = connPool.QueryRow(context.Background(), "SELECT 'Database successfully connected'").Scan(&msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(msg)
	
	server := api.NewServer(store)

	err = server.Start(config.Port)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
