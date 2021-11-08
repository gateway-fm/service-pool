package discovery

import (
	"strings"
)

// Driver represent available Service Statuses
type Driver int

const (
	// DriverConsul is consul-driven service discovery
	DriverConsul Driver = iota

	// DriverManual is "manual" discovery
	// based on given array of addresses address
	DriverManual

	// driverUnsupported is unsupported status
	driverUnsupported
)

// Drivers are slice of Service Statuses string representations
var Drivers = [...]string{
	DriverConsul: "consul",
	DriverManual: "manual",
}

// String return Driver enum as a string
func (s Driver) String() string {
	return Drivers[s]
}

// DriverFromString return new Driver enum from given
// string
func DriverFromString(s string) Driver {
	for i, r := range Drivers {
		if strings.ToLower(s) == r {
			return Driver(i)
		}
	}
	return driverUnsupported
}

// DriverFromStringE return new Driver enum from given
// string or return an error
func DriverFromStringE(s string) (Driver, error) {
	for i, r := range Drivers {
		if strings.ToLower(s) == r {
			return Driver(i), nil
		}
	}
	return driverUnsupported, ErrUnsupportedDriver{s}
}
