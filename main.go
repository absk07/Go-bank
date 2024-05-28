package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/absk07/Go-Bank/api"
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connPool, err := pgxpool.New(context.Background(), "postgresql://root:password@localhost:3000/go-bank?sslmode=disable")
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

	err = server.Start(":8080")
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
