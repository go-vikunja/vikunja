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

package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testInstanceID = "00000000-0000-0000-0000-000000000000"

type testSigner struct {
	priv ed25519.PrivateKey
	kid  string
}

func newTestSigner(seedByte byte) testSigner {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = seedByte
	}
	priv := ed25519.NewKeyFromSeed(seed)
	return testSigner{priv: priv, kid: keyID(priv.Public().(ed25519.PublicKey))}
}

// trustSigners swaps the embedded public keys for the given signers for the duration of the test.
func trustSigners(t *testing.T, signers ...testSigner) {
	prev := offlinePublicKeys
	offlinePublicKeys = make(map[string]ed25519.PublicKey, len(signers))
	for _, s := range signers {
		offlinePublicKeys[s.kid] = s.priv.Public().(ed25519.PublicKey)
	}
	t.Cleanup(func() { offlinePublicKeys = prev })
}

type tokenOpts struct {
	issuer   string
	audience string
	iat      time.Time
	exp      time.Time
	features []string
	kid      string
}

func validTokenOpts(now time.Time) tokenOpts {
	return tokenOpts{
		issuer:   offlineIssuer,
		audience: testInstanceID,
		iat:      now.Add(-time.Hour),
		exp:      now.Add(365 * 24 * time.Hour),
		features: []string{"admin_panel", "audit_logs"},
	}
}

func (s testSigner) sign(t *testing.T, o tokenOpts) string {
	t.Helper()
	token := paseto.NewToken()
	token.SetIssuer(o.issuer)
	token.SetSubject("license-123")
	token.SetAudience(o.audience)
	token.SetJti("token-1")
	token.SetIssuedAt(o.iat)
	token.SetExpiration(o.exp)
	token.SetString("customer", "ACME Inc.")
	require.NoError(t, token.Set("features", o.features))
	require.NoError(t, token.Set("max_users", 42))

	kid := s.kid
	if o.kid != "" {
		kid = o.kid
	}
	token.SetFooter([]byte(`{"kid":"` + kid + `"}`))

	key, err := paseto.NewV4AsymmetricSecretKeyFromEd25519(s.priv)
	require.NoError(t, err)
	return token.V4Sign(key, nil)
}

func TestVerifyOfflineToken(t *testing.T) {
	now := time.Now()
	signer := newTestSigner(1)
	other := newTestSigner(2)
	trustSigners(t, signer)

	t.Run("valid", func(t *testing.T) {
		claims, err := verifyOfflineToken(signer.sign(t, validTokenOpts(now)), testInstanceID, now)
		require.NoError(t, err)
		assert.Equal(t, "ACME Inc.", claims.Customer)
		assert.Equal(t, int64(42), claims.MaxUsers)
		assert.Equal(t, []Feature{FeatureAdminPanel, FeatureAuditLogs}, claims.Features)
	})

	t.Run("unknown features are ignored", func(t *testing.T) {
		o := validTokenOpts(now)
		o.features = []string{"admin_panel", "from_the_future"}
		claims, err := verifyOfflineToken(signer.sign(t, o), testInstanceID, now)
		require.NoError(t, err)
		assert.Equal(t, []Feature{FeatureAdminPanel, FeatureUnknown}, claims.Features)
	})

	t.Run("bad signature", func(t *testing.T) {
		o := validTokenOpts(now)
		o.kid = signer.kid
		_, err := verifyOfflineToken(other.sign(t, o), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineBadSignature)
	})

	t.Run("tampered payload", func(t *testing.T) {
		token := signer.sign(t, validTokenOpts(now))
		parts := strings.Split(token, ".")
		body := []byte(parts[2])
		body[0] = map[bool]byte{true: 'B', false: 'A'}[body[0] == 'A']
		parts[2] = string(body)
		_, err := verifyOfflineToken(strings.Join(parts, "."), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineBadSignature)
	})

	t.Run("unknown kid", func(t *testing.T) {
		_, err := verifyOfflineToken(other.sign(t, validTokenOpts(now)), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineUnknownKey)
	})

	t.Run("malformed footer", func(t *testing.T) {
		_, err := verifyOfflineToken("v4.public.AAAA.bm90LWpzb24", testInstanceID, now)
		require.ErrorIs(t, err, errOfflineTokenMalformed)
	})

	t.Run("wrong issuer", func(t *testing.T) {
		o := validTokenOpts(now)
		o.issuer = "evil.example.com"
		_, err := verifyOfflineToken(signer.sign(t, o), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineWrongIssuer)
	})

	t.Run("wrong instance", func(t *testing.T) {
		o := validTokenOpts(now)
		o.audience = "11111111-1111-1111-1111-111111111111"
		_, err := verifyOfflineToken(signer.sign(t, o), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineWrongInstance)
		assert.Contains(t, err.Error(), "token is for 11111111-1111-1111-1111-111111111111, this instance is "+testInstanceID)
	})

	t.Run("expired", func(t *testing.T) {
		o := validTokenOpts(now)
		o.exp = now.Add(-time.Minute)
		_, err := verifyOfflineToken(signer.sign(t, o), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineExpired)
	})

	t.Run("issued in the future beyond clock skew", func(t *testing.T) {
		o := validTokenOpts(now)
		o.iat = now.Add(25 * time.Hour)
		_, err := verifyOfflineToken(signer.sign(t, o), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineNotYetValid)
	})

	t.Run("issued in the future within clock skew", func(t *testing.T) {
		o := validTokenOpts(now)
		o.iat = now.Add(23 * time.Hour)
		_, err := verifyOfflineToken(signer.sign(t, o), testInstanceID, now)
		require.NoError(t, err)
	})
}

func TestMustDecodePublicKeys(t *testing.T) {
	signer := newTestSigner(1)
	pub := base64.StdEncoding.EncodeToString(signer.priv.Public().(ed25519.PublicKey))

	keys := mustDecodePublicKeys(map[string]string{signer.kid: pub})
	assert.Equal(t, signer.priv.Public(), keys[signer.kid])

	assert.Panics(t, func() { mustDecodePublicKeys(map[string]string{"0000000000000000": pub}) })
	assert.Panics(t, func() { mustDecodePublicKeys(map[string]string{signer.kid: "not base64!"}) })
}

func initOfflineForTest(t *testing.T, token string) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	config.LicenseKey.Set(token)
	t.Cleanup(func() {
		Shutdown()
		stopCh = nil
		config.LicenseKey.Set("")
		ResetForTests()
	})
	Init()
}

func TestInitOffline(t *testing.T) {
	signer := newTestSigner(1)
	trustSigners(t, signer)

	t.Run("valid token enables features", func(t *testing.T) {
		initOfflineForTest(t, signer.sign(t, validTokenOpts(time.Now())))

		assert.True(t, IsFeatureEnabled(FeatureAdminPanel))
		assert.True(t, IsFeatureEnabled(FeatureAuditLogs))
		assert.False(t, IsFeatureEnabled(FeatureTimeTracking))
		assert.Equal(t, int64(42), CurrentInfo().MaxUsers)
	})

	t.Run("token for another instance stays free", func(t *testing.T) {
		o := validTokenOpts(time.Now())
		o.audience = "11111111-1111-1111-1111-111111111111"
		initOfflineForTest(t, signer.sign(t, o))

		assert.False(t, CurrentInfo().Licensed)
		assert.False(t, IsFeatureEnabled(FeatureAdminPanel))
	})

	t.Run("reload re-verifies the token instead of reading the cache", func(t *testing.T) {
		initOfflineForTest(t, signer.sign(t, validTokenOpts(time.Now())))
		ResetForTests()

		require.NoError(t, ReloadFromCache())
		assert.True(t, IsFeatureEnabled(FeatureAdminPanel))
	})
}
