package pool

import (
	"time"

	"github.com/gateway-fm/prover-pool-lib/prover"
)

func Example() {
	pool := NewServicesPool(&ServicesPoolsOpts{
		Name: "example",
		ListOpts: &ServicesListOpts{
			TryUpTries:     5,
			TryUpInterval:  5 * time.Second,
			ChecksInterval: 5 * time.Second,
		},
	})

	pool.Start(true)

	// when prover calls prover-pool-service
	prv, err := prover.NewProver(&prover.ProverOpts{
		Name:        "exampleProver1",
		Addr:        "127.0.0.1:8080",
		Healthcheck: ProverMockHealthcheck(10 * time.Second),
		Tags: map[string]struct{}{
			"forkId1": struct{}{},
			"public":  struct{}{},
		},
	})
	if err != nil {
		panic(err)
	}

	pool.AddService(prv)

	// when we need to call prover we do:
	srv := pool.NextLeastLoaded("forkId1")
	leastLoadedPrv, ok := srv.(prover.IProver)
	if !ok {
		panic("oh shit")
	}

	resp, err := leastLoadedPrv.DoRequest(nil)
	if err != nil {
		panic(err)
	}
	_ = resp
}
