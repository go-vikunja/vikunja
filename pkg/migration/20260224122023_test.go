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

package migration

import (
	"encoding/json"
	"testing"

	"xorm.io/xorm/schemas"
)

func TestConvertBucketConfigurations20260224122023_ConvertsStringAndPreservesObject(t *testing.T) {
	configs := []*bucketConfigFix20260224122023{
		{
			Title:  "Open",
			Filter: json.RawMessage(`"done = false"`),
		},
		{
			Title:  "Sorted",
			Filter: json.RawMessage(`{"filter":"priority > 2","sort_by":["position"],"order_by":["asc"]}`),
		},
		{
			Title:  "Empty",
			Filter: json.RawMessage(`null`),
		},
	}

	converted, changed, err := convertBucketConfigurations20260224122023(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed to be true when a string filter is present")
	}
	if len(converted) != 3 {
		t.Fatalf("expected 3 configs, got %d", len(converted))
	}

	if converted[0].Filter == nil || converted[0].Filter.Filter != "done = false" {
		t.Fatalf("expected first filter to be wrapped, got %#v", converted[0].Filter)
	}
	if converted[1].Filter == nil || converted[1].Filter.Filter != "priority > 2" {
		t.Fatalf("expected object filter preserved, got %#v", converted[1].Filter)
	}
	if len(converted[1].Filter.SortBy) != 1 || converted[1].Filter.SortBy[0] != "position" {
		t.Fatalf("expected object fields preserved, got %#v", converted[1].Filter)
	}
	if converted[2].Filter != nil {
		t.Fatalf("expected null filter to stay nil, got %#v", converted[2].Filter)
	}
}

func TestConvertBucketConfigurations20260224122023_NoChangesWhenAlreadyObjects(t *testing.T) {
	configs := []*bucketConfigFix20260224122023{
		{
			Title:  "Done",
			Filter: json.RawMessage(`{"filter":"done = true"}`),
		},
	}

	converted, changed, err := convertBucketConfigurations20260224122023(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Fatal("expected changed to be false when no string filters are present")
	}
	if len(converted) != 1 || converted[0].Filter == nil || converted[0].Filter.Filter != "done = true" {
		t.Fatalf("unexpected conversion result: %#v", converted)
	}
}

func TestBucketConfigurationWhereClause20260224122023_PostgresUsesTextCast(t *testing.T) {
	got := bucketConfigurationWhereClause20260224122023(schemas.POSTGRES)
	want := "bucket_configuration IS NOT NULL AND bucket_configuration::text != '' AND bucket_configuration::text != '[]' AND bucket_configuration::text != 'null'"

	if got != want {
		t.Fatalf("unexpected clause\nwant: %s\ngot:  %s", want, got)
	}
}

func TestBucketConfigurationWhereClause20260224122023_DefaultDoesNotCast(t *testing.T) {
	got := bucketConfigurationWhereClause20260224122023(schemas.SQLITE)
	want := "bucket_configuration IS NOT NULL AND bucket_configuration != '' AND bucket_configuration != '[]' AND bucket_configuration != 'null'"

	if got != want {
		t.Fatalf("unexpected clause\nwant: %s\ngot:  %s", want, got)
	}
}
