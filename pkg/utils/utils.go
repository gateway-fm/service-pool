package utils

import (
	"math/rand"
	"time"

	"github.com/gateway-fm/prover-pool-lib/service"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func ShuffleSlice(slice []service.IService) {
	swap := func(i int, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	}
	rand.Shuffle(len(slice), swap)
}

// RandomUint64 returns a random uint64 in range [0, n) where
// n max value is int32. We use int type of n because we only use
// length of healthy nodes slice as an input. And len(slice) has max value of int32.
func RandomUint64(n int) uint64 {
	return uint64(rand.Intn(n))
}
