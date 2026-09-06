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

	"xorm.io/builder"
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

// Flags resolve to 0/1, limits to a maximum. Order is the wire order of ForUser.
var (
	Flags  = []Feature{FeatureAdminPanel, FeatureAuditLogs, FeatureTimeTracking, FeatureTeamCreation}
	Limits = []Feature{FeatureMaxProjects, FeatureMaxStorageBytes}
)

// licenseFeatures maps the features the license server knows. Features absent
// here are not license-gated: only a user row can turn them off.
var licenseFeatures = map[Feature]license.Feature{
	FeatureAdminPanel:   license.FeatureAdminPanel,
	FeatureAuditLogs:    license.FeatureAuditLogs,
	FeatureTimeTracking: license.FeatureTimeTracking,
}

// instanceOnly features never get user rows; billing does not know them.
var instanceOnly = map[Feature]bool{
	FeatureAdminPanel: true,
	FeatureAuditLogs:  true,
}

func Parse(s string) (Feature, bool) {
	f := Feature(s)
	for _, known := range append(append([]Feature{}, Flags...), Limits...) {
		if known == f {
			return f, true
		}
	}
	return "", false
}

func (f Feature) IsLimit() bool {
	for _, l := range Limits {
		if l == f {
			return true
		}
	}
	return false
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

// LicenseAllows is the instance-wide half of Has. Features unknown to the
// license are always allowed.
func LicenseAllows(f Feature) bool {
	lf, gated := licenseFeatures[f]
	if !gated {
		return true
	}
	return license.IsFeatureEnabled(lf)
}

// SubjectID returns the user whose entitlements apply to a. Bots resolve to
// their owner so limits cannot be bypassed through bots. Link shares have no
// subject; the second return is false.
func SubjectID(a web.Auth) (int64, bool) {
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return 0, false
	}
	if u.IsBot() {
		return u.BotOwnerID, true
	}
	return u.ID, true
}

func getRow(s *xorm.Session, userID int64, f Feature) (*UserEntitlement, bool, error) {
	row := &UserEntitlement{}
	has, err := s.Where("user_id = ? AND feature = ?", userID, f).Get(row)
	if err != nil {
		return nil, false, err
	}
	return row, has, nil
}

// Has reports whether a flag feature is available to a: the license must allow
// it and, for users, no row may turn it off. Use Check to get the matching error.
func Has(s *xorm.Session, a web.Auth, f Feature) (bool, error) {
	return Check(s, a, f) == nil, nil
}

// Check returns nil when f is available to a. Otherwise it returns
// ErrFeatureNotLicensed (404, as today) when the instance lacks the feature,
// or ErrFeatureDisabledForUser (403) when a user row turns it off.
func Check(s *xorm.Session, a web.Auth, f Feature) error {
	if !LicenseAllows(f) {
		return ErrFeatureNotLicensed{Feature: f}
	}
	if instanceOnly[f] {
		return nil
	}
	userID, ok := SubjectID(a)
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

// Rows returns the raw entitlement rows for userID, keyed by feature.
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
	out := make(map[Feature]int64, len(Flags)+len(Limits))
	for _, f := range Flags {
		on := LicenseAllows(f)
		if v, found := rows[f]; on && found && !instanceOnly[f] {
			on = v != 0
		}
		if on {
			out[f] = 1
		} else {
			out[f] = 0
		}
	}
	for _, f := range Limits {
		if v, found := rows[f]; found {
			out[f] = v
		}
	}
	return out, nil
}

// Replace makes rows the full set for userID: features missing from rows are
// deleted. Instance-only and unknown features are rejected.
func Replace(s *xorm.Session, userID int64, rows map[Feature]int64) error {
	for f := range rows {
		if _, known := Parse(string(f)); !known || instanceOnly[f] {
			return ErrUnknownFeature{Feature: f}
		}
	}

	keep := make([]string, 0, len(rows))
	for f := range rows {
		keep = append(keep, string(f))
	}
	var cond builder.Cond = builder.Eq{"user_id": userID}
	if len(keep) > 0 {
		cond = builder.And(cond, builder.NotIn("feature", keep))
	}
	if _, err := s.Where(cond).Delete(&UserEntitlement{}); err != nil {
		return err
	}

	existing, err := Rows(s, userID)
	if err != nil {
		return err
	}
	for f, v := range rows {
		if _, found := existing[f]; found {
			_, err = s.Where("user_id = ? AND feature = ?", userID, f).
				Cols("value").
				Update(&UserEntitlement{Value: v})
		} else {
			_, err = s.Insert(&UserEntitlement{UserID: userID, Feature: f, Value: v})
		}
		if err != nil {
			return err
		}
	}
	return nil
}
