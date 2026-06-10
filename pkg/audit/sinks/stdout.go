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
	"io"
	"os"
	"sync"
)

// Stdout writes each entry as one line to standard output.
type Stdout struct {
	mu sync.Mutex
	// out exists so tests can capture the output.
	out io.Writer
}

func NewStdout() *Stdout {
	return &Stdout{out: os.Stdout}
}

func (s *Stdout) Write(line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.out.Write(line); err != nil {
		return err
	}
	_, err := s.out.Write([]byte{'\n'})
	return err
}
