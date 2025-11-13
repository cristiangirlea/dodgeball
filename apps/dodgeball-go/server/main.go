package main

import (
	"log"
	"net"
	"os"

	pb "apps/dodgeball-go/proto_gen"
	"apps/dodgeball-go/service"
	"google.golang.org/grpc"
)

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "50051"
	}
	lis, err := net.Listen("tcp", ":"+addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterDodgeballServiceServer(s, service.NewDodgeballServer())

	log.Println("Dodgeball gRPC server running on :" + addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
