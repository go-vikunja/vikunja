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

package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSelfSignedCert(t *testing.T, certFile, keyFile, commonName string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour),
		DNSNames:  []string{commonName},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	keyDer, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	}), 0o600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDer,
	}), 0o600))
}

func servedCommonName(t *testing.T, r *certReloader) string {
	t.Helper()

	cert, err := r.GetCertificate(nil)
	require.NoError(t, err)
	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)
	return parsed.Subject.CommonName
}

func TestValidateTLSFiles(t *testing.T) {
	tests := []struct {
		name        string
		certFile    string
		keyFile     string
		autoTLS     bool
		wantEnabled bool
		wantErr     bool
	}{
		{
			name: "nothing configured",
		},
		{
			name:    "nothing configured with autotls",
			autoTLS: true,
		},
		{
			name:        "cert and key",
			certFile:    "cert.pem",
			keyFile:     "key.pem",
			wantEnabled: true,
		},
		{
			name:     "only cert",
			certFile: "cert.pem",
			wantErr:  true,
		},
		{
			name:    "only key",
			keyFile: "key.pem",
			wantErr: true,
		},
		{
			name:     "cert and key with autotls",
			certFile: "cert.pem",
			keyFile:  "key.pem",
			autoTLS:  true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled, err := validateTLSFiles(tt.certFile, tt.keyFile, tt.autoTLS)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantEnabled, enabled)
		})
	}
}

func TestCertReloader(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "fullchain.pem")
	keyFile := filepath.Join(dir, "privkey.pem")

	t.Run("fails on missing files", func(t *testing.T) {
		_, err := newCertReloader(certFile, keyFile)
		require.Error(t, err)
	})

	writeSelfSignedCert(t, certFile, keyFile, "first.example.com")
	r, err := newCertReloader(certFile, keyFile)
	require.NoError(t, err)
	assert.Equal(t, "first.example.com", servedCommonName(t, r))

	t.Run("picks up a renewed certificate", func(t *testing.T) {
		writeSelfSignedCert(t, certFile, keyFile, "second.example.com")
		require.NoError(t, r.reload())
		assert.Equal(t, "second.example.com", servedCommonName(t, r))
	})

	t.Run("keeps the previous certificate when reloading fails", func(t *testing.T) {
		require.NoError(t, os.WriteFile(certFile, []byte("not a certificate"), 0o600))
		require.Error(t, r.reload())
		assert.Equal(t, "second.example.com", servedCommonName(t, r))
	})
}
