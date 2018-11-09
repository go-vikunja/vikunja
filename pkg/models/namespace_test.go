package models

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNamespace_Create(t *testing.T) {
	// Create test database
	//assert.NoError(t, PrepareTestDatabase())

	// Dummy namespace
	dummynamespace := Namespace{
		Name:        "Test",
		Description: "Lorem Ipsum",
	}

	// Doer
	doer, err := GetUserByID(1)
	assert.NoError(t, err)

	// Try creating it
	assert.True(t, dummynamespace.CanCreate(&doer))
	err = dummynamespace.Create(&doer)
	assert.NoError(t, err)

	// check if it really exists
	assert.True(t, dummynamespace.CanRead(&doer))
	newOne := Namespace{ID: dummynamespace.ID}
	err = newOne.ReadOne()
	assert.NoError(t, err)
	assert.Equal(t, newOne.Name, "Test")

	// Try creating one without a name
	n2 := Namespace{}
	err = n2.Create(&doer)
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceNameCannotBeEmpty(err))

	// Try inserting one with a nonexistant user
	nUser := &User{ID: 9482385}
	dnsp2 := dummynamespace
	err = dnsp2.Create(nUser)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Update it
	assert.True(t, dummynamespace.CanUpdate(&doer))
	dummynamespace.Description = "Dolor sit amet."
	err = dummynamespace.Update()
	assert.NoError(t, err)

	// Try updating one with a nonexistant owner
	dummynamespace.Owner.ID = 94829838572
	err = dummynamespace.Update()
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Try updating without a name
	dummynamespace.Name = ""
	err = dummynamespace.Update()
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceNameCannotBeEmpty(err))

	// Try updating a nonexistant one
	n := Namespace{ID: 284729, Name: "Lorem"}
	err = n.Update()
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Delete it
	assert.True(t, dummynamespace.CanDelete(&doer))
	err = dummynamespace.Delete()
	assert.NoError(t, err)

	// Try deleting a nonexistant one
	err = n.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Check if it was successfully deleted
	err = dummynamespace.ReadOne()
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Get all namespaces of a user
	nsps, err := n.ReadAll(&doer, 1)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(nsps).Kind(), reflect.Slice)
	s := reflect.ValueOf(nsps)
	assert.Equal(t, 1, s.Len())
}
