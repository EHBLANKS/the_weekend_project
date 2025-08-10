package main

import (
	"log"
	"net"
	"sports/service"

	"google.golang.org/grpc"

	db "sports/db"
	sports "sports/proto/sports"
)

func main() {
	sqlDB, err := db.Open("file:sports.db?mode=memory&cache=shared")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.EnsureSchemaAndSeed(sqlDB); err != nil {
		log.Fatalf("seed db: %v", err)
	}

	repo := db.NewEventsRepo(sqlDB)
	svc := service.NewSportsService(repo)

	lis, err := net.Listen("tcp", "localhost:9100")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	sports.RegisterSportsServiceServer(grpcServer, svc)

	log.Println("gRPC server listening on: localhost:9100")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
