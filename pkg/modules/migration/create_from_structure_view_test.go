// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func TestInsertFromStructureRejectsCrossViewDefaultAndDoneBuckets(t *testing.T) {
	for _, field := range []string{"default", "done"} {
		t.Run(field, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			viewOne := &models.ProjectView{ID: 11, Title: "One", ViewKind: models.ProjectViewKindKanban, BucketConfigurationMode: models.BucketConfigurationModeManual}
			viewTwo := &models.ProjectView{ID: 22, Title: "Two", ViewKind: models.ProjectViewKindKanban, BucketConfigurationMode: models.BucketConfigurationModeManual}
			if field == "default" {
				viewOne.DefaultBucketID = 202
			} else {
				viewOne.DoneBucketID = 202
			}
			structure := []*models.ProjectWithTasksAndBuckets{{
				Project: models.Project{ID: 100, Title: "Cross-view invalid", Views: []*models.ProjectView{viewOne, viewTwo}},
				Buckets: []*models.Bucket{{ID: 101, Title: "One bucket", ProjectViewID: 11}, {ID: 202, Title: "Two bucket", ProjectViewID: 22}},
			}}
			err := InsertFromStructure(structure, &user.User{ID: 1})
			require.ErrorIs(t, err, errImportedViewBucketMismatch)
			db.AssertMissing(t, "projects", map[string]any{"title": "Cross-view invalid"})
		})
	}
}

func TestInsertFromStructureRejectsDuplicateLegacyViewIDs(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	structure := []*models.ProjectWithTasksAndBuckets{{
		Project: models.Project{ID: 100, Title: "Duplicate-view invalid", Views: []*models.ProjectView{
			{ID: 11, Title: "One", ViewKind: models.ProjectViewKindKanban, BucketConfigurationMode: models.BucketConfigurationModeManual, DefaultBucketID: 101},
			{ID: 11, Title: "Duplicate", ViewKind: models.ProjectViewKindKanban, BucketConfigurationMode: models.BucketConfigurationModeManual},
		}},
		Buckets: []*models.Bucket{{ID: 101, Title: "One bucket", ProjectViewID: 11}},
	}}
	err := InsertFromStructure(structure, &user.User{ID: 1})
	require.ErrorIs(t, err, errDuplicateImportedViewID)
	db.AssertMissing(t, "projects", map[string]any{"title": "Duplicate-view invalid"})
}
