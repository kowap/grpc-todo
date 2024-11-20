package main

import (
	"log"
	"net"

	"grpc-todo/proto"
	"grpc-todo/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterToDoServiceServer(grpcServer, server.NewToDoServer())

	// Регистрируем сервис Reflection на gRPC-сервере
	reflection.Register(grpcServer)

	log.Println("Server is running on port :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
