package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := utils.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Problem loading configs...")
	}

	if config.Env == "Dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	connPool, err := pgxpool.New(ctx, config.DBSource)
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

	waitGroup, ctx := errgroup.WithContext(ctx)

	var msg string
	err = connPool.QueryRow(ctx, "SELECT 'Database successfully connected'").Scan(&msg)
	if err != nil {
		log.Fatal().Err(err).Msg("QueryRow failed")
		// os.Exit(1)
	}
	log.Print(msg)

	runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
	// runGinServer(config, store, taskDistributor)
	runGatewayServer(ctx, waitGroup, config, store, taskDistributor)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
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

func runTaskProcessor(ctx context.Context, wg *errgroup.Group, config utils.Config, redisOpt asynq.RedisClientOpt, store *db.Store) {
	mailer := utils.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.Info().Msg("starting task processor")

	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutdown task processor")

		taskProcessor.Shutdown()
		log.Info().Msg("task processor is stopped")

		return nil
	})
}

func runGrpcServer(ctx context.Context, wg *errgroup.Group, config utils.Config, store *db.Store, taskDistributor worker.TaskDistributor) {
	server := grpc_api.NewServer(config, store, taskDistributor)

	grpcLogger := grpc.UnaryInterceptor(utils.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterGoBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPC_Port)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	wg.Go(func() error {
		log.Printf("starting gRPC server at %s", listener.Addr().String())
		err = grpcServer.Serve(listener)
		if err != nil {
			log.Error().Err(err).Msg("cannot start grpc server")
			return err
		}
		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutdown gRPC server")
		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server is stopped")
		return nil
	})
}

func runGatewayServer(ctx context.Context, wg *errgroup.Group, config utils.Config, store *db.Store, taskDistributor worker.TaskDistributor) {
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

	err := pb.RegisterGoBankHandlerServer(ctx, grpc_mux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server")
	}
	mux := http.NewServeMux()
	mux.Handle("/", grpc_mux)

	httpServer := &http.Server{
		Handler: utils.HttpLogger(mux),
		Addr:    config.HTTP_Port,
	}

	wg.Go(func() error {
		log.Info().Msgf("start HTTP gateway server at %s", httpServer.Addr)
		err = httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("HTTP gateway server failed to serve")
			return err
		}
		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutdown HTTP gateway server")

		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown HTTP gateway server")
			return err
		}

		log.Info().Msg("HTTP gateway server is stopped")
		return nil
	})
}

func runGinServer(config utils.Config, store *db.Store, taskDistributor worker.TaskDistributor) {
	server := rest_api.NewServer(config, store, taskDistributor)

	err := server.Start(config.HTTP_Port)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start gin HTTP server")
	}
}
