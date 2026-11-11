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
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/jsonschema-go/jsonschema"
)

const webUIBasesParam = "web_ui_bases"

type webUIBase struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

func linkKindFor(operation string) string {
	switch operation {
	case "tasks-read", "tasks-create":
		return "authenticated_task"
	case "shares-read", "shares-create", "shares-list":
		return "project_share"
	default:
		return ""
	}
}

func webUIBasesSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "array",
		MaxItems:    ptrInt(8),
		Description: "User-supplied Web UI bases, including any path prefix. HTTP(S) only, without credentials, query or fragment. Labels distinguish public, VPN or other access variants. Never use the MCP endpoint as a guessed Web UI base. No network requests or persistent configuration changes are made.",
		Items: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"label": {Type: "string", MinLength: ptrInt(1), MaxLength: ptrInt(80)},
				"url":   {Type: "string", MinLength: ptrInt(1), MaxLength: ptrInt(2048)},
			},
			Required:             []string{"label", "url"},
			AdditionalProperties: falseSchema(),
		},
	}
}

func ptrInt(n int) *int {
	return &n
}

func takeWebUIBases(args map[string]json.RawMessage) ([]webUIBase, error) {
	raw, ok := args[webUIBasesParam]
	if !ok {
		return nil, nil
	}
	delete(args, webUIBasesParam)
	var bases []webUIBase
	if err := json.Unmarshal(raw, &bases); err != nil {
		return nil, errors.New("web_ui_bases must be an array of labeled Web UI URLs")
	}
	for i := range bases {
		normalized, err := normalizeWebUIBase(bases[i].URL)
		if err != nil {
			return nil, err
		}
		bases[i].URL = normalized
	}
	return bases, nil
}

func normalizeWebUIBase(raw string) (string, error) {
	invalid := errors.New("invalid web_ui_bases URL: supply an absolute HTTP(S) Web UI base without credentials, query, fragment, whitespace or dot path segments")
	u, err := url.Parse(raw)
	if err != nil {
		return "", invalid
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(raw, "#") || strings.HasSuffix(u.Host, ":") || strings.ContainsAny(raw, "\\") || strings.ContainsAny(u.Path, "\\") || strings.IndexFunc(raw, unicode.IsSpace) >= 0 {
		return "", invalid
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return "", invalid
		}
	}
	for _, part := range strings.Split(u.Path, "/") {
		if part == "." || part == ".." {
			return "", invalid
		}
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = strings.TrimRight(u.RawPath, "/")
	return strings.TrimRight(u.String(), "/"), nil
}

func addWebLinks(operation string, args map[string]json.RawMessage, bases []webUIBase, projected, original any) any {
	kind := linkKindFor(operation)
	if kind == "" {
		return projected
	}
	out, ok := projected.(map[string]any)
	if !ok {
		return projected
	}
	source, ok := original.(map[string]any)
	if !ok {
		return projected
	}
	if operation == "shares-list" {
		items, _ := source["items"].([]any)
		links := make([]any, 0, len(items))
		for _, item := range items {
			if share, ok := item.(map[string]any); ok {
				links = append(links, resourceWebLinks(kind, args, bases, share))
			}
		}
		out["web_links"] = links
		return out
	}
	out["web_links"] = resourceWebLinks(kind, args, bases, source)
	return out
}

func resourceWebLinks(kind string, args map[string]json.RawMessage, bases []webUIBase, item map[string]any) map[string]any {
	links := []map[string]string{}
	result := map[string]any{
		"kind":   kind,
		"id":     item["id"],
		"status": "missing_base_url",
		"urls":   links,
	}
	if kind == "project_share" {
		var project json.Number
		_ = json.Unmarshal(args["project"], &project)
		result["project_id"] = project
	}
	if len(bases) == 0 {
		return result
	}
	for _, base := range bases {
		u, _ := url.Parse(base.URL)
		if kind == "authenticated_task" {
			id, ok := item["id"].(json.Number)
			if !ok {
				return result
			}
			u.Path += "/tasks/" + id.String()
			if u.RawPath != "" {
				u.RawPath += "/tasks/" + id.String()
			}
		} else {
			hash, ok := item["hash"].(string)
			if !ok || hash == "" {
				return result
			}
			u.Path += "/"
			if u.RawPath != "" {
				u.RawPath += "/"
			}
			u.Fragment = "share-auth-token=" + hash
		}
		links = append(links, map[string]string{
			"label":    base.Label,
			"base_url": base.URL,
			"source":   "web_ui_bases argument",
			"url":      u.String(),
		})
	}
	result["status"] = "available"
	result["urls"] = links
	return result
}
