package prover

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gateway-fm/prover-pool-lib/prover/client"
	"github.com/gateway-fm/prover-pool-lib/service"
)

// IProver is high-level abstraction that
// provide communications with prover
type IProver interface {
	// Close stop prover instance
	Close() error

	DoRequest(data []byte) ([]byte, error)

	SetStatus(service.Status)

	SetLoad(float32)

	service.IService
}

// Prover is basic implementation for IProver abstraction
type Prover struct {
	status int32
	id     string
	addr   string
	name   string

	healthcheck func(n IProver) error

	client client.INodeClient

	mu sync.Mutex

	tags map[string]struct{}

	load float32 // rating between [0.0, 1.0]
}

type ProverOpts struct {
	Name        string
	Addr        string
	Healthcheck func(n IProver) error
	Tags        map[string]struct{}
}

func NewProver(opts *ProverOpts) (*Prover, error) {
	p := &Prover{
		name:        opts.Name,
		addr:        opts.Addr,
		healthcheck: opts.Healthcheck,
		tags:        opts.Tags,
		status:      int32(service.StatusUnHealthy),
		id:          service.GenerateServiceID(opts.Addr),
	}
	if err := p.initNodeClient(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Prover) DoRequest(data []byte) ([]byte, error) {
	return nil, nil
}

// initNodeClient initialise prover NodeClient instance (ws or http)
func (p *Prover) initNodeClient() (err error) {
	if p.client != nil {
		p.client.Close()
	}

	p.client, err = client.NewHttpClient(p.addr)
	if err != nil {
		return fmt.Errorf("create new prover client: %w", err)
	}

	return nil
}

// HealthCheck check prover health by
// sending status request
func (p *Prover) HealthCheck() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.healthcheck == nil {
		// this case won't happen if we mutate a prover from a service
		// but better to have this to prevent any panics
		return errors.New("nil healthcheck function")
	}

	return p.healthcheck(p)
}

func (p *Prover) Load() float32 {
	return p.load
}

func (p *Prover) SetLoad(load float32) {
	p.load = load
}

// Status return Prover current status
func (p *Prover) Status() service.Status {
	return service.Status(atomic.LoadInt32(&p.status))
}

// SetStatus set new prover status
func (p *Prover) SetStatus(status service.Status) {
	atomic.StoreInt32(&p.status, int32(status))
}

// ID return Prover unique ID
func (p *Prover) ID() string {
	return p.id
}

// Address return Prover address
func (p *Prover) Address() string {
	return p.addr
}

func (p *Prover) Tags() map[string]struct{} {
	return p.tags
}

// Close all prover connections
func (p *Prover) Close() error {
	p.client.Close()
	return nil
}

func (p *Prover) NodeName() string {
	return p.name
}
