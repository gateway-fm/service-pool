package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gateway-fm/scriptorium/logger"

	"github.com/gateway-fm/prover-pool-lib/pkg/utils"
	"github.com/gateway-fm/prover-pool-lib/service"
)

// IServicesList is generic interface for services list
type IServicesList interface {
	// Healthy return slice of all healthy services
	Healthy() []service.IService

	// Unhealthy return slice of all unHealthy services
	Unhealthy() []service.IService

	// Next returns next healthy service
	// to take a connection
	Next() service.IService

	NextLeastLoaded(tag string) service.IService

	// Add service to the list
	Add(srv service.IService)

	// IsServiceExists check is given service is
	// already in list (healthy or jail)
	IsServiceExists(srv service.IService) bool

	// HealthChecks pings the healthy services
	// and update the statuses
	HealthChecks()

	// HealthChecksLoop spawn healthchecks for
	// all healthy services periodically
	HealthChecksLoop()

	// TryUpService recursively try to up service
	TryUpService(srv service.IService, try int)

	// FromHealthyToJail move Unhealthy service
	// from Healthy slice to Jail map
	FromHealthyToJail(id string)

	// FromJailToHealthy move Healthy service
	// from Jail map to Healthy slice
	FromJailToHealthy(srv service.IService)

	// RemoveFromJail remove given
	// service from jail map
	RemoveFromJail(srv service.IService)

	// RemoveFromHealthyByIndex removes
	// service from healthy slice by given srv index in that slice
	RemoveFromHealthyByIndex(i int)

	// Close Stop service list
	Close()

	// Shuffle randomly shuffles list
	Shuffle()

	// CountAll returns sum of num healthy and num jailed services together
	CountAll() int

	// Jailed returns a copy of jail map
	Jailed() map[string]service.IService

	ModifyHealthy(modifier func(srv service.IService))
}

// ServicesList is service list implementation that
// manage healthchecks, jail and try up mechanics
type ServicesList struct {
	serviceName string

	current uint64

	healthy []service.IService

	jail map[string]service.IService

	//muMain sync.Mutex
	//muJail sync.Mutex

	mu sync.RWMutex

	TryUpTries    int
	CheckInterval time.Duration
	TryUpInterval time.Duration

	Stop chan struct{}
}

// ServicesListOpts is options that needs
// to configure ServicesList instance
type ServicesListOpts struct {
	TryUpTries     int           // number of attempts to try up service from jail (0 for infinity tries)
	TryUpInterval  time.Duration // interval for try up service from jail
	ChecksInterval time.Duration // healthchecks interval
}

// NewServicesList create new ServiceList instance
// with given configuration
func NewServicesList(serviceName string, opts *ServicesListOpts) IServicesList {
	return &ServicesList{
		serviceName:   serviceName,
		jail:          make(map[string]service.IService),
		TryUpTries:    opts.TryUpTries,
		CheckInterval: opts.ChecksInterval,
		TryUpInterval: opts.TryUpInterval,
		Stop:          make(chan struct{}),
	}
}

// Healthy return slice of all healthy services
func (l *ServicesList) Healthy() []service.IService {
	defer l.mu.RUnlock()
	l.mu.RLock()

	var healthy []service.IService
	healthy = append(healthy, l.healthy...)

	return healthy
}

// Unhealthy return slice of all unHealthy services
func (l *ServicesList) Unhealthy() []service.IService {
	defer l.mu.RUnlock()
	l.mu.RLock()

	var unHealthy []service.IService

	for _, s := range l.jail {
		unHealthy = append(unHealthy, s)
	}

	return unHealthy
}

// Next returns next healthy service
// to take a connection
func (l *ServicesList) Next() service.IService {
	defer l.mu.Unlock()
	l.mu.Lock()

	if len(l.healthy) == 0 {
		logger.Log().Info(fmt.Sprintf("list name %s no healthy services are present during list's Next() call", l.serviceName))
		return nil
	}

	next := l.nextIndex()
	length := len(l.healthy) + next
	for i := next; i < length; i++ {
		idx := i % len(l.healthy)
		if l.healthy[idx].Status() == service.StatusHealthy {
			if i != next {
				atomic.StoreUint64(&l.current, uint64(idx))
			}
			return l.healthy[idx]
		}
	}

	logger.Log().Info(fmt.Sprintf("list name %s no healthy services are present after forloop during list's Next() call", l.serviceName))
	return nil
}

func (l *ServicesList) NextLeastLoaded(tag string) service.IService {
	defer l.mu.Unlock()
	l.mu.Lock()

	if len(l.healthy) == 0 {
		logger.Log().Info(fmt.Sprintf("list name %s no healthy services are present during list's Next() call", l.serviceName))
		return nil
	}

	var leastLoadedSrv service.IService
	minLoad := float32(1.01)

	for _, srv := range l.healthy {
		_, isTagPresent := srv.Tags()[tag]
		if !isTagPresent {
			continue
		}

		load := srv.Load()
		if load < minLoad {
			leastLoadedSrv = srv
			minLoad = load
		}
	}

	return leastLoadedSrv
}

// Add service to the list
func (l *ServicesList) Add(srv service.IService) {
	if l.IsServiceExists(srv) {
		logger.Log().Info(fmt.Sprintf("list name %s service already exists during Add, service with id %s with nodeName %s", l.serviceName, srv.ID(), srv.NodeName()))
		return
	}

	l.mu.Lock()

	if err := srv.HealthCheck(); err != nil {
		l.jail[srv.ID()] = srv
		logger.Log().Warn(fmt.Sprintf("list name %s service with id %s with nodeName %s can't be added to healthy due to healthcheck error: %s", l.serviceName, srv.ID(), srv.NodeName(), err.Error()))

		go l.TryUpService(srv, 0)

		l.mu.Unlock()
		return
	}

	l.healthy = append(l.healthy, srv)
	logger.Log().Info(fmt.Sprintf("list name %s service with id %s with nodeName %s with address %s added to list", l.serviceName, srv.ID(), srv.NodeName(), srv.Address()))
	l.mu.Unlock()
}

// IsServiceExists check is given service is
// already in list (healthy or jail)
func (l *ServicesList) IsServiceExists(srv service.IService) bool {
	defer l.mu.RUnlock()
	l.mu.RLock()

	if l.isServiceInJail(srv) {
		return true
	}

	if l.isServiceInHealthy(srv) {
		return true
	}

	return false
}

// HealthChecks pings the healthy services
// and update the status
func (l *ServicesList) HealthChecks() {
	for _, srv := range l.Healthy() {
		if srv == nil {
			logger.Log().Info(fmt.Sprintf("list name %s service is nil during hc loop, skipping the healthcheck for it", l.serviceName))
			continue
		}

		// TODO need to implement advanced logging level

		if err := srv.HealthCheck(); err != nil {
			logger.Log().Warn(fmt.Errorf("healthcheck error on list with name %s, service with id %s with nodeName %s: %w", l.serviceName, srv.ID(), srv.NodeName(), err).Error())

			go func(service service.IService) {
				l.FromHealthyToJail(service.ID())
				logger.Log().Warn(fmt.Sprintf("%s service %s added to jail", l.serviceName, service.ID()))
				l.TryUpService(service, 0)
			}(srv)

			continue
		}
	}
}

// HealthChecksLoop spawn healthchecks for
// all healthy periodically
func (l *ServicesList) HealthChecksLoop() {
	logger.Log().Info("start healthchecks loop")

	for {
		select {
		case <-l.Stop:
			logger.Log().Warn("stop healthchecks loop")
			return
		default:
			l.HealthChecks()
			Sleep(l.CheckInterval, l.Stop)
		}
	}
}

// TryUpService recursively try to up service
func (l *ServicesList) TryUpService(srv service.IService, try int) {
	if l.TryUpTries != 0 && try >= l.TryUpTries {
		logger.Log().Warn(fmt.Sprintf("list name %s maximum %d try to Up service with id %s with nodeName %s reached.... service will remove from service list", l.serviceName, l.TryUpTries, srv.ID(), srv.NodeName()))
		l.RemoveFromJail(srv)
		return
	}

	logger.Log().Info(fmt.Sprintf("list name %s %d try to up service with id %s with address %s with nodeName %s", l.serviceName, try, srv.ID(), srv.Address(), srv.NodeName()))

	if err := srv.HealthCheck(); err != nil {
		logger.Log().Warn(fmt.Errorf("list name %s service with id %s with nodeName %s healthcheck error: %w", l.serviceName, srv.ID(), srv.NodeName(), err).Error())

		Sleep(l.TryUpInterval, l.Stop)
		l.TryUpService(srv, try+1)
		return
	}

	logger.Log().Info(fmt.Sprintf("list name %s service with id %s with nodeName %s is alive!", l.serviceName, srv.ID(), srv.NodeName()))

	l.FromJailToHealthy(srv)
}

// FromHealthyToJail move Unhealthy service
// from Healthy slice to Jail map
func (l *ServicesList) FromHealthyToJail(id string) {
	defer l.mu.Unlock()
	l.mu.Lock()

	var (
		index = -1
		srv   service.IService
	)

	for i, s := range l.healthy {
		if s.ID() == id {
			index = i
			srv = s
			break
		}
	}

	if index == -1 {
		logger.Log().Warn(fmt.Sprintf("list name %s service with id %s is not found in healthy during FromHealthyToJail", l.serviceName, id))
		return
	}

	l.healthy = deleteFromSlice(l.healthy, index)
	l.jail[srv.ID()] = srv

	logger.Log().Info(fmt.Sprintf("list name %s service with id %s is moved from healthy to jail", l.serviceName, id))
}

// FromJailToHealthy move Healthy service
// from Jail map to Healthy slice
func (l *ServicesList) FromJailToHealthy(srv service.IService) {
	l.mu.Lock()
	delete(l.jail, srv.ID())
	l.mu.Unlock()

	l.Add(srv)

	logger.Log().Info(fmt.Sprintf("list name %s service with id %s with nodeName %s is moved from jail to healthy", l.serviceName, srv.ID(), srv.NodeName()))
}

func (l *ServicesList) RemoveFromHealthyByIndex(i int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	srv := l.healthy[i]
	logger.Log().Info(fmt.Sprintf("list name %s service with id %s with nodeName %s is about to be removed from healthy by index", l.serviceName, srv.ID(), srv.NodeName()))

	if err := srv.Close(); err != nil {
		logger.Log().Warn(fmt.Errorf("unexpected error during service Close(): %w", err).Error())
	}

	l.healthy = deleteFromSlice(l.healthy, i)
}

// RemoveFromJail remove given
// service from jail map
func (l *ServicesList) RemoveFromJail(srv service.IService) {
	defer l.mu.Unlock()
	l.mu.Lock()

	logger.Log().Info(fmt.Sprintf("list name %s service with id %s with nodeName %s is about to be removed from jail", l.serviceName, srv.ID(), srv.NodeName()))

	if err := srv.Close(); err != nil {
		logger.Log().Warn(fmt.Errorf("unexpected error during service Close(): %w", err).Error())
	}

	delete(l.jail, srv.ID())
}

// Close Stop service list handling
func (l *ServicesList) Close() {
	close(l.Stop)
}

func (l *ServicesList) Shuffle() {
	defer l.mu.Unlock()
	l.mu.Lock()

	length := len(l.healthy)
	if length == 0 {
		return
	}

	utils.ShuffleSlice(l.healthy)

	newCurrent := utils.RandomUint64(length)
	atomic.StoreUint64(&l.current, newCurrent)
}

func (l *ServicesList) CountAll() int {
	// no mutex lock here since l.Healthy() has its own lock
	numHealthy := len(l.Healthy())

	// len is not concurrency safe
	l.mu.RLock()
	defer l.mu.RUnlock()

	return numHealthy + len(l.jail)
}

func (l *ServicesList) Jailed() map[string]service.IService {
	defer l.mu.RUnlock()
	l.mu.RLock()

	// make copy of jailed map
	jailed := make(map[string]service.IService)
	for k, v := range l.jail {
		jailed[k] = v
	}

	return jailed
}

func (l *ServicesList) ModifyHealthy(modifier func(srv service.IService)) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, srv := range l.healthy {
		modifier(srv)
	}
}

// isServiceInJail check if service exist in jail
func (l *ServicesList) isServiceInJail(srv service.IService) bool {
	if srv == nil {
		logger.Log().Warn("nil srv provided when calling isServiceInJail")
		return false
	}

	if _, ok := l.jail[srv.ID()]; ok {
		return true
	}

	return false
}

// isServiceInHealthy check if service exist in
// healthy slice
func (l *ServicesList) isServiceInHealthy(srv service.IService) bool {
	if srv == nil {
		logger.Log().Warn("nil srv provided when calling isServiceInHealthy")
		return false
	}

	for _, oldService := range l.healthy {
		if oldService == nil {
			logger.Log().Warn("nil oldService in healthy slice of ServicesList")
			continue
		}

		if srv.ID() == oldService.ID() {
			return true
		}
	}
	return false
}

// nextIndex atomically increase the
// counter and return an index
func (l *ServicesList) nextIndex() int {
	return int(atomic.AddUint64(&l.current, uint64(1)) % uint64(len(l.healthy)))
}
