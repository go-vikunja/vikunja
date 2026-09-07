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

// Package entitlement resolves per-user feature flags and limits. It sits
// above pkg/license: the license says what the instance can offer, a
// user_entitlements row says what one user gets. Gates call this package,
// never license directly.
package entitlement

import (
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

func init() {
	db.RegisterTables(GetTables())
}

func GetTables() []any {
	return []any{&UserEntitlement{}}
}

type Feature string

const (
	FeatureAdminPanel      Feature = "admin_panel"
	FeatureAuditLogs       Feature = "audit_logs"
	FeatureTimeTracking    Feature = "time_tracking"
	FeatureTeamCreation    Feature = "team_creation"
	FeatureMaxProjects     Feature = "max_projects"
	FeatureMaxStorageBytes Feature = "max_storage_bytes"
)

// instanceWide features never get user rows; a zero license means not license-gated.
type spec struct {
	limit        bool
	instanceWide bool
	license      license.Feature
}

var features = map[Feature]spec{
	FeatureAdminPanel:      {instanceWide: true, license: license.FeatureAdminPanel},
	FeatureAuditLogs:       {instanceWide: true, license: license.FeatureAuditLogs},
	FeatureTimeTracking:    {license: license.FeatureTimeTracking},
	FeatureTeamCreation:    {},
	FeatureMaxProjects:     {limit: true},
	FeatureMaxStorageBytes: {limit: true},
}

func (f Feature) isLimit() bool {
	return features[f].limit
}

func InstanceWide(f Feature) bool {
	return features[f].instanceWide
}

// UserEntitlement is one resolved entitlement row. Absent row = no restriction.
type UserEntitlement struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"-"`
	UserID  int64     `xorm:"bigint not null index unique(user_feature)" json:"-"`
	Feature Feature   `xorm:"varchar(50) not null unique(user_feature)" json:"feature"`
	Value   int64     `xorm:"bigint not null" json:"value"`
	Created time.Time `xorm:"created not null" json:"-"`
	Updated time.Time `xorm:"updated not null" json:"-"`
}

func (UserEntitlement) TableName() string {
	return "user_entitlements"
}

// LicenseAllows is the instance-wide half of Check. Features unknown to the
// license are always allowed.
func LicenseAllows(f Feature) bool {
	lf := features[f].license
	if lf == license.FeatureUnknown {
		return true
	}
	return license.IsFeatureEnabled(lf)
}

func getRow(s *xorm.Session, userID int64, f Feature) (*UserEntitlement, bool, error) {
	row := &UserEntitlement{}
	has, err := s.Where("user_id = ? AND feature = ?", userID, f).Get(row)
	if err != nil {
		return nil, false, err
	}
	return row, has, nil
}

// Check returns nil when f is available to a. Otherwise it returns
// ErrFeatureNotLicensed (404, as today) when the instance lacks the feature,
// or ErrFeatureDisabledForUser (403) when a user row turns it off.
func Check(s *xorm.Session, a web.Auth, f Feature) error {
	if !LicenseAllows(f) {
		return ErrFeatureNotLicensed{Feature: f}
	}
	if InstanceWide(f) {
		return nil
	}
	userID, ok := user.SubjectID(a)
	if !ok {
		return nil
	}
	row, found, err := getRow(s, userID, f)
	if err != nil {
		return err
	}
	if found && row.Value == 0 {
		return ErrFeatureDisabledForUser{Feature: f}
	}
	return nil
}

// Limit returns the maximum for a limit feature charged to userID. Absent row
// means unlimited (limited == false).
func Limit(s *xorm.Session, userID int64, f Feature) (limit int64, limited bool, err error) {
	row, found, err := getRow(s, userID, f)
	if err != nil {
		return 0, false, err
	}
	if !found {
		return 0, false, nil
	}
	return row.Value, true, nil
}

func LimitsIn(entitlements map[Feature]int64) []Feature {
	limits := make([]Feature, 0, len(entitlements))
	for f := range entitlements {
		if f.isLimit() {
			limits = append(limits, f)
		}
	}
	return limits
}

func Rows(s *xorm.Session, userID int64) (map[Feature]int64, error) {
	rows := []*UserEntitlement{}
	if err := s.Where("user_id = ?", userID).Find(&rows); err != nil {
		return nil, err
	}
	out := make(map[Feature]int64, len(rows))
	for _, r := range rows {
		out[r.Feature] = r.Value
	}
	return out, nil
}

// ForUser resolves everything for the /user response: every flag as 0/1
// (license ∩ row), limits only when a row exists. Gates never call this.
func ForUser(s *xorm.Session, userID int64) (map[Feature]int64, error) {
	rows, err := Rows(s, userID)
	if err != nil {
		return nil, err
	}
	out := make(map[Feature]int64, len(features))
	for f, sp := range features {
		if sp.limit {
			if v, found := rows[f]; found {
				out[f] = v
			}
			continue
		}
		on := LicenseAllows(f)
		if v, found := rows[f]; on && found && !sp.instanceWide {
			on = v != 0
		}
		if on {
			out[f] = 1
		} else {
			out[f] = 0
		}
	}
	return out, nil
}

// Replace makes rows the full set for userID: features missing from rows are
// deleted. Instance-wide and unknown features are rejected.
func Replace(s *xorm.Session, userID int64, rows map[Feature]int64) error {
	for f, v := range rows {
		if _, known := features[f]; !known || InstanceWide(f) {
			return ErrUnknownFeature{Feature: f}
		}
		if f.isLimit() {
			if v < 0 {
				return ErrInvalidValue{Feature: f, Value: v}
			}
			continue
		}
		if v != 0 && v != 1 {
			return ErrInvalidValue{Feature: f, Value: v}
		}
	}

	if _, err := s.Where("user_id = ?", userID).Delete(&UserEntitlement{}); err != nil {
		return err
	}

	if len(rows) == 0 {
		return nil
	}
	insert := make([]*UserEntitlement, 0, len(rows))
	for f, v := range rows {
		insert = append(insert, &UserEntitlement{UserID: userID, Feature: f, Value: v})
	}
	_, err := s.Insert(&insert)
	return err
}
