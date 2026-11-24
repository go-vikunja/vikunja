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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/log"

	"aidanwoods.dev/go-paseto"
)

const (
	offlineTokenPrefix  = "v4.public."
	offlineIssuer       = "console.vikunja.io"
	offlineWarnBefore   = 30 * 24 * time.Hour
	offlineWarnInterval = 24 * time.Hour
)

var embeddedPublicKeys = []string{
	"82l81wrQRxRM/t1h7M1lB45y6d5YN3rk9gUuRRbLpZo=",
}

var offlinePublicKeys = mustDecodePublicKeys(embeddedPublicKeys)

var (
	errOfflineTokenMalformed = errors.New("token is malformed")
	errOfflineUnknownKey     = errors.New("token is signed with an unknown key")
	errOfflineBadSignature   = errors.New("token signature is invalid")
	errOfflineWrongIssuer    = errors.New("token has the wrong issuer")
	errOfflineWrongInstance  = errors.New("token is for a different instance")
	errOfflineExpired        = errors.New("token has expired")
)

type offlineClaims struct {
	Issuer    string    `json:"iss"`
	Audience  string    `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	Customer  string    `json:"customer"`
	Features  []Feature `json:"features"`
	MaxUsers  int64     `json:"max_users"`
}

func isOfflineToken(key string) bool {
	return strings.HasPrefix(key, offlineTokenPrefix)
}

func keyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:8])
}

func mustDecodePublicKeys(encoded []string) map[string]paseto.V4AsymmetricPublicKey {
	keys := make(map[string]paseto.V4AsymmetricPublicKey, len(encoded))
	for _, b64 := range encoded {
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			panic(fmt.Sprintf("invalid embedded license public key %q", b64))
		}
		key, err := paseto.NewV4AsymmetricPublicKeyFromBytes(raw)
		if err != nil {
			panic(fmt.Sprintf("invalid embedded license public key %q", b64))
		}
		keys[keyID(raw)] = key
	}
	return keys
}

func verifyOfflineToken(token, instance string, now time.Time) (*offlineClaims, error) {
	parser := paseto.NewParserWithoutExpiryCheck()

	rawFooter, err := parser.UnsafeParseFooter(paseto.V4Public, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errOfflineTokenMalformed, err)
	}
	var footer struct {
		KeyID string `json:"kid"`
	}
	if err := json.Unmarshal(rawFooter, &footer); err != nil {
		return nil, fmt.Errorf("%w: footer: %w", errOfflineTokenMalformed, err)
	}

	key, ok := offlinePublicKeys[footer.KeyID]
	if !ok {
		return nil, fmt.Errorf("%w %q; this Vikunja version may be too old", errOfflineUnknownKey, footer.KeyID)
	}

	parsed, err := parser.ParseV4Public(key, token, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errOfflineBadSignature, err)
	}

	claims := &offlineClaims{}
	if err := json.Unmarshal(parsed.ClaimsJSON(), claims); err != nil {
		return nil, fmt.Errorf("%w: claims: %w", errOfflineTokenMalformed, err)
	}

	switch {
	case claims.Issuer != offlineIssuer:
		return nil, fmt.Errorf("%w %q", errOfflineWrongIssuer, claims.Issuer)
	case claims.Audience != instance:
		return nil, fmt.Errorf("%w: token is for %s, this instance is %s; generate a new token for %s in Console",
			errOfflineWrongInstance, claims.Audience, instance, instance)
	case !now.Before(claims.ExpiresAt):
		return nil, fmt.Errorf("%w on %s", errOfflineExpired, claims.ExpiresAt.Format(time.RFC3339))
	}

	return claims, nil
}

func initOffline(token string) {
	claims, err := verifyOfflineToken(token, instanceID, time.Now())
	if err != nil {
		degradeToFree(fmt.Sprintf("Offline license is invalid: %s.", err))
		return
	}

	applyResponse(&Response{
		Valid:     true,
		Features:  claims.Features,
		MaxUsers:  claims.MaxUsers,
		ExpiresAt: claims.ExpiresAt,
	})
	log.Infof("Offline license for %s valid until %s.", claims.Customer, claims.ExpiresAt.Format(time.RFC3339))

	stopCh = make(chan struct{})
	go offlineBackgroundLoop(claims.ExpiresAt, stopCh)
}

func offlineBackgroundLoop(expiresAt time.Time, stop <-chan struct{}) {
	var lastWarning time.Time
	for {
		now := time.Now()
		if !now.Before(expiresAt) {
			degradeToFree("Offline license expired on " + expiresAt.Format(time.RFC3339) + ".")
			return
		}
		if expiresAt.Sub(now) <= offlineWarnBefore && now.Sub(lastWarning) >= offlineWarnInterval {
			log.Warningf("Offline license expires on %s. Generate a new token in Console.", expiresAt.Format(time.RFC3339))
			lastWarning = now
		}

		select {
		case <-stop:
			return
		case <-time.After(time.Hour):
		}
	}
}
