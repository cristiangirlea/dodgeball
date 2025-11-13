package service

import (
	"context"

	pb "apps/dodgeball-go/proto_gen"
	"apps/dodgeball-go/compute"
)

type DodgeballServer struct {
	pb.UnimplementedDodgeballServiceServer
}

func NewDodgeballServer() *DodgeballServer {
	return &DodgeballServer{}
}

func (s *DodgeballServer) RunSimulation(
	ctx context.Context,
	req *pb.SimulationInput,
) (*pb.SimulationResult, error) {

	return compute.RunSimulation(req), nil
}
