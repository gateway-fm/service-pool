package discovery

import "fmt"

// ErrUnsupportedDriver is error when
// discovery driver is unsupported
type ErrUnsupportedDriver struct {
	driver string
}

// Error is throw error as a string
func (e ErrUnsupportedDriver) Error() string {
	return fmt.Sprintf("unsupported discovery driver %q", e.driver)
}

// ErrServiceNotFound is error when
// given service can't be found
type ErrServiceNotFound struct {
	service string
}

// Error is throw error as a string
func (e ErrServiceNotFound) Error() string {
	return fmt.Sprintf("service %v not found", e.service)
}

// ErrInvalidArgumentsLength  is error when
// given arguments length for new discovery
// is invalid
type ErrInvalidArgumentsLength struct {
	length int
	driver Driver
}

// Error is throw error as a string
func (e ErrInvalidArgumentsLength) Error() string {
	return fmt.Sprintf("%d is invalid argument lenght to create new %s discovery", e.length, e.driver.String())
}
