package pool

import (
	"github.com/gateway-fm/service-pool/discovery"
	"github.com/gateway-fm/service-pool/service"
	"testing"
	"time"
)

// TestServicesPoolStart tests whether discovery loop is spawned on pool.Start()
func TestServicesPoolStart(t *testing.T) {
	pool := newServicesPool(1*time.Second, 1*time.Second, dummyMutationFunc)
	pool.Start(true)

	time.Sleep(200 * time.Millisecond) // wait until discovery and healthchecks are finished

	count := pool.Count()
	if count != 1 {
		t.Errorf("num services in pool want 1, got: %d", count)
	}

	// ok, the service was discovered

	nextSrv := pool.NextService()
	if nextSrv != nil {
		t.Errorf("next service is not nil, got id: %s, status: %s", nextSrv.ID(), nextSrv.Status())
	}

	// ok, the service is in healthy slice, but has status UnHealthy (coz healthcheck func returns nil)
}

// TestServicesPoolHealthCheckLoop tests whether hc loop is spawned on pool.Start()
// and how many times hc is called if checksInterval=1s
func TestServicesPoolHealthCheckLoop(t *testing.T) {
	pool := newServicesPool(100*time.Hour, 1*time.Second, healthySrvMutationFunc)
	pool.Start(true)

	time.Sleep(200 * time.Millisecond) // wait until discovery and healthchecks are finished

	pool.List().SetOnSrvAddCallback(func(srv service.IService) error { // to make srv being always added as unhealthy
		s, _ := srv.(*healthyService)
		s.SetStatus(service.StatusUnHealthy)
		return nil
	})

	for i := 0; i < 5; i++ {
		pool.List().RemoveFromHealthyByIndex(0)
		pool.List().Add(&healthyService{0, &service.BaseService{}}) // add service (callback makes it unhealthy)

		if pool.NextService() != nil {
			t.Errorf("unexpected healthy service was found")
		}

		time.Sleep(1 * time.Second) // wait until healthcheck is finished

		if pool.NextService() == nil {
			t.Errorf("unexpected no healthy services occured")
		}
	}

	time.Sleep(3 * time.Second) // wait until 3 healthchecks are done

	healthchecksDoneCount := pool.List().Next().(*healthyService).HCheckCounter

	if healthchecksDoneCount > 5 { // 5 is because 1 (healthcheck during add) + 1 (seconds passed) + 3 (seconds passed)
		t.Errorf("Too much healthchecks are done")
	}
}

func TestServicesPoolCustomList(t *testing.T) {
	manualDisc, _ := discovery.NewManualDiscovery(discovery.TransportHttp, "localhost")

	customList := NewServicesList("testServiceList", &ServicesListOpts{
		TryUpTries:     100,
		TryUpInterval:  100 * time.Second,
		ChecksInterval: 100 * time.Second,
	})
	customList.Add(&healthyService{0, &service.BaseService{}})

	opts := &ServicesPoolsOpts{
		Name:              "TestServicePool",
		Discovery:         manualDisc,
		DiscoveryInterval: 1 * time.Second,
		ListOpts: &ServicesListOpts{
			TryUpTries:     5,
			TryUpInterval:  1 * time.Second,
			ChecksInterval: 1 * time.Second,
		},
		MutationFnc: func(srv service.IService) (service.IService, error) {
			return srv, nil
		},
		CustomList: customList,
	}
	pool := NewServicesPool(opts)

	if pool.List().Next() == nil {
		t.Errorf("unexpected no healthy services")
	}

	if pool.List().Unhealthy() != nil {
		t.Errorf("unexpected unhealthy services")
	}
}

func TestServicesPoolDiscoveryError(t *testing.T) {
	consulDisc, _ := discovery.NewConsulDiscovery(discovery.TransportHttp, "localhost")
	opts := &ServicesPoolsOpts{
		Name:              "TestServicePool",
		Discovery:         consulDisc,
		DiscoveryInterval: 1 * time.Second,
		ListOpts: &ServicesListOpts{
			TryUpTries:     5,
			TryUpInterval:  1 * time.Second,
			ChecksInterval: 1 * time.Second,
		},
		MutationFnc: func(srv service.IService) (service.IService, error) {
			return srv, nil
		},
	}

	pool := NewServicesPool(opts)
	if err := pool.DiscoverServices(); err == nil {
		t.Errorf("unexpected nil error during consul discovery")
	}
}

func TestServicesPoolClose(t *testing.T) {
	pool := newServicesPool(1*time.Second, 1*time.Second, healthySrvMutationFunc)
	pool.Start(true)

	time.Sleep(200 * time.Millisecond)

	if pool.NextService() == nil {
		t.Errorf("unexpected no healthy services")
	}

	pool.List().RemoveFromHealthyByIndex(0)
	pool.Close()
	time.Sleep(1 * time.Second)

	if pool.NextService() != nil {
		t.Errorf("unexpected healthy service was found")
	}
}

func TestServicesPoolCallbacks(t *testing.T) {
	pool := newServicesPool(1*time.Second, 1*time.Second, healthySrvMutationFunc)

	onNewDiscIsExecuted := false
	pool.SetOnNewDiscCallback(func(srv service.IService) error {
		onNewDiscIsExecuted = true
		return nil
	})

	onDiscCompletedIsExecuted := false
	pool.SetOnDiscCompletedCallback(func() {
		onDiscCompletedIsExecuted = true
	})

	pool.Start(true)
	time.Sleep(200 * time.Millisecond)

	if !onNewDiscIsExecuted {
		t.Errorf("onNewDiscCallback was not executed")
	}

	if !onDiscCompletedIsExecuted {
		t.Errorf("onDiscCompletedCallback was not executed")
	}
}
