package services

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentPermissions(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	s := db.NewSession()
	defer s.Close()

	service := &CommentService{DB: db.GetEngine()}

	t.Run("read returns max permission for project owner", func(t *testing.T) {
		perms := service.Can(s, &models.TaskComment{TaskID: 1}, &user.User{ID: 1})
		canRead, maxPermission, err := perms.Read()
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(models.PermissionAdmin), maxPermission)
	})

	t.Run("read denies user without access", func(t *testing.T) {
		perms := service.Can(s, &models.TaskComment{TaskID: 1}, &user.User{ID: 999})
		canRead, maxPermission, err := perms.Read()
		require.NoError(t, err)
		assert.False(t, canRead)
		assert.Equal(t, 0, maxPermission)
	})

	t.Run("read returns link share permission", func(t *testing.T) {
		perms := service.Can(s, &models.TaskComment{TaskID: 37}, &user.User{ID: -2})
		canRead, maxPermission, err := perms.Read()
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(models.PermissionWrite), maxPermission)
	})

	t.Run("read returns read-only link share permission", func(t *testing.T) {
		perms := service.Can(s, &models.TaskComment{TaskID: 1}, &user.User{ID: -1})
		canRead, maxPermission, err := perms.Read()
		require.NoError(t, err)
		assert.True(t, canRead)
		assert.Equal(t, int(models.PermissionRead), maxPermission)
	})

	t.Run("create requires write permission", func(t *testing.T) {
		t.Run("allows project owner", func(t *testing.T) {
			canCreate, err := service.Can(s, &models.TaskComment{TaskID: 1}, &user.User{ID: 1}).Create()
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("denies user without access", func(t *testing.T) {
			canCreate, err := service.Can(s, &models.TaskComment{TaskID: 1}, &user.User{ID: 999}).Create()
			require.NoError(t, err)
			assert.False(t, canCreate)
		})

		t.Run("allows link share with write permission", func(t *testing.T) {
			canCreate, err := service.Can(s, &models.TaskComment{TaskID: 37}, &user.User{ID: -2}).Create()
			require.NoError(t, err)
			assert.True(t, canCreate)
		})

		t.Run("denies link share with read-only permission", func(t *testing.T) {
			canCreate, err := service.Can(s, &models.TaskComment{TaskID: 1}, &user.User{ID: -1}).Create()
			require.NoError(t, err)
			assert.False(t, canCreate)
		})
	})

	t.Run("update and delete require authorship", func(t *testing.T) {
		author := &user.User{ID: 1}
		otherUser := &user.User{ID: 3}

		t.Run("author can update", func(t *testing.T) {
			canUpdate, err := service.Can(s, &models.TaskComment{ID: 1, TaskID: 1}, author).Update()
			require.NoError(t, err)
			assert.True(t, canUpdate)
		})

		t.Run("non author cannot update", func(t *testing.T) {
			canUpdate, err := service.Can(s, &models.TaskComment{ID: 1, TaskID: 1}, otherUser).Update()
			require.NoError(t, err)
			assert.False(t, canUpdate)
		})

		t.Run("author can delete", func(t *testing.T) {
			canDelete, err := service.Can(s, &models.TaskComment{ID: 1, TaskID: 1}, author).Delete()
			require.NoError(t, err)
			assert.True(t, canDelete)
		})

		t.Run("non author cannot delete", func(t *testing.T) {
			canDelete, err := service.Can(s, &models.TaskComment{ID: 1, TaskID: 1}, otherUser).Delete()
			require.NoError(t, err)
			assert.False(t, canDelete)
		})
	})

	t.Run("read errors when task is missing", func(t *testing.T) {
		perms := service.Can(s, &models.TaskComment{TaskID: 99999}, &user.User{ID: 1})
		_, _, err := perms.Read()
		require.Error(t, err)
		assert.True(t, models.IsErrTaskDoesNotExist(err))
	})
}
