package service

import "fmt"

// ErrUnsupportedStatus is error when
// service status is unsupported
type ErrUnsupportedStatus struct {
	status string
}

// Error is throw error as a string
func (e ErrUnsupportedStatus) Error() string {
	return fmt.Sprintf("unsupported service status %q", e.status)
}
