package service

import (
	"fmt"
	"strings"
)

// Status represent available BaseService Statuses
type Status int

const (
	// StatusHealthy is means that service is active
	// and ready for incoming requests
	StatusHealthy Status = iota

	// StatusUnHealthy is mean that service is inactive
	StatusUnHealthy

	// statusUnsupported is unsupported status
	statusUnsupported
)

// serviceStatuses is slice of BaseService
// Statuses string representations
var serviceStatuses = [...]string{
	StatusHealthy:   "healthy",
	StatusUnHealthy: "unhealthy",
}

// String return ServiceStatus enum as a string
func (s Status) String() string {
	return serviceStatuses[s]
}

// ServiceStatusFromString return new ServiceStatus
// enum from given string
func ServiceStatusFromString(s string) (Status, error) {
	for i, r := range serviceStatuses {
		if strings.ToLower(s) == r {
			return Status(i), nil
		}
	}
	return statusUnsupported, fmt.Errorf("invalid service status value %q", s)
}
