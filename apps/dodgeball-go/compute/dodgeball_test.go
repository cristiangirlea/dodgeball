package compute

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	pb "apps/dodgeball-go/proto_gen"
)

func dirStrToCode(s string) int {
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

// helper to get repository root relative to this file
func repoRoot(t *testing.T) string {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile = apps/dodgeball-go/compute/dodgeball_test.go
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	return root
}

func TestSample2(t *testing.T) {
	root := repoRoot(t)
	inPath := filepath.Join(root, "tests", "samples", "sample2.in")
	outPath := filepath.Join(root, "tests", "samples", "sample2.out")

	inF, err := os.Open(inPath)
	if err != nil {
		t.Fatalf("open in: %v", err)
	}
	defer inF.Close()
	outF, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("open out: %v", err)
	}
	defer outF.Close()

	inSc := bufio.NewScanner(inF)
	inSc.Split(bufio.ScanWords)
	outSc := bufio.NewScanner(outF)
	outSc.Split(bufio.ScanWords)

	next := func(sc *bufio.Scanner) string {
		if !sc.Scan() {
			t.Fatalf("unexpected EOF while reading")
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

		input := &pb.SimulationInput{
			Players:        players,
			StartDirection: int32(dirStrToCode(dirStr)),
			StartIndex:     int32(s - 1),
		}
		res := RunSimulation(input)

		exThrows, _ := strconv.Atoi(next(outSc))
		exLast, _ := strconv.Atoi(next(outSc))

		if int32(exThrows) != res.Throws || int32(exLast-1) != res.LastPlayer {
			t.Fatalf("tc%d mismatch: got (%d %d), want (%d %d)", tc+1, res.Throws, res.LastPlayer+1, exThrows, exLast)
		}
	}
}
