package main

import (
	"database/sql"
	"flag"
	"log"
	"net"
	"racing/db"
	"racing/proto/racing"
	"racing/service"

	"google.golang.org/grpc"
)

var grpcEndpoint = flag.String("grpc-endpoint", "localhost:9000", "gRPC server endpoint")

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running grpc server: %s\n", err)
	}
}

func run() error {
	conn, err := net.Listen("tcp", ":9000")
	if err != nil {
		return err
	}

	racingDB, err := sql.Open("sqlite3", "./db/racing.db")
	if err != nil {
		return err
	}

	racesRepo := db.NewRacesRepo(racingDB)

	grpcServer := grpc.NewServer()

	racing.RegisterRacingServiceServer(
		grpcServer,
		service.NewRacingService(
			racesRepo,
		),
	)

	log.Printf("gRPC server listening on: %s\n", *grpcEndpoint)

	if err := grpcServer.Serve(conn); err != nil {
		return err
	}

	return nil
}
