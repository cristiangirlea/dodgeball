package service

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"apps/dodgeball-go/compute"
	pb "apps/dodgeball-go/proto_gen"
)

type DodgeballServer struct {
	pb.UnimplementedDodgeballServiceServer
}

func NewDodgeballServer() *DodgeballServer {
	return &DodgeballServer{}
}

var logIO = os.Getenv("DODGEBALL_LOG_IO") != ""

func (s *DodgeballServer) RunSimulation(
	ctx context.Context,
	req *pb.SimulationInput,
) (*pb.SimulationResult, error) {
 start := time.Now()
 if logIO {
 	if b, err := (protojson.MarshalOptions{EmitUnpopulated: true}).Marshal(req); err == nil {
 		log.Printf("[IO] RunSimulation request: %s", string(b))
 	} else {
 		log.Printf("[IO] RunSimulation request (marshal error=%v): %+v", err, req)
 	}
 }

 res := compute.RunSimulation(req)

 if logIO {
 	if b, err := (protojson.MarshalOptions{EmitUnpopulated: true}).Marshal(res); err == nil {
 		log.Printf("[IO] RunSimulation response: %s (duration=%s)", string(b), time.Since(start))
 	} else {
 		log.Printf("[IO] RunSimulation response (marshal error=%v): %+v (duration=%s)", err, res, time.Since(start))
 	}
 }

 return res, nil
}
