//go:build goexperiment.jsonv2

package configulator

import "github.com/USA-RedDragon/configulator/v2/impl"

// EnvName is impl.EnvName.
//
// Deprecated: only code generated before v2.4.0 calls this. Run go generate.
func EnvName(prefix, sep string, segments ...string) string {
	return impl.EnvName(prefix, sep, segments...)
}

// SplitList is impl.SplitList.
//
// Deprecated: only code generated before v2.4.0 calls this. Run go generate.
func SplitList(s, sep string) []string { return impl.SplitList(s, sep) }

// Duration is impl.Duration.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type Duration = impl.Duration

// IPNet is impl.IPNet.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type IPNet = impl.IPNet

// FileMode is impl.FileMode.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type FileMode = impl.FileMode

// Location is impl.Location.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type Location = impl.Location

// TCPAddr is impl.TCPAddr.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type TCPAddr = impl.TCPAddr

// UDPAddr is impl.UDPAddr.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type UDPAddr = impl.UDPAddr

// HardwareAddr is impl.HardwareAddr.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type HardwareAddr = impl.HardwareAddr

// URL is impl.URL.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type URL = impl.URL

// Month is impl.Month.
//
// Deprecated: only code generated before v2.4.0 uses this. Run go generate.
type Month = impl.Month
