// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package sinks

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Hand-rolled RFC 5424 instead of log/syslog: the stdlib package only emits
// the older RFC 3164 format and does not build on Windows.
type Syslog struct {
	network  string
	address  string
	facility int
	hostname string
	procid   string

	mu   sync.Mutex
	conn net.Conn
}

var syslogFacilities = map[string]int{
	"kern": 0, "user": 1, "mail": 2, "daemon": 3, "auth": 4, "syslog": 5,
	"lpr": 6, "news": 7, "uucp": 8, "cron": 9, "authpriv": 10, "ftp": 11,
	"local0": 16, "local1": 17, "local2": 18, "local3": 19,
	"local4": 20, "local5": 21, "local6": 22, "local7": 23,
}

// NewSyslog creates a syslog sink. The address has the form
// udp://host:port or tcp://host:port; the scheme defaults to udp.
func NewSyslog(address, facility string) (*Syslog, error) {
	if address == "" {
		return nil, fmt.Errorf("syslog forwarder requires an address")
	}
	if !strings.Contains(address, "://") {
		address = "udp://" + address
	}
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("invalid syslog address %q: %w", address, err)
	}
	if u.Scheme != "udp" && u.Scheme != "tcp" {
		return nil, fmt.Errorf("unsupported syslog scheme %q, must be udp or tcp", u.Scheme)
	}

	if facility == "" {
		facility = "local0"
	}
	facilityCode, ok := syslogFacilities[strings.ToLower(facility)]
	if !ok {
		return nil, fmt.Errorf("unknown syslog facility %q", facility)
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "-"
	}

	return &Syslog{
		network:  u.Scheme,
		address:  u.Host,
		facility: facilityCode,
		hostname: hostname,
		procid:   fmt.Sprintf("%d", os.Getpid()),
	}, nil
}

func (s *Syslog) Write(line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		conn, err := dialer.DialContext(context.Background(), s.network, s.address)
		if err != nil {
			return fmt.Errorf("could not connect to syslog at %s://%s: %w", s.network, s.address, err)
		}
		s.conn = conn
	}

	pri := s.facility*8 + 6 // severity: informational
	frame := fmt.Sprintf("<%d>1 %s %s vikunja %s audit - %s",
		pri, time.Now().UTC().Format(time.RFC3339Nano), s.hostname, s.procid, line)
	if s.network == "tcp" {
		frame += "\n" // RFC 6587 non-transparent framing
	}

	if _, err := s.conn.Write([]byte(frame)); err != nil {
		// Drop the connection so the next write redials.
		_ = s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
}
