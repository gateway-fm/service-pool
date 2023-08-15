package discovery

import (
	"github.com/gateway-fm/service-pool/service"
)

// ManualDiscovery is "manual" implementation
// of Near addresses discovery
type ManualDiscovery struct {
	addresses []string
	transport TransportProtocol
	opts      *DiscoveryOpts
}

// NewManualDiscovery create new manual
// NodesDiscovery with given addresses
func NewManualDiscovery(transport TransportProtocol, opts *DiscoveryOpts, addrs ...string) (IServiceDiscovery, error) {
	opts = NilDiscoveryOptions()
	return &ManualDiscovery{addresses: addrs, opts: opts, transport: transport}, nil
}

// Discover is discover and return list of the active
// blockchain addresses for requested networks
func (d *ManualDiscovery) Discover(string) (nodes []service.IService, err error) {
	for _, n := range d.addresses {
		nodes = append(nodes, service.NewService(d.transport.FormatAddress(n), "", nil))
	}
	return
}
