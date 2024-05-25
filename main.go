package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connPool, err := pgxpool.New(context.Background(), "postgresql://root:password@localhost:3000/go-bank?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer connPool.Close()

	var msg string
	err = connPool.QueryRow(context.Background(), "select 'Database successfully connected").Scan(&msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(msg)
}
