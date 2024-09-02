package main

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/absk07/Go-Bank/api/grpc_api"
	"github.com/absk07/Go-Bank/api/rest_api"
	db "github.com/absk07/Go-Bank/db/sqlc"
	"github.com/absk07/Go-Bank/pb"
	"github.com/absk07/Go-Bank/utils"
	"github.com/absk07/Go-Bank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := utils.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Problem loading configs...")
	}

	if config.Env == "Dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	defer connPool.Close()

	runDBMigration(config.DBMigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.Redis_Port,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	var msg string
	err = connPool.QueryRow(context.Background(), "SELECT 'Database successfully connected'").Scan(&msg)
	if err != nil {
		log.Fatal().Err(err).Msg("QueryRow failed")
		// os.Exit(1)
	}
	log.Print(msg)

	go runTaskProcessor(config, redisOpt, store)
	// go runGinServer(config, store, taskDistributor)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}
	log.Print("DB migrated successfully")
}

func runTaskProcessor(config utils.Config, redisOpt asynq.RedisClientOpt, store *db.Store) {
	mailer := utils.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.Info().Msg("starting task processor")

	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func runGrpcServer(config utils.Config, store *db.Store, taskDistributor worker.TaskDistributor) {
	server := grpc_api.NewServer(config, store, taskDistributor)

	grpcLogger := grpc.UnaryInterceptor(utils.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterGoBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPC_Port)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Printf("starting gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start grpc server")
	}
}

func runGatewayServer(config utils.Config, store *db.Store, taskDistributor worker.TaskDistributor) {
	server := grpc_api.NewServer(config, store, taskDistributor)
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
		log.Fatal().Err(err).Msg("cannot register handler server")
	}
	mux := http.NewServeMux()
	mux.Handle("/", grpc_mux)

	listener, err := net.Listen("tcp", config.HTTP_Port)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Info().Msgf("starting HTTP gateway server at %s", listener.Addr().String())
	handler := utils.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start HTTP gateway server")
	}
}

func runGinServer(config utils.Config, store *db.Store, taskDistributor worker.TaskDistributor) {
	server := rest_api.NewServer(config, store, taskDistributor)

	err := server.Start(config.HTTP_Port)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start http server")
	}
}
