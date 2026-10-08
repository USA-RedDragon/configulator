package impl

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Duration decodes a time.Duration from time.ParseDuration syntax ("30s").
type Duration struct {
	v  time.Duration
	ok bool
}

func (s *Duration) UnmarshalText(b []byte) error {
	d, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	s.v, s.ok = d, true
	return nil
}

// Value returns the decoded value and whether UnmarshalText actually ran.
func (s *Duration) Value() (time.Duration, bool) { return s.v, s.ok }

// IPNet decodes a net.IPNet from CIDR notation ("10.0.0.0/8").
type IPNet struct {
	v  net.IPNet
	ok bool
}

func (s *IPNet) UnmarshalText(b []byte) error {
	_, n, err := net.ParseCIDR(string(b))
	if err != nil {
		return err
	}
	s.v, s.ok = *n, true
	return nil
}

func (s *IPNet) Value() (net.IPNet, bool) { return s.v, s.ok }

// FileMode decodes an os.FileMode from octal ("0644").
type FileMode struct {
	v  os.FileMode
	ok bool
}

func (s *FileMode) UnmarshalText(b []byte) error {
	v, err := strconv.ParseUint(string(b), 8, 32)
	if err != nil {
		return err
	}
	s.v, s.ok = os.FileMode(v), true
	return nil
}

func (s *FileMode) Value() (os.FileMode, bool) { return s.v, s.ok }

// Location decodes a *time.Location by IANA name ("America/New_York").
type Location struct {
	v  *time.Location
	ok bool
}

func (s *Location) UnmarshalText(b []byte) error {
	loc, err := time.LoadLocation(string(b))
	if err != nil {
		return err
	}
	s.v, s.ok = loc, true
	return nil
}

func (s *Location) Value() (*time.Location, bool) { return s.v, s.ok }

// TCPAddr decodes a net.TCPAddr from "host:port". Hostnames are resolved
// with the OS resolver when decoded.
type TCPAddr struct {
	v  net.TCPAddr
	ok bool
}

func (s *TCPAddr) UnmarshalText(b []byte) error {
	a, err := net.ResolveTCPAddr("tcp", string(b))
	if err != nil {
		return err
	}
	s.v, s.ok = *a, true
	return nil
}

func (s *TCPAddr) Value() (net.TCPAddr, bool) { return s.v, s.ok }

// UDPAddr decodes a net.UDPAddr from "host:port". Hostnames are resolved
// with the OS resolver when decoded.
type UDPAddr struct {
	v  net.UDPAddr
	ok bool
}

func (s *UDPAddr) UnmarshalText(b []byte) error {
	a, err := net.ResolveUDPAddr("udp", string(b))
	if err != nil {
		return err
	}
	s.v, s.ok = *a, true
	return nil
}

func (s *UDPAddr) Value() (net.UDPAddr, bool) { return s.v, s.ok }

// HardwareAddr decodes a net.HardwareAddr ("aa:bb:cc:dd:ee:ff").
type HardwareAddr struct {
	v  net.HardwareAddr
	ok bool
}

func (s *HardwareAddr) UnmarshalText(b []byte) error {
	a, err := net.ParseMAC(string(b))
	if err != nil {
		return err
	}
	s.v, s.ok = a, true
	return nil
}

func (s *HardwareAddr) Value() (net.HardwareAddr, bool) { return s.v, s.ok }

// URL decodes a url.URL with url.Parse.
type URL struct {
	v  url.URL
	ok bool
}

func (s *URL) UnmarshalText(b []byte) error {
	u, err := url.Parse(string(b))
	if err != nil {
		return err
	}
	s.v, s.ok = *u, true
	return nil
}

func (s *URL) Value() (url.URL, bool) { return s.v, s.ok }

// Month decodes a time.Month from an English month name (any case) or a
// number 1-12.
type Month struct {
	v  time.Month
	ok bool
}

func (s *Month) UnmarshalText(b []byte) error {
	t := strings.TrimSpace(string(b))
	if n, err := strconv.Atoi(t); err == nil {
		if n < 1 || n > 12 {
			return fmt.Errorf("month %d out of range 1-12", n)
		}
		s.v, s.ok = time.Month(n), true
		return nil
	}
	for m := time.January; m <= time.December; m++ {
		if strings.EqualFold(t, m.String()) {
			s.v, s.ok = m, true
			return nil
		}
	}
	return fmt.Errorf("unknown month %q", t)
}

func (s *Month) Value() (time.Month, bool) { return s.v, s.ok }

// Complex128 decodes a complex128 from strconv.ParseComplex syntax ("1+2i").
type Complex128 struct {
	v  complex128
	ok bool
}

func (s *Complex128) UnmarshalText(b []byte) error {
	c, err := strconv.ParseComplex(strings.TrimSpace(string(b)), 128)
	if err != nil {
		return err
	}
	s.v, s.ok = c, true
	return nil
}

func (s *Complex128) Value() (complex128, bool) { return s.v, s.ok }

// Complex64 decodes a complex64 from strconv.ParseComplex syntax ("1+2i").
type Complex64 struct {
	v  complex64
	ok bool
}

func (s *Complex64) UnmarshalText(b []byte) error {
	c, err := strconv.ParseComplex(strings.TrimSpace(string(b)), 64)
	if err != nil {
		return err
	}
	s.v, s.ok = complex64(c), true
	return nil
}

func (s *Complex64) Value() (complex64, bool) { return s.v, s.ok }
