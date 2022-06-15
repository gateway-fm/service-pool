package utils

import (
	"math/rand"
	"time"

	"github.com/gateway-fm/service-pool/service"
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

const maxInt64 uint64 = 1<<63 - 1

func RandomUint64(min, max uint64) uint64 {
	return uint64Helper(max-min) + min
}

// https://stackoverflow.com/questions/47856543/generate-random-uint64-between-min-and-max
func uint64Helper(n uint64) uint64 {
	if n < maxInt64 {
		return uint64(rand.Int63n(int64(n + 1)))
	}
	x := rand.Uint64()
	for x > n {
		x = rand.Uint64()
	}
	return x
}
