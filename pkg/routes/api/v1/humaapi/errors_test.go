package humaapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVikunjaErrorShape_BasicCodeMessage(t *testing.T) {
	err := NewVikunjaError(http.StatusForbidden, "Forbidden")
	b, marshalErr := json.Marshal(err)
	require.NoError(t, marshalErr)

	var got map[string]any
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, "Forbidden", got["message"])
	// must not include RFC 9457 fields
	_, hasType := got["type"]
	_, hasTitle := got["title"]
	assert.False(t, hasType, "unexpected RFC 9457 field 'type'")
	assert.False(t, hasTitle, "unexpected RFC 9457 field 'title'")
}

func TestVikunjaErrorShape_StatusCoderInterface(t *testing.T) {
	var e huma.StatusError = NewVikunjaError(http.StatusNotFound, "not found")
	assert.Equal(t, http.StatusNotFound, e.GetStatus())
}
