package migration

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestBackfillLegacyTaskCreators(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	engine := db.GetEngine()
	session := engine.NewSession()
	defer session.Close()

	legacyTasks := []*models.Task{
		{
			Title:       "legacy project 1",
			ProjectID:   1,
			CreatedByID: 0,
			Index:       995,
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		{
			Title:       "legacy project 2",
			ProjectID:   2,
			CreatedByID: 0,
			Index:       996,
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		{
			Title:       "legacy project missing owner",
			ProjectID:   999,
			CreatedByID: 0,
			Index:       997,
			Created:     time.Now(),
			Updated:     time.Now(),
		},
	}

	for _, task := range legacyTasks {
		_, err := session.Insert(task)
		require.NoError(t, err)
	}

	err := backfillLegacyTaskCreators(engine)
	require.NoError(t, err)

	type result struct {
		CreatedByID int64 `xorm:"created_by_id"`
	}

	for i, task := range legacyTasks {
		var row result
		has, err := engine.Table("tasks").Where("id = ?", task.ID).Get(&row)
		require.NoError(t, err)
		require.True(t, has)

		expected := int64(1)
		if i == 1 {
			expected = 3
		}
		require.Equal(t, expected, row.CreatedByID)
	}
}
