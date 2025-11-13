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


func TestRunSimulation_ExactIO(t *testing.T) {
	// This test constructs typed SimulationInput objects and asserts the exact
	// numeric fields of SimulationResult. It is the source of truth for the
	// simulation I/O contract in unit tests (no file parsing involved).
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
			name: "single player no throws",
			in: &pb.SimulationInput{
				Players:        mkPlayers([2]int64{0, 0}),
				StartDirection: 0, // N
				StartIndex:     0,
			},
			wantThrows: 0,
			wantLast:   0,
		},
		{
			name: "two players aligned with startDir -> will throw after checking all, including incoming dir",
			in: &pb.SimulationInput{
				Players:        mkPlayers([2]int64{0, 0}, [2]int64{1, 0}), // target due East
				StartDirection: 2,                                        // E (current dir is checked last)
				StartIndex:     0,
			},
			wantThrows: 1,
			wantLast:   1,
		},
		{
			name: "two players east with startDir north -> one throw to east",
			in: &pb.SimulationInput{
				Players:        mkPlayers([2]int64{0, 0}, [2]int64{2, 0}), // target to the East
				StartDirection: 0,                                        // N
				StartIndex:     0,
			},
			wantThrows: 1,
			wantLast:   1,
		},
		{
			name: "chain two throws across three players on a line east",
			in: &pb.SimulationInput{
				Players:        mkPlayers([2]int64{0, 0}, [2]int64{1, 0}, [2]int64{2, 0}),
				StartDirection: 0, // N; scan finds E, then next dir becomes W, then finds E to player2
				StartIndex:     0,
			},
			wantThrows: 2,
			wantLast:   2,
		},
		{
			name: "no reachable direction",
			in: &pb.SimulationInput{
				Players:        mkPlayers([2]int64{0, 0}, [2]int64{1, 2}), // not aligned N,NE,E,SE,S,SW,W,NW
				StartDirection: 0,
				StartIndex:     0,
			},
			wantThrows: 0,
			wantLast:   0,
		},
		{
			name: "diagonal NE target found when scanning from N",
			in: &pb.SimulationInput{
				Players:        mkPlayers([2]int64{0, 0}, [2]int64{2, 2}), // NE
				StartDirection: 0,                                        // N; scanDir=1 (NE) will match
				StartIndex:     0,
			},
			wantThrows: 1,
			wantLast:   1,
		},
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
				StartDirection: 7, // NW incoming
				StartIndex:     4, // player 5 (0-based 4)
			},
			wantThrows: 4,
			wantLast:   7, // player 8 (0-based 7)
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RunSimulation(tc.in)
			if got.Throws != tc.wantThrows || got.LastPlayer != tc.wantLast {
				t.Fatalf("%s: got {throws=%d, last=%d}, want {throws=%d, last=%d}", tc.name, got.Throws, got.LastPlayer, tc.wantThrows, tc.wantLast)
			}
		})
	}
}
