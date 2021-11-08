package discovery

import (
	"github.com/gateway-fm/service-pool/service"
)

// ManualDiscovery is "manual" implementation
// of Near addresses discovery
type ManualDiscovery struct {
	addresses []string
}

// NewManualDiscovery create new manual
// NodesDiscovery with given addresses
func NewManualDiscovery(addrs ...string) (IServiceDiscovery, error) {
	return &ManualDiscovery{addresses: addrs}, nil
}

// Discover is discover and return list of the active
// blockchain addresses for requested networks
func (d *ManualDiscovery) Discover(string) (nodes []service.IService, err error) {
	for _, n := range d.addresses {
		nodes = append(nodes, service.NewService(n))
	}
	return
}
