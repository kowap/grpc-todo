package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"grpc-todo/proto"
	"grpc-todo/repository"
	"grpc-todo/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoClient, err := repository.ConnectToMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database("grpc_todo_db")
	repo := repository.NewRepository(db)

	grpcServer := grpc.NewServer()
	todoServer := server.NewToDoServer(repo)
	proto.RegisterToDoServiceServer(grpcServer, todoServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	cronJob := todoServer.StartCronJob()
	defer cronJob.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		log.Println("Server is running on port :50051")
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-quit:
		log.Println("Shutting down server...")
		grpcServer.GracefulStop()
		log.Println("Server gracefully stopped.")
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
	}
}