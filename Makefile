postgres:
	docker run --name postgres16 --network gobank-network -p 3000:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password -d postgres:16-alpine

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root go-bank

dropdb:
	docker exec -it postgres16 dropdb go-bank

migrateup:
	migrate -path db/migration -database "postgresql://root:password@localhost:3000/go-bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:password@localhost:3000/go-bank?sslmode=disable" -verbose down

migrateup1:
	migrate -path db/migration -database "postgresql://root:password@localhost:3000/go-bank?sslmode=disable" -verbose up 1

migratedown1:
	migrate -path db/migration -database "postgresql://root:password@localhost:3000/go-bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

server:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc server