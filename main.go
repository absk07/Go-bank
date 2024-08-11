package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/absk07/Go-Bank/api/grpc_api"
	"github.com/absk07/Go-Bank/api/rest_api"
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
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

	// runGinServer(config, store)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
}

func runGrpcServer(config utils.Config, store *db.Store) {
	grpcServer := grpc.NewServer()
	server := grpc_api.NewServer(config, store)
	pb.RegisterGoBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPC_Port)
	if err != nil {
		log.Fatal("cannot create listener: ", err)
	}
	log.Printf("starting gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start grpc server: ", err)
	}
}

func runGatewayServer(config utils.Config, store *db.Store) {
	server := grpc_api.NewServer(config, store)
	grpc_mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	err := pb.RegisterGoBankHandlerServer(ctx, grpc_mux, server)
	if err != nil {
		log.Fatal("cannot register handler server: ", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", grpc_mux)

	listener, err := net.Listen("tcp", config.HTTP_Port)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}
	log.Printf("starting HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server:", err)
	}
}

func runGinServer(config utils.Config, store *db.Store) {
	server := rest_api.NewServer(config, store)

	err := server.Start(config.HTTP_Port)
	if err != nil {
		log.Fatal("cannot start http server:", err)
	}
}
