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

package deck

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/log"
)

func init() {
	// Initialize logger for tests
	log.InitLogger()
}

func TestConvertColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid 6-char hex",
			input:    "ff0000",
			expected: "#ff0000",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid length",
			input:    "abc",
			expected: "",
		},
		{
			name:     "valid lowercase",
			input:    "00ff00",
			expected: "#00ff00",
		},
		{
			name:     "valid uppercase",
			input:    "0000FF",
			expected: "#0000FF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertColor(tt.input)
			if result != tt.expected {
				t.Errorf("convertColor(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDeckDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		checkYear int
	}{
		{
			name:      "valid RFC3339 date",
			input:     "2025-12-31T23:59:59+00:00",
			expectErr: false,
			checkYear: 2025,
		},
		{
			name:      "empty string",
			input:     "",
			expectErr: false,
			checkYear: 0,
		},
		{
			name:      "null string",
			input:     "null",
			expectErr: false,
			checkYear: 0,
		},
		{
			name:      "invalid format",
			input:     "2025-12-31",
			expectErr: true,
			checkYear: 0,
		},
		{
			name:      "valid with timezone",
			input:     "2024-06-15T10:30:00+02:00",
			expectErr: false,
			checkYear: 2024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDeckDate(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("parseDeckDate(%q) expected error but got none", tt.input)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("parseDeckDate(%q) unexpected error: %v", tt.input, err)
			}
			if tt.checkYear > 0 && result.Year() != tt.checkYear {
				t.Errorf("parseDeckDate(%q) year = %d, want %d", tt.input, result.Year(), tt.checkYear)
			}
			if tt.checkYear == 0 && !result.IsZero() && tt.input != "" && tt.input != "null" {
				t.Errorf("parseDeckDate(%q) expected zero time but got %v", tt.input, result)
			}
		})
	}
}

func TestFormatAssignees(t *testing.T) {
	m := &Migration{}

	tests := []struct {
		name     string
		input    []deckAssignedUser
		expected string
	}{
		{
			name:     "empty slice",
			input:    []deckAssignedUser{},
			expected: "",
		},
		{
			name: "single assignee",
			input: []deckAssignedUser{
				{
					Participant: deckUser{
						UID:         "user1",
						DisplayName: "John Doe",
					},
				},
			},
			expected: "\n\n---\n**Originally assigned to:**\n- @John Doe\n",
		},
		{
			name: "multiple assignees",
			input: []deckAssignedUser{
				{
					Participant: deckUser{
						UID:         "user1",
						DisplayName: "John Doe",
					},
				},
				{
					Participant: deckUser{
						UID:         "user2",
						DisplayName: "Jane Smith",
					},
				},
			},
			expected: "\n\n---\n**Originally assigned to:**\n- @John Doe\n- @Jane Smith\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.formatAssignees(tt.input)
			if result != tt.expected {
				t.Errorf("formatAssignees() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNormalizeDescription(t *testing.T) {
	m := &Migration{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "plain text without breaks",
			input:    "This is plain text",
			expected: "<p>This is plain text</p>\n",
		},
		{
			name:     "text with HTML br tags",
			input:    "Line 1<br>Line 2<br>Line 3",
			expected: "<p>Line 1\nLine 2\nLine 3</p>\n",
		},
		{
			name:     "text with self-closing br tags",
			input:    "Item 1<br/>Item 2<br/>Item 3",
			expected: "<p>Item 1\nItem 2\nItem 3</p>\n",
		},
		{
			name:     "text with br tags with space",
			input:    "URL 1<br />URL 2<br />URL 3",
			expected: "<p>URL 1\nURL 2\nURL 3</p>\n",
		},
		{
			name:     "mixed br tag formats",
			input:    "Line 1<br>Line 2<br/>Line 3<br />Line 4",
			expected: "<p>Line 1\nLine 2\nLine 3\nLine 4</p>\n",
		},
		{
			name:     "actual newlines preserved",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "<p>Line 1\nLine 2\nLine 3</p>\n",
		},
		{
			name:     "combined newlines and br tags",
			input:    "Line 1\nLine 2<br>Line 3",
			expected: "<p>Line 1\nLine 2\nLine 3</p>\n",
		},
		{
			name:     "wrapped HTTPS URL",
			input:    "<https://example.com>",
			expected: "<p>https://example.com</p>\n",
		},
		{
			name:     "wrapped HTTP URL",
			input:    "<http://example.com>",
			expected: "<p>http://example.com</p>\n",
		},
		{
			name:     "wrapped FTP URL",
			input:    "<ftp://example.com>",
			expected: "<p>ftp://example.com</p>\n",
		},
		{
			name:     "multiple wrapped URLs",
			input:    "<https://example.com>\n<https://example.org>",
			expected: "<p>https://example.com\nhttps://example.org</p>\n",
		},
		{
			name:     "wrapped URL with query parameters",
			input:    "<https://example.com?param=value&other=test>",
			expected: "<p>https://example.com?param=value&amp;other=test</p>\n",
		},
		{
			name:     "real Nextcloud Deck case",
			input:    "<https://www.exertis-connect.fr/fr/catalog/category/destockingproductlist/?cat=3678&destocking=1&mode=grid>\n\n<https://www.exertis-connect.fr/fr/613524.html>",
			expected: "<p>https://www.exertis-connect.fr/fr/catalog/category/destockingproductlist/?cat=3678&amp;destocking=1&amp;mode=grid</p>\n<p>https://www.exertis-connect.fr/fr/613524.html</p>\n",
		},
		{
			name:     "mixed content with wrapped URLs and br tags",
			input:    "Check out:<br><https://example.com><br>And also:<br><https://other.com>",
			expected: "<p>Check out:\nhttps://example.com\nAnd also:\nhttps://other.com</p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.normalizeDescription(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeDescription() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildLabelMap(t *testing.T) {
	m := &Migration{}

	labels := []deckLabel{
		{
			ID:        1,
			Title:     "Bug",
			Color:     "ff0000",
			BoardID:   42,
			DeletedAt: 0,
		},
		{
			ID:        2,
			Title:     "Feature",
			Color:     "00ff00",
			BoardID:   42,
			DeletedAt: 0,
		},
		{
			ID:        3,
			Title:     "Deleted",
			Color:     "0000ff",
			BoardID:   42,
			DeletedAt: time.Now().Unix(),
		},
	}

	labelMap := m.buildLabelMap(labels)

	// Should have 2 labels (third is deleted)
	if len(labelMap) != 2 {
		t.Errorf("buildLabelMap() returned %d labels, want 2", len(labelMap))
	}

	// Check first label
	if label, exists := labelMap[1]; !exists {
		t.Error("buildLabelMap() missing label with ID 1")
	} else {
		if label.Title != "Bug" {
			t.Errorf("Label 1 title = %q, want %q", label.Title, "Bug")
		}
		if label.HexColor != "#ff0000" {
			t.Errorf("Label 1 color = %q, want %q", label.HexColor, "#ff0000")
		}
	}

	// Check that deleted label is not in map
	if _, exists := labelMap[3]; exists {
		t.Error("buildLabelMap() should not include deleted label with ID 3")
	}
}
