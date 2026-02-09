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
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"
	"xorm.io/xorm/schemas"
)

type quoter interface {
	Quote(value string) string
}

func defaultWeights() []*UrgencyWeight {
	// Use sensible weights when none have been configured:
	return []*UrgencyWeight{
		{ProjectID: -1, Property: UrgencyDueDate.String(), Weight: 100},
		{ProjectID: -1, Property: UrgencyPercentDone.String(), Weight: 100},
		{ProjectID: -1, Property: UrgencyPriority.String(), Weight: 10},
	}
}

func normalizeMap(weights []*UrgencyWeight) []*UrgencyWeight {
	total := 0.0
	for _, weight := range weights {
		total += weight.Weight
	}
	var newWeights []*UrgencyWeight
	for _, weight := range weights {
		newWeight := *weight
		newWeight.Weight /= total
		newWeights = append(newWeights, &newWeight)
	}
	return newWeights
}

func urgencyScoreQuery(weights []*UrgencyWeight, quoter quoter, dbType schemas.DBType) (string, error) {
	if len(weights) == 0 {
		weights = defaultWeights()
	}
	weights = normalizeMap(weights)
	weightedScoreQuery, err := weightedScore(quoter, dbType, weights)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`(%s) as urgency`, weightedScoreQuery), nil
}

func weightedScore(quoter quoter, dbType schemas.DBType, weights []*UrgencyWeight) (string, error) {
	var weightedQueries []string
	for _, weight := range weights {
		var property UrgencyProperty
		if err := property.UnmarshalText([]byte(weight.Property)); err != nil {
			return "", err
		}
		query, err := property.normalizedPropertyScore(weight.Filter, quoter, dbType)
		if err != nil {
			return "", err
		}
		weightedQueries = append(weightedQueries, fmt.Sprintf("COALESCE(%s, 0) * %.3f", query, weight.Weight))
	}
	return fmt.Sprintf("\n\t%s\n", strings.Join(weightedQueries, " +\n\t")), nil
}

func divideColumn(quoter quoter, columnName, denominator string) string {
	return fmt.Sprintf("%s / %s", quoter.Quote(columnName), denominator)
}

// dueDateScoreQuery returns an urgency score for task's due_date.
//
// Must use a formula such that:
//
// - score is a function of x days overdue
// - score(4 days overdue)          = 100%
// - score(-8 days, in the future) = ~ 0%
// - Exponentially grows for a more "natural" sense of urgency. Human perception tends to be logarithmic.
// - Behaves identically in most common database engines.
//
// With this in mind, and selecting a gradual exponential growth factor of 2, our target formula is:
//
//	score(x) = 2^((x-4) / 2)
//
// Without SQLite's math library compiled in this is non-trivial, but not impossible.
// Using a few small transforms, we can represent this with positive (valid) bit shifts over our domain:
//
//	score = 2^((x+8-12)/2)
//	score = 2^((x+8)/2) / 2^(12/2)
//	score = 2^((x+8)/2) / 2^6
//
// And converting to SQL representation:
//
//	score = ( 1 << int((x+8)/2) ) / 64.0
func dueDateScoreQuery(quoter quoter, dbType schemas.DBType) (string, error) {
	const (
		upToDaysLate  = 2.5 // 100%
		showDaysEarly = -8  // 0%
		growthFactor  = 2

		domainWidth = upToDaysLate - showDaysEarly
	)
	positiveBitShiftCorrection := math.Pow(2, domainWidth/growthFactor)

	queryer, err := newUrgencyQueryerForType(quoter, dbType)
	if err != nil {
		return "", err
	}

	daysOverdue := fmt.Sprintf("- %s", daysUntil(queryer, "tasks.due_date"))
	exponent := queryer.ClampMin(0, fmt.Sprintf("(%s - %d) / %d", daysOverdue, showDaysEarly, growthFactor))
	return queryer.ClampMax(1, fmt.Sprintf(`(1 << (%s)) / %.2f`, exponent, positiveBitShiftCorrection)), nil
}

func daysUntil(queryer urgencyQueryer, timestampColumn string) string {
	const day = 24 * time.Hour
	return fmt.Sprintf(`cast(%s as %s)/%d`, queryer.SecondsUntil(timestampColumn), queryer.IntegerType(), int(day.Seconds()))
}

func newUrgencyQueryerForType(quoter quoter, dbType schemas.DBType) (urgencyQueryer, error) {
	switch dbType {
	case schemas.MYSQL:
		return mysqlUrgencyQueryer{
			standardUrgencyClampQueryer: standardUrgencyClampQueryer{quoter: quoter},
			quoter:                      quoter,
		}, nil
	case schemas.POSTGRES:
		return postgresUrgencyQueryer{
			standardUrgencyClampQueryer: standardUrgencyClampQueryer{quoter: quoter},
			quoter:                      quoter,
		}, nil
	case schemas.SQLITE:
		return sqliteUrgencyQueryer{quoter: quoter}, nil
	case
		schemas.DAMENG,
		schemas.GBASE8S,
		schemas.MSSQL,
		schemas.ORACLE:
	}
	return nil, errors.Errorf("unsupported database type: %s", dbType)
}

type urgencyQueryer interface {
	IntegerType() string
	ClampMin(minimum int, rawQuery string) string
	ClampMax(maximum int, rawQuery string) string
	SecondsUntil(timestampColumn string) string
}

type standardUrgencyClampQueryer struct {
	quoter quoter
}

func (s standardUrgencyClampQueryer) ClampMin(minimum int, rawQuery string) string {
	return fmt.Sprintf("greatest(%d, %s)", minimum, rawQuery)
}

func (s standardUrgencyClampQueryer) ClampMax(maximum int, rawQuery string) string {
	return fmt.Sprintf("least(%d, %s)", maximum, rawQuery)
}

type mysqlUrgencyQueryer struct {
	standardUrgencyClampQueryer
	quoter quoter
}

func (mysqlUrgencyQueryer) IntegerType() string { return "signed" }

func (m mysqlUrgencyQueryer) SecondsUntil(timestampColumn string) string {
	return fmt.Sprintf(`unix_timestamp(%s - now())`, m.quoter.Quote(timestampColumn))
}

type postgresUrgencyQueryer struct {
	standardUrgencyClampQueryer
	quoter quoter
}

func (postgresUrgencyQueryer) IntegerType() string { return "int" }

func (p postgresUrgencyQueryer) SecondsUntil(timestampColumn string) string {
	return fmt.Sprintf(`extract(epoch from %s - localtimestamp)`, p.quoter.Quote(timestampColumn))
}

type sqliteUrgencyQueryer struct {
	quoter quoter
}

func (sqliteUrgencyQueryer) IntegerType() string { return "int" }

func (s sqliteUrgencyQueryer) ClampMin(minimum int, rawQuery string) string {
	return fmt.Sprintf("max(%d, %s)", minimum, rawQuery)
}

func (s sqliteUrgencyQueryer) ClampMax(maximum int, rawQuery string) string {
	return fmt.Sprintf("min(%d, %s)", maximum, rawQuery)
}

func (s sqliteUrgencyQueryer) SecondsUntil(timestampColumn string) string {
	return fmt.Sprintf(`unixepoch(%s) - unixepoch('now')`, s.quoter.Quote(timestampColumn))
}
