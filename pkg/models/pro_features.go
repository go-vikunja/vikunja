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

package models

import (
	"time"

	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// ProFeatureInstanceDefault stores the admin-set instance-wide default for a
// per-user toggleable pro feature. Without a row, the code default applies.
type ProFeatureInstanceDefault struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"-"`
	Feature string    `xorm:"varchar(50) not null unique" json:"feature"`
	Enabled bool      `xorm:"not null" json:"enabled"`
	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`
}

func (ProFeatureInstanceDefault) TableName() string {
	return "pro_feature_instance_defaults"
}

// GetProFeatureInstanceDefaults returns all admin-set instance defaults keyed
// by feature string.
func GetProFeatureInstanceDefaults(s *xorm.Session) (map[string]bool, error) {
	defaults := []*ProFeatureInstanceDefault{}
	if err := s.Find(&defaults); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(defaults))
	for _, d := range defaults {
		out[d.Feature] = d.Enabled
	}
	return out, nil
}

// SetProFeatureInstanceDefault upserts the instance-wide default for a feature.
func SetProFeatureInstanceDefault(s *xorm.Session, feature license.Feature, enabled bool) error {
	existing := &ProFeatureInstanceDefault{}
	has, err := s.Where("feature = ?", feature.String()).Get(existing)
	if err != nil {
		return err
	}
	if has {
		existing.Enabled = enabled
		_, err = s.ID(existing.ID).Cols("enabled").Update(existing)
		return err
	}
	_, err = s.Insert(&ProFeatureInstanceDefault{Feature: feature.String(), Enabled: enabled})
	return err
}

// ResetProFeatureInstanceDefault removes the instance-wide default for a
// feature so the code default applies again.
func ResetProFeatureInstanceDefault(s *xorm.Session, feature license.Feature) error {
	_, err := s.Where("feature = ?", feature.String()).Delete(&ProFeatureInstanceDefault{})
	return err
}

// resolvePerUserProFeature resolves the per-user layer only (override →
// instance default → code default). The caller must already have checked the
// license layer.
func resolvePerUserProFeature(s *xorm.Session, u *user.User, feature license.Feature) (bool, error) {
	if u != nil {
		if v, ok := u.ProFeatureOverrides[feature.String()]; ok {
			return v, nil
		}
	}
	defaults, err := GetProFeatureInstanceDefaults(s)
	if err != nil {
		return false, err
	}
	if v, ok := defaults[feature.String()]; ok {
		return v, nil
	}
	return license.PerUserDefault(feature), nil
}

// IsProFeatureEnabledForUser returns whether a feature is effectively enabled
// for the given user: the instance license must include it, and for per-user
// toggleable features the user override / instance default / code default
// chain must resolve to enabled. The user must carry its DB state — claim-
// derived users miss ProFeatureOverrides.
func IsProFeatureEnabledForUser(s *xorm.Session, u *user.User, feature license.Feature) (bool, error) {
	if !license.IsFeatureEnabled(feature) {
		return false, nil
	}
	if !license.IsPerUserToggleable(feature) {
		return true, nil
	}
	return resolvePerUserProFeature(s, u, feature)
}

// IsProFeatureEnabledForAuth resolves the authenticated principal and checks
// the feature for it. Link shares carry no per-user override, so only the
// instance default / code default chain applies to them.
func IsProFeatureEnabledForAuth(s *xorm.Session, a web.Auth, feature license.Feature) (bool, error) {
	if !license.IsFeatureEnabled(feature) {
		return false, nil
	}
	if !license.IsPerUserToggleable(feature) {
		return true, nil
	}

	var u *user.User
	if _, isUser := a.(*user.User); isUser {
		// Re-read from the DB: the auth user is claim-derived and does not
		// include ProFeatureOverrides.
		fresh, err := user.GetUserByID(s, a.GetID())
		if err != nil {
			return false, err
		}
		u = fresh
	}
	return resolvePerUserProFeature(s, u, feature)
}

// EffectiveProFeaturesForUser returns the pro features effectively enabled for
// the given user, for exposure to clients.
func EffectiveProFeaturesForUser(s *xorm.Session, u *user.User) ([]license.Feature, error) {
	enabled := license.EnabledProFeatures()
	out := make([]license.Feature, 0, len(enabled))
	for _, f := range enabled {
		if license.IsPerUserToggleable(f) {
			on, err := resolvePerUserProFeature(s, u, f)
			if err != nil {
				return nil, err
			}
			if !on {
				continue
			}
		}
		out = append(out, f)
	}
	return out, nil
}

// ProFeatureState describes one pro feature for the admin panel: its license
// state and, for per-user toggleable features, the effective instance default.
type ProFeatureState struct {
	Feature           string `json:"feature" doc:"The feature key, e.g. time_tracking."`
	Licensed          bool   `json:"licensed" doc:"Whether the instance license includes this feature."`
	PerUserToggleable bool   `json:"per_user_toggleable" doc:"Whether admins can grant or revoke this feature per user. Instance-wide features are always on for everyone when licensed."`
	DefaultEnabled    bool   `json:"default_enabled" doc:"The default for users without an override. Only meaningful for per-user toggleable features."`
	DefaultSource     string `json:"default_source" enum:"code,instance" doc:"Where the default comes from: the built-in code default or an admin-set instance default."`
}

// GetProFeatureStates returns the admin view of every known pro feature.
func GetProFeatureStates(s *xorm.Session) ([]*ProFeatureState, error) {
	instanceDefaults, err := GetProFeatureInstanceDefaults(s)
	if err != nil {
		return nil, err
	}

	features := license.AllFeatures()
	out := make([]*ProFeatureState, 0, len(features))
	for _, f := range features {
		st := &ProFeatureState{
			Feature:           f.String(),
			Licensed:          license.IsFeatureEnabled(f),
			PerUserToggleable: license.IsPerUserToggleable(f),
		}
		if st.PerUserToggleable {
			st.DefaultEnabled = license.PerUserDefault(f)
			st.DefaultSource = "code"
			if v, ok := instanceDefaults[f.String()]; ok {
				st.DefaultEnabled = v
				st.DefaultSource = "instance"
			}
		}
		out = append(out, st)
	}
	return out, nil
}

// UserProFeatureState describes one per-user toggleable feature for a single user.
type UserProFeatureState struct {
	Feature   string `json:"feature" doc:"The feature key, e.g. time_tracking."`
	Override  *bool  `json:"override" doc:"The admin-set override for this user, null when the instance default applies."`
	Effective bool   `json:"effective" doc:"Whether the feature is effectively enabled for this user, license included."`
}

// GetUserProFeatureStates returns the per-user toggleable features with the
// user's override and effective state.
func GetUserProFeatureStates(s *xorm.Session, u *user.User) ([]*UserProFeatureState, error) {
	out := []*UserProFeatureState{}
	for _, f := range license.AllFeatures() {
		if !license.IsPerUserToggleable(f) {
			continue
		}
		st := &UserProFeatureState{Feature: f.String()}
		if v, ok := u.ProFeatureOverrides[f.String()]; ok {
			override := v
			st.Override = &override
		}
		effective, err := IsProFeatureEnabledForUser(s, u, f)
		if err != nil {
			return nil, err
		}
		st.Effective = effective
		out = append(out, st)
	}
	return out, nil
}

// SetUserProFeatureOverride sets or clears (enabled == nil) the per-user
// override for a feature.
func SetUserProFeatureOverride(s *xorm.Session, u *user.User, feature license.Feature, enabled *bool) error {
	if enabled == nil {
		delete(u.ProFeatureOverrides, feature.String())
	} else {
		if u.ProFeatureOverrides == nil {
			u.ProFeatureOverrides = map[string]bool{}
		}
		u.ProFeatureOverrides[feature.String()] = *enabled
	}
	_, err := s.Where("id = ?", u.ID).
		Cols("pro_feature_overrides").
		Update(u)
	return err
}
