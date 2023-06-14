package pool

import (
	"fmt"
	"time"

	"github.com/gateway-fm/service-pool/discovery"
	"github.com/gateway-fm/service-pool/service"
)

type healthyService struct {
	HCheckCounter int
	*service.BaseService
}

func (s *healthyService) HealthCheck() error {
	if s != nil {
		s.HCheckCounter = s.HCheckCounter + 1
		s.SetStatus(service.StatusHealthy)
	}
	return nil
}

func healthySrvMutationFunc(srv service.IService) (service.IService, error) {
	baseSrv, ok := srv.(*service.BaseService)
	if !ok {
		return nil, fmt.Errorf("service is not BaseService")
	}

	return &healthyService{
		0,
		baseSrv,
	}, nil
}

func dummyMutationFunc(srv service.IService) (service.IService, error) {
	return srv, nil
}

func newHealthyService(addr string) service.IService {
	srv := service.NewService(addr, "", nil)

	baseSrv := srv.(*service.BaseService)
	baseSrv.SetStatus(service.StatusHealthy)

	return baseSrv
}

func newServicesPool(discoveryInterval time.Duration, hcInterval time.Duration, mutationFunc func(srv service.IService) (service.IService, error)) IServicesPool {
	manualDisc, _ := discovery.NewManualDiscovery(discovery.TransportHttp, "localhost")

	opts := &ServicesPoolsOpts{
		Name:              "TestServicePool",
		Discovery:         manualDisc,
		DiscoveryInterval: discoveryInterval,
		ListOpts: &ServicesListOpts{
			TryUpTries:     5,
			TryUpInterval:  1 * time.Second,
			ChecksInterval: hcInterval,
		},
		MutationFnc: mutationFunc,
	}

	return NewServicesPool(opts)
}
