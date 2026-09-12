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

package mcp

import (
	"testing"

	"code.vikunja.io/api/pkg/models"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogActions(t *testing.T) {
	Init(newTestAPI(t), "")
	prev := routeAuthorizer
	routeAuthorizer = func(_ *models.APIToken, path, _ string) bool { return path == "/things/:id" }
	t.Cleanup(func() { routeAuthorizer = prev })
	all := catalogActions(nil, "", "")
	var names []string
	for _, a := range all {
		names = append(names, a.Name)
		assert.Nil(t, a.InputSchema)
	}
	assert.ElementsMatch(t, []string{
		"things_read",
		"things_update",
		"things_delete",
	}, names)
	one := catalogActions(nil, "things_read", "")
	require.Len(t, one, 1)
	assert.NotNil(t, one[0].InputSchema)
	assert.Len(t, catalogActions(nil, "", "things"), 3)
}

func TestCatalogTools_RejectsUnknownArguments(t *testing.T) {
	ctx := withTestCaller(t)
	srv := sdk.NewServer(&sdk.Implementation{
		Name:    "test",
		Version: "1",
	}, nil)
	installCatalogTools(srv, nil)
	clientTransport, serverTransport := sdk.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer ss.Close()
	client := sdk.NewClient(&sdk.Implementation{
		Name:    "test-client",
		Version: "1",
	}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer cs.Close()
	result, err := cs.CallTool(ctx, &sdk.CallToolParams{
		Name:      "find_action",
		Arguments: map[string]any{"bogus": true},
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*sdk.TextContent).Text, "invalid arguments for find_action")
}
