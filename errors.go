package pool

import "github.com/gateway-fm/service-pool/service"

// ErrHealthCheck is wrapped error from
// healthcheck with embedded service
type ErrHealthCheck struct {
	err error
	srv *service.IService
}

// Error is throw error as a string
func (e ErrHealthCheck) Error() string {
	panic("implement me")
}
