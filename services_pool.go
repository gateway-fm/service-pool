package pool

import (
	"github.com/gateway-fm/prover-pool-lib/service"
)

// IServicesPool holds information about reachable
// active services, manage connections and discovery
type IServicesPool interface {
	// Start run service pool discovering
	// and healthchecks loops
	Start(healthchecks bool)

	// NextService returns next active service
	// to take a connection
	NextService() service.IService

	ServiceById(id string) service.IService

	SetProverLoadById(id string, load *service.ProverLoad)
	// Count return numbers of
	// all healthy services in pool
	Count() int

	// List return ServicesPool ServicesList instance
	//List() IServicesList

	// Close Stop all service pool
	Close()

	AddService(srv service.IService)

	AnyByTag(tag string) service.IService

	NextLeastLoaded(tag string) service.IService

	NextLeastLoadedProver(tag string, upload bool) service.IService
	// FromHealthyToJail move Unhealthy service
	// from Healthy slice to Jail map
	FromHealthyToJail(id string)

	// FromJailToHealthy move Healthy service
	// from Jail map to Healthy slice
	FromJailToHealthy(srv service.IService)

	// RemoveFromJail remove given
	// service from jail map
	RemoveFromJail(srv service.IService)
}

// ServicesPool holds information about reachable
// active services, manage connections and discovery
type ServicesPool struct {
	name string

	list IServicesList

	stop chan struct{}

	MutationFnc func(srv service.IService) (service.IService, error)
}

// ServicesPoolsOpts is options that needs
// to configure ServicePool instance
type ServicesPoolsOpts struct {
	Name     string            // service name to use in service pool
	ListOpts *ServicesListOpts // service list configuration
}

type ServiceCallbackE func(srv service.IService) error
type ServiceCallback func(srv service.IService)
type ServiceCallbackB func(srv service.IService) bool

// NewServicesPool create new Services Pool
// based on given params
func NewServicesPool(opts *ServicesPoolsOpts) IServicesPool {
	pool := &ServicesPool{
		name: opts.Name,
		stop: make(chan struct{}),
	}

	pool.list = NewServicesList(opts.Name, opts.ListOpts)

	return pool
}

// Start run service pool discovering
// and healthchecks loops
func (p *ServicesPool) Start(healthchecks bool) {
	if healthchecks {
		go p.list.HealthChecksLoop()
	}
}

// NextService returns next active service
// to take a connection
func (p *ServicesPool) NextService() service.IService {
	// TODO maybe is better to return error if next service is nil
	return p.list.Next()
}

func (p *ServicesPool) NextLeastLoaded(tag string) service.IService {
	// TODO maybe is better to return error if next service is nil
	return p.list.NextLeastLoaded(tag)
}

func (p *ServicesPool) NextLeastLoadedProver(tag string, upload bool) service.IService {
	// TODO maybe is better to return error if next service is nil
	return p.list.NextLeastLoadedProver(tag, upload)
}

func (p *ServicesPool) ServiceById(id string) service.IService {
	return p.list.ServiceById(id)
}

func (p *ServicesPool) SetProverLoadById(id string, load *service.ProverLoad) {
	p.list.SetProverLoadById(id, load)
}

func (p *ServicesPool) AddService(srv service.IService) {
	p.list.Add(srv)
}

// Count return numbers of
// all healthy services in pool
func (p *ServicesPool) Count() int {
	return len(p.list.Healthy())
}

// List return ServicesPool ServicesList instance
//func (p *ServicesPool) List() IServicesList {
//	return p.list
//}

// Close Stop all service pool
func (p *ServicesPool) Close() {
	p.list.Close()
	close(p.stop)
}

func (p *ServicesPool) AnyByTag(tag string) service.IService {
	return p.list.AnyByTag(tag)
}

func (p *ServicesPool) FromHealthyToJail(id string) {
	p.list.FromHealthyToJail(id)
}

func (p *ServicesPool) FromJailToHealthy(srv service.IService) {
	p.list.FromJailToHealthy(srv)
}

func (p *ServicesPool) RemoveFromJail(srv service.IService) {
	p.list.RemoveFromJail(srv)
}
