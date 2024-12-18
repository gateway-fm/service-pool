package pool

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gateway-fm/prover-pool-lib/prover"
	srv "github.com/gateway-fm/prover-pool-lib/service"
	"github.com/gateway-fm/scriptorium/logger"
)

const (
	maxHCNumTries        = 5
	hcRetrySleepInterval = time.Millisecond * 200
)

type HealthcheckFunc func(timeOut time.Duration, p prover.IProver) (bool, error)

func ProverMockHealthcheck(timeOut time.Duration) func(iProver prover.IProver) error {
	return func(p prover.IProver) error {
		return healthcheckWithRetry(timeOut, p, 0, nil, proverMockHealthcheck)
	}
}

func proverMockHealthcheck(timeOut time.Duration, p prover.IProver) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	_ = ctx

	load := rand.Float32()
	p.SetStatus(srv.StatusHealthy)
	p.SetLoad(load)

	return false, nil
}

func healthcheckWithRetry(
	timeOut time.Duration, p prover.IProver,
	try int, lastErr error,
	hcFunc HealthcheckFunc) error {

	if try >= maxHCNumTries {
		p.SetStatus(srv.StatusUnHealthy)
		return lastErr
	}

	if try > 0 {
		logger.Log().Warn(fmt.Sprintf("retrying healthcheck to service %s %s... current try is %d out of %d, the last error was: %s", p.NodeName(), p.ID(), try+1, maxHCNumTries, lastErr.Error()))
	}

	retryNeeded, err := hcFunc(timeOut, p)
	if err == nil {
		if try > 0 {
			logger.Log().Info(fmt.Sprintf("the healthcheck to service %s %s is recovered after retry (•‿•)", p.NodeName(), p.ID()))
		}
		return nil
	}

	if !retryNeeded {
		return err
	}

	time.Sleep(hcRetrySleepInterval)
	return healthcheckWithRetry(timeOut, p, try+1, err, hcFunc)
}
