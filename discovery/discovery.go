package discovery

import (
	"errors"
	"github.com/gateway-fm/service-pool/service"
)

// IServiceDiscovery is interface that provides
// service Discovery and healthchecks
type IServiceDiscovery interface {
	// Discover service by given name
	Discover(service string) ([]service.IService, error)
}
type DiscoveryOpts struct {
	isOptional   bool
	optionalPath string
}

// Creator is discovery factory function
type Creator func(TransportProtocol, *DiscoveryOpts, ...string) (IServiceDiscovery, error)

// ParseDiscoveryDriver create new addresses discovery
// Creator based on given discovery driver
func ParseDiscoveryDriver(driver Driver) (Creator, error) {
	switch driver {
	case DriverConsul:
		return NewConsulDiscovery, nil
	case DriverManual:
		return NewManualDiscovery, nil
	default:
		return nil, ErrUnsupportedDriver{driver.String()}
	}
}

func NewDiscoveryOpts(isPathOptional bool, optionalPath string) *DiscoveryOpts {
	return &DiscoveryOpts{
		isOptional:   isPathOptional,
		optionalPath: optionalPath,
	}
}

// NilDiscoveryOptions to prevent nil pointers if there are no options
func NilDiscoveryOptions() *DiscoveryOpts {
	return &DiscoveryOpts{}
}

var ErrEmptyOptionalPath = errors.New("optional path is empty")
