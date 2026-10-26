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

func trustSigners(t *testing.T, signers ...testSigner) {
	prev := offlinePublicKeys
	pubs := make([]string, 0, len(signers))
	for _, s := range signers {
		pubs = append(pubs, base64.StdEncoding.EncodeToString(s.priv.Public().(ed25519.PublicKey)))
	}
	offlinePublicKeys = mustDecodePublicKeys(pubs)
	t.Cleanup(func() { offlinePublicKeys = prev })
}

type tokenOpts struct {
	issuer   string
	audience string
	exp      time.Time
	features []string
	kid      string
}

func validTokenOpts(now time.Time) tokenOpts {
	return tokenOpts{
		issuer:   offlineIssuer,
		audience: testInstanceID,
		exp:      now.Add(365 * 24 * time.Hour),
		features: []string{"admin_panel", "audit_logs"},
	}
}

func (s testSigner) sign(t *testing.T, o tokenOpts) string {
	t.Helper()
	token := paseto.NewToken()
	token.SetIssuer(o.issuer)
	token.SetAudience(o.audience)
	if !o.exp.IsZero() {
		token.SetExpiration(o.exp)
	}
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
	require.Equal(t, "34750f98bd59fcfc", signer.kid)
	other := newTestSigner(2)
	trustSigners(t, signer)

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

	t.Run("unknown kid", func(t *testing.T) {
		_, err := verifyOfflineToken(other.sign(t, validTokenOpts(now)), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineUnknownKey)
	})

	t.Run("malformed footer", func(t *testing.T) {
		token := "v4.public." + base64.RawURLEncoding.EncodeToString(make([]byte, 64)) + "." + base64.RawURLEncoding.EncodeToString([]byte("not-json"))
		_, err := verifyOfflineToken(token, testInstanceID, now)
		require.ErrorIs(t, err, errOfflineTokenMalformed)
		assert.Contains(t, err.Error(), "footer:")
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

	t.Run("missing expiry", func(t *testing.T) {
		o := validTokenOpts(now)
		o.exp = time.Time{}
		_, err := verifyOfflineToken(signer.sign(t, o), testInstanceID, now)
		require.ErrorIs(t, err, errOfflineExpired)
	})
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
		SetForTests([]Feature{FeatureAdminPanel})
		initOfflineForTest(t, signer.sign(t, o))

		assert.False(t, CurrentInfo().Licensed)
		assert.False(t, IsFeatureEnabled(FeatureAdminPanel))
	})
}

func TestOfflineBackgroundLoopDegradesOnExpiry(t *testing.T) {
	SetForTests([]Feature{FeatureAdminPanel})
	t.Cleanup(ResetForTests)

	offlineBackgroundLoop(time.Now().Add(-time.Second), make(chan struct{}))

	assert.False(t, IsFeatureEnabled(FeatureAdminPanel))
}

func TestInstanceID(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	id, err := InstanceID()
	require.NoError(t, err)
	assert.Equal(t, testInstanceID, id)

	s := db.NewSession()
	_, err = s.Where("1 = 1").Delete(&Status{})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	created, err := InstanceID()
	require.NoError(t, err)
	assert.NotEqual(t, testInstanceID, created)
	again, err := InstanceID()
	require.NoError(t, err)
	assert.Equal(t, created, again)
}
