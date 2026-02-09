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
	"fmt"
	"iter"

	"code.vikunja.io/api/pkg/db"
	"github.com/pkg/errors"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// UrgencyProperty is a type of urgency measurement, like due date or percentage complete
//
//nolint:recvcheck // Can unmarshal from text, but is otherwise an immutable value and should use non-pointer receivers.
type UrgencyProperty int

const (
	UrgencyDueDate UrgencyProperty = iota + 1 // ensure zero-value is invalid to better detect bugs
	UrgencyMatchesFilter
	UrgencyPercentDone
	UrgencyPriority

	maxUrgencyProperty
)

func AllUrgencyProperties() iter.Seq[UrgencyProperty] {
	return func(yield func(UrgencyProperty) bool) {
		for property := UrgencyProperty(1); property < maxUrgencyProperty; property++ {
			if !yield(property) {
				return
			}
		}
	}
}

func (u UrgencyProperty) name() (string, error) {
	switch u {
	case UrgencyDueDate:
		return "due_date", nil
	case UrgencyMatchesFilter:
		return "matches_filter", nil
	case UrgencyPercentDone:
		return "percent_done", nil
	case UrgencyPriority:
		return "priority", nil
	case maxUrgencyProperty: // Also invalid
	}
	return "", fmt.Errorf("invalid urgency property enum value: %d", u)
}

func (u UrgencyProperty) String() string {
	name, err := u.name()
	if err != nil {
		return fmt.Sprintf("<err: %s>", err)
	}
	return name
}

func (u UrgencyProperty) MarshalText() ([]byte, error) {
	name, err := u.name()
	return []byte(name), err
}

func (u *UrgencyProperty) UnmarshalText(b []byte) error {
	name := string(b)
	for property := range AllUrgencyProperties() {
		if name == property.String() {
			*u = property
			return nil
		}
	}
	return fmt.Errorf("unknown urgency property: %q", string(b))
}

func (u UrgencyProperty) normalizedPropertyScore(filter *TaskCollection, quoter quoter, dbType schemas.DBType) (string, error) {
	switch u {
	case UrgencyDueDate:
		return dueDateScoreQuery(quoter, dbType)
	case UrgencyMatchesFilter:
		cond, err := filter.FilterCondition()
		if err != nil {
			return "", err
		}
		query := builder.NewWriter()
		if err := cond.WriteTo(query); err != nil {
			return "", errors.Wrap(err, "failed to render saved filter condition")
		}
		queryStr, err := builder.ConvertToBoundSQL(query.String(), query.Args())
		if err != nil {
			return "", errors.Wrap(err, "failed to bind filter args")
		}
		return fmt.Sprintf("CASE WHEN (%s) THEN 1 ELSE 0 END", queryStr), nil
	case UrgencyPercentDone:
		return quoter.Quote("tasks.percent_done"), nil
	case UrgencyPriority:
		return divideColumn(quoter, "tasks.priority", "5.0"), nil
	case maxUrgencyProperty: // Also invalid
	}
	return "", errors.Errorf("unrecognized urgency score property: %s", u)
}

type UrgencyWeight struct {
	ProjectID int64  `xorm:"not null"`
	Property  string `xorm:"varchar(50) not null"`
	// Filter is an optional reference to a filter. Property must be set to [UrgencyMatchesFilter].
	Filter *TaskCollection `xorm:"json null"`
	Weight float64         `xorm:"double not null"`
}

func (*UrgencyWeight) TableName() string {
	return "urgency_weights"
}

// GetUrgencyWeights returns this user's urgency weights.
func GetUrgencyWeights(s *xorm.Session, projectID int64) ([]*UrgencyWeight, error) {
	var urgencyWeights []*UrgencyWeight
	if err := s.Where(builder.Eq{"project_id": projectID}).Find(&urgencyWeights); err != nil {
		return nil, err
	}
	for _, weight := range urgencyWeights {
		var property UrgencyProperty
		if err := property.UnmarshalText([]byte(weight.Property)); err != nil {
			return nil, errors.Wrap(err, "found invalid property, which should only happen if the API was downgraded")
		}
	}
	return urgencyWeights, nil
}

type urgencyUniqueKey struct {
	Property string      `json:"property"`
	Filter   BasicFilter `json:"filter,omitempty"`
}

type BasicFilter struct {
	Query        string `json:"query"`
	IncludeNulls bool   `json:"include_nulls"`
}

// SetUrgencyWeights validates allWeights, then replaces this project's urgency weights with allWeights.
// allWeights should skip the ProjectID field, as those are overridden.
func SetUrgencyWeights(s *xorm.Session, projectID int64, allWeights []UrgencyWeight) (returnedErr error) {
	properties := make(map[urgencyUniqueKey]struct{})
	var newWeights []*UrgencyWeight
	for _, weight := range allWeights {
		weight.ProjectID = projectID
		uniqueKey := urgencyUniqueKey{
			Property: weight.Property,
		}
		var property UrgencyProperty
		if err := property.UnmarshalText([]byte(weight.Property)); err != nil {
			return err
		}
		if weight.Filter != nil && property != UrgencyMatchesFilter {
			return errors.Errorf("%s does not support 'filter'", weight.Property)
		}
		if property == UrgencyMatchesFilter {
			if weight.Filter == nil {
				return errors.New("filter must be set for matches_filter weight")
			}
			uniqueKey.Filter = BasicFilter{
				Query:        weight.Filter.Filter,
				IncludeNulls: weight.Filter.FilterIncludeNulls,
			}
			if weight.Filter.Filter == "" {
				return errors.New("filter query must be set")
			}
		}

		if _, exists := properties[uniqueKey]; exists {
			// TODO return user formattable error
			return fmt.Errorf("duplicate weight: %q", weight.Property)
		}
		properties[uniqueKey] = struct{}{}
		newWeights = append(newWeights, &weight)
	}
	return db.DoTransaction(s, func() error {
		if _, err := s.Where(builder.Eq{"project_id": projectID}).Delete(&UrgencyWeight{}); err != nil {
			sql, _ := s.LastSQL()
			return errors.Wrapf(err, "failed to mark existing weights for replacement: %s", sql)
		}
		if len(newWeights) > 0 {
			if _, err := s.InsertMulti(newWeights); err != nil {
				return errors.Wrap(err, "failed to set weights")
			}
		}
		return nil
	})
}
