package discovery

import (
	"github.com/gateway-fm/service-pool/service"
)

// IServiceDiscovery is interface that provides
// service Discovery and healthchecks
type IServiceDiscovery interface {
	// Discover service by given name
	Discover(service string) ([]service.IService, error)
}

// Creator is discovery factory function
type Creator func(...string) (IServiceDiscovery, error)

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
