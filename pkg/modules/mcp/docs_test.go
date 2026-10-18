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
	"context"
	"slices"
	"testing"

	"code.vikunja.io/api/pkg/models"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func connect(t *testing.T, install func(srv *sdk.Server)) *sdk.ClientSession {
	t.Helper()
	ctx := context.Background()
	srv := sdk.NewServer(&sdk.Implementation{Name: "test", Version: "1"}, nil)
	install(srv)
	clientTransport, serverTransport := sdk.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func connectDocs(t *testing.T) *sdk.ClientSession {
	t.Helper()
	return connect(t, installDocs)
}

func TestFilterSyntaxOnlyWithFilterTools(t *testing.T) {
	for name, tc := range map[string]struct {
		path string
		want bool
	}{
		"filter tool authorized":    {"/api/v2/things", true},
		"no filter tool authorized": {"/api/v2/things/:id", false},
	} {
		t.Run(name, func(t *testing.T) {
			m := newTestModule(t)
			m.authorize = func(_ *models.APIToken, path, _ string) bool { return path == tc.path }
			cs := connect(t, func(srv *sdk.Server) { m.addToolsAuthorizedBy(srv, nil) })
			tools, err := cs.ListTools(context.Background(), nil)
			require.NoError(t, err)
			var names []string
			for _, tl := range tools.Tools {
				names = append(names, tl.Name)
			}
			assert.Equal(t, tc.want, slices.Contains(names, toolFilterSyntax), names)
		})
	}
}

func TestFilterSyntaxTool(t *testing.T) {
	cs := connectDocs(t)
	result, err := cs.CallTool(context.Background(), &sdk.CallToolParams{Name: toolFilterSyntax})
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.Equal(t, filterSyntax, result.Content[0].(*sdk.TextContent).Text)
}

func TestFilterSyntaxResource(t *testing.T) {
	cs := connectDocs(t)
	ctx := context.Background()
	list, err := cs.ListResources(ctx, nil)
	require.NoError(t, err)
	require.Len(t, list.Resources, 1)
	assert.Equal(t, filterSyntaxURI, list.Resources[0].URI)

	res, err := cs.ReadResource(ctx, &sdk.ReadResourceParams{URI: filterSyntaxURI})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	assert.Equal(t, filterSyntax, res.Contents[0].Text)
	assert.Equal(t, "text/markdown", res.Contents[0].MIMEType)
}
