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

package credentials

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// service is the keyring service name. Per-host accounts are encoded as
// `<server>::<account>` since OS keychains key on (service, account) pairs.
const service = "veans"

// KeyringBackend persists tokens in the OS keychain (macOS Keychain,
// Windows Credential Manager, libsecret on Linux). On systems without a
// usable keychain (e.g. headless CI containers), Get/Set return errors that
// the chain treats as NotFound, allowing the file backend to take over.
type KeyringBackend struct{}

func NewKeyringBackend() *KeyringBackend { return &KeyringBackend{} }
func (*KeyringBackend) Name() string     { return "keyring" }

func (*KeyringBackend) Get(server, account string) (string, error) {
	tok, err := keyring.Get(service, key(server, account))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		// Treat any keyring backend error (no daemon, etc) as NotFound so
		// the chain falls through to the file backend transparently.
		return "", ErrNotFound
	}
	return tok, nil
}

func (*KeyringBackend) Set(server, account, token string) error {
	if err := keyring.Set(service, key(server, account), token); err != nil {
		return err
	}
	return nil
}

func (*KeyringBackend) Delete(server, account string) error {
	if err := keyring.Delete(service, key(server, account)); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func key(server, account string) string {
	return server + "::" + account
}
