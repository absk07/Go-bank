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

proto:
	rm -f pb/*.go
	rm -f docs/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=docs/swagger --openapiv2_opt=allow_merge=true \
    proto/*.proto

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc server proto