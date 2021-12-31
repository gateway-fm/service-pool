package pool

import (
	"fmt"
	"time"

	"github.com/gateway-fm/service-pool/discovery"
	"github.com/gateway-fm/service-pool/pkg/logger"
	"github.com/gateway-fm/service-pool/service"
)

// IServicesPool holds information about reachable
// active services, manage connections and discovery
type IServicesPool interface {
	// Start run service pool discovering
	// and healthchecks loops
	Start()

	// DiscoverServices discover all visible active
	// services via service-discovery
	DiscoverServices() error

	// NextService returns next active service
	// to take a connection
	NextService() service.IService

	// Count return numbers of
	// all healthy services in pool
	Count() int

	// List return ServicesPool ServicesList instance
	List() IServicesList

	// Close stop all service pool
	Close()
}

// ServicesPool holds information about reachable
// active services, manage connections and discovery
type ServicesPool struct {
	// TODO maybe is better to change this field to func
	discovery         discovery.IServiceDiscovery
	discoveryInterval time.Duration

	name string

	list IServicesList

	stop chan struct{}

	MutationFnc func(srv service.IService) (service.IService, error)
}

// ServicesPoolsOpts is options that needs
// to configure ServicePool instance
type ServicesPoolsOpts struct {
	Name              string                      // service name to use in service pool
	Discovery         discovery.IServiceDiscovery // discovery interface
	DiscoveryInterval time.Duration               // reconnection interval for unreachable active rediscovery
	ListOpts          *ServicesListOpts           // service list configuration

	MutationFnc func(srv service.IService) (service.IService, error)

	CustomList IServicesList
}

// NewServicesPool create new Services Pool
// based on given params
func NewServicesPool(opts *ServicesPoolsOpts) IServicesPool {
	pool := &ServicesPool{
		discovery:         opts.Discovery,
		discoveryInterval: opts.DiscoveryInterval,
		name:              opts.Name,
		stop:              make(chan struct{}),
		MutationFnc:       opts.MutationFnc,
	}

	if opts.CustomList != nil {
		pool.list = opts.CustomList
	} else {
		pool.list = NewServicesList(opts.Name, opts.ListOpts)

	}

	if err := pool.DiscoverServices(); err != nil {
		logger.Log().Error(fmt.Errorf("error discovering %s services: %w", pool.name, err).Error())
	}

	return pool
}

// Start run service pool discovering
// and healthchecks loops
func (p *ServicesPool) Start() {
	go p.discoverServicesLoop()
	go p.list.HealthChecksLoop()
}

// DiscoverServices discover all visible active
// services via service-discovery
func (p *ServicesPool) DiscoverServices() error {
	newServices, err := p.discovery.Discover(p.name)
	if err != nil {
		return fmt.Errorf("error discovering %s active: %w", p.name, err)
	}

	// TODO for the best scaling we need to change this part to map-based compare mechanic
	for _, newService := range newServices {
		if p.list.IsServiceExists(newService) {
			continue
		}

		mutatedService, err := p.MutationFnc(newService)
		if err != nil {
			return fmt.Errorf("mutate new discovered service: %w", err)
		}

		p.list.Add(mutatedService)
	}
	return nil
}

// NextService returns next active service
// to take a connection
func (p *ServicesPool) NextService() service.IService {
	// TODO maybe is better to return error if next service is nill
	return p.list.Next()
}

// Count return numbers of
// all healthy services in pool
func (p *ServicesPool) Count() int {
	return len(p.list.Healthy())
}

// List return ServicesPool ServicesList instance
func (p *ServicesPool) List() IServicesList {
	return p.list
}

// Close stop all service pool
func (p *ServicesPool) Close() {
	p.list.Close()
	close(p.stop)
}

// discoverServicesLoop spawn discovery for
// services periodically
func (p *ServicesPool) discoverServicesLoop() {
	logger.Log().Info("start discovery loop")

	for {
		select {
		case <-p.stop:
			logger.Log().Warn("stop discovery loop")
			return
		default:
			if err := p.DiscoverServices(); err != nil {
				logger.Log().Warn(fmt.Errorf("error discovery services: %w", err).Error())
			}

			sleep(p.discoveryInterval, p.stop)
		}
	}
}
