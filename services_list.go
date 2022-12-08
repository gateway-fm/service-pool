package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gateway-fm/scriptorium/logger"

	"github.com/gateway-fm/service-pool/pkg/utils"
	"github.com/gateway-fm/service-pool/service"
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

	// Close Stop service list
	Close()

	// Shuffle randomly shuffles list
	Shuffle()
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
	return nil
}

// Add service to the list
func (l *ServicesList) Add(srv service.IService) {
	defer l.mu.Unlock()
	l.mu.Lock()

	if err := srv.HealthCheck(); err != nil {
		l.jail[srv.ID()] = srv
		logger.Log().Warn(fmt.Sprintf("can't be added to healthy pool: %s", err.Error()))
		go l.TryUpService(srv, 0)
		return
	}

	l.healthy = append(l.healthy, srv)
	logger.Log().Info(fmt.Sprintf("%s service %s with address %s added to list", l.serviceName, srv.ID(), srv.Address()))
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
			continue
		}

		// TODO need to implement advanced logging level

		//logger.Log().Info(fmt.Sprintf("checking %s service %s...", l.serviceName, srv.ID()))

		if err := srv.HealthCheck(); err != nil {
			logger.Log().Warn(fmt.Errorf("healthcheck error on %s service %s: %w", l.serviceName, srv.ID(), err).Error())

			go func(service service.IService) {
				l.FromHealthyToJail(service.ID())
				logger.Log().Warn(fmt.Sprintf("%s service %s added to jail", l.serviceName, service.ID()))
				l.TryUpService(service, 0)
			}(srv)

			continue
		}

		//logger.Log().Info(fmt.Sprintf("%s service %s on %s is healthy", l.serviceName, srv.ID(), srv.Address()))
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
		logger.Log().Warn(fmt.Sprintf("maximum %d try to Up %s service %s reached.... service will remove from service list", l.TryUpTries, l.serviceName, srv.ID()))
		l.RemoveFromJail(srv)
		return
	}

	logger.Log().Info(fmt.Sprintf("%d try to up %s service %s on %s", try, l.serviceName, srv.ID(), srv.Address()))

	if err := srv.HealthCheck(); err != nil {
		logger.Log().Warn(fmt.Errorf("service %s healthcheck error: %w", srv.ID(), err).Error())

		Sleep(l.TryUpInterval, l.Stop)
		l.TryUpService(srv, try+1)
		return
	}

	logger.Log().Info(fmt.Sprintf("service %s is alive!", srv.ID()))

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
		return
	}

	l.healthy = deleteFromSlice(l.healthy, index)

	l.jail[srv.ID()] = srv
}

// FromJailToHealthy move Healthy service
// from Jail map to Healthy slice
func (l *ServicesList) FromJailToHealthy(srv service.IService) {
	l.mu.Lock()
	delete(l.jail, srv.ID())
	l.mu.Unlock()

	l.Add(srv)
}

// RemoveFromJail remove given
// service from jail map
func (l *ServicesList) RemoveFromJail(srv service.IService) {
	defer l.mu.Unlock()
	l.mu.Lock()

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
