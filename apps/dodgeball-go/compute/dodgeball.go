package compute

import (
	pb "apps/dodgeball-go/proto_gen"
	"math"
)

func RunSimulation(input *pb.SimulationInput) *pb.SimulationResult {
	n := len(input.Players)

	alive := make([]bool, n)
	xs := make([]int64, n)
	ys := make([]int64, n)

	for i, p := range input.Players {
		alive[i] = p.Alive
		xs[i] = p.X
		ys[i] = p.Y
	}

	cur := int(input.StartIndex)
	dir := int(input.StartDirection)
	throws := int32(0)

	directionOf := func(ax, ay, bx, by int64) int {
		dx := bx - ax
		dy := by - ay

		if dx == 0 {
			if dy > 0 { return 0 }
			if dy < 0 { return 4 }
			return -1
		}
		if dy == 0 {
			if dx > 0 { return 2 }
			if dx < 0 { return 6 }
			return -1
		}
		if dx == dy {
			if dx > 0 { return 1 }
			return 5
		}
		if dx == -dy {
			if dx > 0 { return 3 }
			return 7
		}
		return -1
	}

	for {
		found := false
		bestIdx := -1
		bestDist := int64(math.MaxInt64)

		for step := 1; step <= 7; step++ {
			scanDir := (dir + step) & 7

			bestIdx = -1
			bestDist = int64(math.MaxInt64)

			for i := 0; i < n; i++ {
				if !alive[i] || i == cur {
					continue
				}

				if directionOf(xs[cur], ys[cur], xs[i], ys[i]) != scanDir {
					continue
				}

				dx := xs[i] - xs[cur]
				dy := ys[i] - ys[cur]
				dist := dx*dx + dy*dy

				if dist < bestDist {
					bestDist = dist
					bestIdx = i
				}
			}

			if bestIdx != -1 {
				// Thrower leaves the field; receiver gets the ball next.
				// The next player "received" the ball from the opposite of the throw direction.
				dir = (scanDir + 4) & 7
				alive[cur] = false
				cur = bestIdx
				throws++
				found = true
				break
			}
		}

		if !found {
			break
		}
	}

	return &pb.SimulationResult{
		Throws:     throws,
		LastPlayer: int32(cur),
	}
}
