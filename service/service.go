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

	// NodeName return node name from discovery
	NodeName() string

	Tags() map[string]struct{}

	SetTags(map[string]struct{})

	Close() error
}

// TODO split address field to host and port

// BaseService represent basic service
// model implementation
type BaseService struct {
	id       string              // service unique id - sha256(address)
	status   Status              // service current status
	address  string              // service address to connect
	nodeName string              // node name from discovery
	tags     map[string]struct{} // service tags
}

// NewService create new BaseService with address and discovery
func NewService(address, nodeName string, tags map[string]struct{}) IService {
	return &BaseService{
		id:       generateServiceID(address),
		status:   StatusUnHealthy,
		address:  address,
		nodeName: nodeName,
		tags:     tags,
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

// NodeName return node name from discovery
func (n *BaseService) NodeName() string {
	return n.nodeName
}

func (n *BaseService) SetStatus(status Status) {
	n.status = status
}

func (n *BaseService) Tags() map[string]struct{} {
	return n.tags
}

func (n *BaseService) SetTags(tags map[string]struct{}) {
	n.tags = tags
}

func (n *BaseService) Close() error {
	return nil
}

// generateServiceID create BaseService unique id by
// hashing given address string
func generateServiceID(addr string) string {
	h := sha256.New()
	h.Write([]byte(addr))
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}
