//go:build e2e

package e2e

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	pb "apps/dodgeball-go/proto_gen"
	"apps/dodgeball-go/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func dirStrToCode(s string) int32 {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "N":
		return 0
	case "NE":
		return 1
	case "E":
		return 2
	case "SE":
		return 3
	case "S":
		return 4
	case "SW":
		return 5
	case "W":
		return 6
	case "NW":
		return 7
	default:
		return 0
	}
}

func repoRoot(tb testing.TB) string {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile = apps/dodgeball-go/e2e/e2e_test.go
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	return root
}

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

	root := repoRoot(t)
	samples := []string{"sample1", "sample2"}
	for _, name := range samples {
		inPath := filepath.Join(root, "tests", "samples", name+".in")
		outPath := filepath.Join(root, "tests", "samples", name+".out")

		inF, err := os.Open(inPath)
		if err != nil {
			t.Fatalf("open in: %v", err)
		}
		outF, err := os.Open(outPath)
		if err != nil {
			inF.Close()
			t.Fatalf("open out: %v", err)
		}

		inSc := bufio.NewScanner(inF)
		inSc.Split(bufio.ScanWords)
		outSc := bufio.NewScanner(outF)
		outSc.Split(bufio.ScanWords)

		next := func(sc *bufio.Scanner) string {
			if !sc.Scan() {
				t.Fatalf("unexpected EOF")
			}
			return sc.Text()
		}

		T, _ := strconv.Atoi(next(inSc))
		for tc := 0; tc < T; tc++ {
			N, _ := strconv.Atoi(next(inSc))
			players := make([]*pb.Player, 0, N)
			for i := 0; i < N; i++ {
				sx, _ := strconv.ParseInt(next(inSc), 10, 64)
				sy, _ := strconv.ParseInt(next(inSc), 10, 64)
				players = append(players, &pb.Player{X: sx, Y: sy, Alive: true})
			}
			dirStr := next(inSc)
			s, _ := strconv.Atoi(next(inSc))

			res, err := client.RunSimulation(ctx, &pb.SimulationInput{
				Players:        players,
				StartDirection: dirStrToCode(dirStr),
				StartIndex:     int32(s - 1),
			})
			if err != nil {
				inF.Close(); outF.Close()
				t.Fatalf("rpc: %v", err)
			}

			exThrows, _ := strconv.Atoi(next(outSc))
			exLast, _ := strconv.Atoi(next(outSc))

			if int32(exThrows) != res.GetThrows() || int32(exLast-1) != res.GetLastPlayer() {
				inF.Close(); outF.Close()
				t.Fatalf("%s tc%d mismatch: got (%d %d), want (%d %d)", name, tc+1, res.GetThrows(), res.GetLastPlayer()+1, exThrows, exLast)
			}
		}

		inF.Close(); outF.Close()
	}
}
