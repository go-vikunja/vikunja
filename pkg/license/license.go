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
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"

	"github.com/google/uuid"
)

// Feature name constants for licensed features.
const (
	FeatureCustomFields = "custom_fields"
	FeatureAuditLog     = "audit_log"
)

// Status represents the license_status table.
type Status struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	InstanceID  string    `xorm:"varchar(36) not null" json:"instance_id"`
	LicenseKey  string    `xorm:"text not null" json:"-"`
	Response    string    `xorm:"text not null" json:"response"`
	ValidatedAt time.Time `xorm:"datetime null" json:"validated_at"`
	Created     time.Time `xorm:"created not null" json:"created"`
	Updated     time.Time `xorm:"updated not null" json:"updated"`
}

func (Status) TableName() string {
	return "license_status"
}

// state holds the current in-memory license state.
type state struct {
	mu              sync.RWMutex
	licensed        bool
	features        map[string]bool
	maxUsers        int64
	expiresAt       time.Time
	lastCheckFailed bool
}

var (
	currentState = &state{
		features: make(map[string]bool),
	}
	stopCh     chan struct{}
	instanceID string
)

// Init initializes the license system. It must be called after the database
// is ready and before the web server starts.
func Init() {
	key := config.LicenseKey.GetString()

	// Load or generate instance ID
	var err error
	instanceID, err = loadOrCreateInstanceID()
	if err != nil {
		log.Fatalf("Could not initialize license system: %s", err)
	}

	// No license key configured — community mode
	if key == "" {
		log.Infof("No license key configured. Running in community mode.")
		return
	}

	// Check for cached validation
	cached, err := loadCachedStatus()
	if err != nil {
		log.Errorf("Error loading cached license status: %s", err)
	}

	// If cache exists but key changed, invalidate it
	if cached != nil && cached.LicenseKey != key {
		log.Infof("License key changed, invalidating cache.")
		cached = nil
	}

	// Perform initial license check
	resp, err := checkLicense(key)
	switch {
	case err != nil:
		// Servers unreachable — check cache
		if cached != nil && time.Since(cached.ValidatedAt) < 72*time.Hour {
			log.Warningf("License check failed, using cached validation from %s.", cached.ValidatedAt.Format(time.RFC3339))
			if err := applyFromCache(cached); err != nil {
				log.Fatalf("Could not apply cached license: %s", err)
			}
		} else {
			log.Fatalf("Could not reach any license server and no cached validation exists. Vikunja will not start. Please check your network connectivity.")
		}
	case !resp.Valid:
		log.Fatalf("License key is invalid: %s. Vikunja will not start. Please check your license key configuration.", resp.Message)
	default:
		applyResponse(resp)
		if err := cacheResponse(key, resp); err != nil {
			log.Errorf("Error caching license response: %s", err)
		}
		log.Infof("License valid. Licensed features enabled.")
	}

	// Start background goroutine
	stopCh = make(chan struct{})
	go backgroundLoop(key)
}

// IsFeatureEnabled returns whether a specific licensed feature is enabled.
func IsFeatureEnabled(feature string) bool {
	currentState.mu.RLock()
	defer currentState.mu.RUnlock()
	if !currentState.licensed {
		return false
	}
	return currentState.features[feature]
}

// MaxUsersReached returns whether the licensed user limit has been reached.
// Returns false in community mode (no limit).
func MaxUsersReached() bool {
	currentState.mu.RLock()
	defer currentState.mu.RUnlock()
	if !currentState.licensed || currentState.maxUsers <= 0 {
		return false
	}

	s := db.NewSession()
	defer s.Close()

	count, err := s.Table("users").Count()
	if err != nil {
		log.Errorf("Error counting users for license check: %s", err)
		return false
	}

	return count >= currentState.maxUsers
}

// Shutdown stops the background license check goroutine.
func Shutdown() {
	if stopCh != nil {
		close(stopCh)
	}
}

func loadOrCreateInstanceID() (string, error) {
	s := db.NewSession()
	defer s.Close()

	status := &Status{}
	has, err := s.Get(status)
	if err != nil {
		return "", err
	}

	if has && status.InstanceID != "" {
		return status.InstanceID, nil
	}

	id := uuid.New().String()
	_, err = s.Insert(&Status{
		InstanceID: id,
		LicenseKey: "",
		Response:   "{}",
	})
	if err != nil {
		return "", err
	}

	if err := s.Commit(); err != nil {
		return "", err
	}

	return id, nil
}

func loadCachedStatus() (*Status, error) {
	s := db.NewSession()
	defer s.Close()

	status := &Status{}
	has, err := s.Get(status)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return status, nil
}

func applyResponse(resp *Response) {
	currentState.mu.Lock()
	defer currentState.mu.Unlock()

	currentState.licensed = true
	currentState.features = make(map[string]bool)
	for _, f := range resp.Features {
		currentState.features[f] = true
	}
	currentState.maxUsers = resp.MaxUsers
	currentState.expiresAt = resp.ExpiresAt
	currentState.lastCheckFailed = false
}

func applyFromCache(cached *Status) error {
	resp, err := parseResponse(cached.Response)
	if err != nil {
		return err
	}
	applyResponse(resp)
	return nil
}

func degradeToCommunity(reason string) {
	currentState.mu.Lock()
	defer currentState.mu.Unlock()

	currentState.licensed = false
	currentState.features = make(map[string]bool)
	currentState.maxUsers = 0
	currentState.lastCheckFailed = true

	log.Warningf("%s Licensed features have been disabled.", reason)
}

func cacheResponse(key string, resp *Response) error {
	raw, err := serializeResponse(resp)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	// Update the existing row
	_, err = s.Where("1=1").Update(&Status{
		LicenseKey:  key,
		Response:    raw,
		ValidatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	return s.Commit()
}

func backgroundLoop(key string) {
	for {
		interval := 24 * time.Hour
		currentState.mu.RLock()
		if currentState.lastCheckFailed {
			interval = 1 * time.Hour
		}
		currentState.mu.RUnlock()

		select {
		case <-stopCh:
			return
		case <-time.After(interval):
		}

		resp, err := checkLicense(key)
		if err != nil {
			// Servers unreachable
			cached, cacheErr := loadCachedStatus()
			if cacheErr != nil || cached == nil || time.Since(cached.ValidatedAt) >= 72*time.Hour {
				degradeToCommunity("License cache expired and no license server is reachable.")
				log.Warningf("Next retry in 1 hour.")
			} else {
				currentState.mu.Lock()
				currentState.lastCheckFailed = true
				currentState.mu.Unlock()
				log.Warningf("License check failed, using cached validation from %s. Next retry in 1 hour.", cached.ValidatedAt.Format(time.RFC3339))
			}
			continue
		}

		if !resp.Valid {
			// Clear cache
			if err := clearCache(); err != nil {
				log.Errorf("Error clearing license cache: %s", err)
			}
			degradeToCommunity("License is no longer valid: " + resp.Message + ".")
			continue
		}

		// Success
		wasFailure := false
		currentState.mu.RLock()
		wasFailure = currentState.lastCheckFailed || !currentState.licensed
		currentState.mu.RUnlock()

		applyResponse(resp)
		if err := cacheResponse(key, resp); err != nil {
			log.Errorf("Error caching license response: %s", err)
		}

		if wasFailure {
			log.Infof("License check successful. Licensed features re-enabled.")
		}
	}
}

func clearCache() error {
	s := db.NewSession()
	defer s.Close()

	_, err := s.Where("1=1").Update(&Status{
		LicenseKey:  "",
		Response:    "{}",
		ValidatedAt: time.Time{},
	})
	if err != nil {
		return err
	}

	return s.Commit()
}
