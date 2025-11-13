//go:build e2e

package e2e

import (
	"context"
	"net"
	"testing"
	"time"

	pb "apps/dodgeball-go/proto_gen"
	"apps/dodgeball-go/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func startServer(tb testing.TB) (addr string, stop func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDodgeballServiceServer(s, service.NewDodgeballServer())
	go func() {
		_ = s.Serve(lis)
	}()
	return lis.Addr().String(), func() { s.GracefulStop() }
}

func TestGRPCE2E_Samples1And2(t *testing.T) {
	addr, stop := startServer(t)
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewDodgeballServiceClient(conn)

	mkPlayers := func(coords ...[2]int64) []*pb.Player {
		ps := make([]*pb.Player, 0, len(coords))
		for _, c := range coords {
			ps = append(ps, &pb.Player{X: c[0], Y: c[1], Alive: true})
		}
		return ps
	}

	cases := []struct {
		name       string
		in         *pb.SimulationInput
		wantThrows int32
		wantLast   int32
	}{
		{
			name: "spec sample1 testcase1",
			in: &pb.SimulationInput{
				Players: mkPlayers(
					[2]int64{-10, -10},
					[2]int64{-10, 10},
					[2]int64{0, -10},
					[2]int64{0, 10},
					[2]int64{10, -10}, // starting player (index 4 -> 0-based)
					[2]int64{10, 10},
					[2]int64{-9, -10},
					[2]int64{-9, 0},
				),
				StartDirection: 7, // NW
				StartIndex:     4, // player 5 -> 0-based
			},
			wantThrows: 4,
			wantLast:   7, // player 8 -> 0-based
		},
		{
			name: "spec sample1 testcase2",
			in: &pb.SimulationInput{
				Players: mkPlayers(
					[2]int64{-1000000, -1000000},
					[2]int64{-1000000, 1000000},
					[2]int64{0, -1000000},
					[2]int64{0, 1000000},
					[2]int64{1000000, -1000000},
					[2]int64{1000000, 1000000},
					[2]int64{-999999, -1000000},
					[2]int64{-999999, 0},
				),
				StartDirection: 3, // SE
				StartIndex:     3, // player 4 -> 0-based
			},
			wantThrows: 5,
			wantLast:   5, // player 6 -> 0-based
		},
	}

	for _, tc := range cases {
		res, err := client.RunSimulation(ctx, tc.in)
		if err != nil {
			t.Fatalf("rpc %s: %v", tc.name, err)
		}
		if res.GetThrows() != tc.wantThrows || res.GetLastPlayer() != tc.wantLast {
			t.Fatalf("%s mismatch: got (%d %d), want (%d %d)", tc.name, res.GetThrows(), res.GetLastPlayer()+1, tc.wantThrows, tc.wantLast+1)
		}
	}
}
