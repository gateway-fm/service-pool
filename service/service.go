package service

import (
	"crypto/sha256"
	"encoding/hex"
)

type IService interface {
	// HealthCheck check service health by
	// sending status request
	HealthCheck() error

	// Status return service current status
	Status() Status

	// ID return service unique ID
	ID() string

	// Address return service address
	Address() string

	// NodeName return prover name from discovery
	NodeName() string

	Tags() map[string]struct{}

	Close() error

	Load() float32 // rating between [0.0, 1.0]

	SetLoad(float32)

	ProverLoad() *ProverLoad

	SetProverLoad(*ProverLoad)
}

// TODO split address field to host and port

// BaseService represent basic service
// model implementation
type BaseService struct {
	id       string              // service unique id - sha256(address)
	status   Status              // service current status
	address  string              // service address to connect
	nodeName string              // prover name from discovery
	tags     map[string]struct{} // service tags
	load     float32             // rating between [0.0, 1.0]

	proverLoad *ProverLoad //specific prover load
}

// NewService create new BaseService with address and discovery
func NewService(address, nodeName string, tags map[string]struct{}, load float32) IService {
	return &BaseService{
		id:       GenerateServiceID(address),
		status:   StatusUnHealthy,
		address:  address,
		nodeName: nodeName,
		tags:     tags,
		load:     load,
	}
}

// HealthCheck check service health by
// sending status request
func (n *BaseService) HealthCheck() error {
	// TODO implement basic http or tcp healthchecks
	return nil
}

// Status return BaseService current status
func (n *BaseService) Status() Status {
	return n.status
}

// ID return service unique ID
func (n *BaseService) ID() string {
	return n.id
}

// Address return service address
func (n *BaseService) Address() string {
	return n.address
}

// NodeName return prover name from discovery
func (n *BaseService) NodeName() string {
	return n.nodeName
}

func (n *BaseService) SetStatus(status Status) {
	n.status = status
}

func (n *BaseService) Load() float32 {
	return n.load
}

func (n *BaseService) SetLoad(load float32) {
	n.load = load
}

func (n *BaseService) Tags() map[string]struct{} {
	return n.tags
}

func (n *BaseService) Close() error {
	return nil
}

// GenerateServiceID create BaseService unique id by
// hashing given address string
func GenerateServiceID(addr string) string {
	h := sha256.New()
	h.Write([]byte(addr))
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

type ProverLoad struct {
	// status
	ProverStatus GetStatusResponse_Status
	// number or tasks at hand
	TasksQueue int
	// number of cores
	NumberCores int
	// start time of the task running
	CurrentComputingStartTime int64

	// not used yet
	//TotalMemory int64
	//FreeMemory int64
}

func (n *BaseService) ProverLoad() *ProverLoad {
	return n.proverLoad
}

func (n *BaseService) SetProverLoad(load *ProverLoad) {
	n.proverLoad = load
}

// Copied from the aggregator proto spec
type GetStatusResponse_Status int32

const (
	GetStatusResponse_STATUS_UNSPECIFIED GetStatusResponse_Status = 0
	GetStatusResponse_STATUS_BOOTING     GetStatusResponse_Status = 1
	GetStatusResponse_STATUS_COMPUTING   GetStatusResponse_Status = 2
	GetStatusResponse_STATUS_IDLE        GetStatusResponse_Status = 3
	GetStatusResponse_STATUS_HALT        GetStatusResponse_Status = 4
)
