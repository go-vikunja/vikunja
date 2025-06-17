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

package models

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/jaswdr/faker/v2"
)

func initBenchmarkConfig() {
	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") == "1" {
		config.InitConfig()
	} else {
		config.InitDefaultConfig()
		config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))
	}
}

// createBenchmarkData creates projects and tasks used for search benchmarks.
func createBenchmarkData(b *testing.B, needle string) *user.User {

	numberOfProjects := 10
	numberOfTasks := 2500

	s := db.NewSession()
	defer s.Close()

	f := faker.New()

	u, err := user.GetUserByID(s, 1)
	if err != nil {
		b.Fatalf("get user: %v", err)
	}

	for i := range numberOfProjects {
		p := &Project{Title: fmt.Sprintf("Project %d", i), OwnerID: u.ID}
		if _, err := s.Insert(p); err != nil {
			b.Fatalf("insert project: %v", err)
		}

		for j := range numberOfTasks {
			title := f.Lorem().Sentence(6)
			if rand.Intn(100) == 0 { //nolint:gosec
				title += " " + needle
			}
			desc := ""
			if j%2 == 0 {
				desc = f.Lorem().Paragraph(1)
			}
			if j%100 == 0 {
				if desc == "" {
					desc = f.Lorem().Paragraph(1)
				}
				words := strings.Split(desc, " ")
				mid := len(words) / 2
				words = append(words[:mid], append([]string{needle}, words[mid:]...)...)
				desc = strings.Join(words, " ")
			}
			t := &Task{
				Title:       title,
				Description: desc,
				ProjectID:   p.ID,
				CreatedByID: u.ID,
				Index:       int64(j + 1),
			}
			if _, err := s.Insert(t); err != nil {
				b.Fatalf("insert task: %v", err)
			}
		}
	}

	return u
}

func BenchmarkTaskSearch(b *testing.B) {
	const needle = "llama"

	initBenchmarkConfig()
	SetupTests()
	err := db.LoadFixtures()
	if err != nil {
		b.Fatalf("load fixtures: %v", err)
	}

	// Log database configuration
	b.Logf("Database Type: %s", config.DatabaseType.GetString())
	if config.TypesenseEnabled.GetBool() {
		b.Log("Typesense is enabled")
	}

	auth := createBenchmarkData(b, needle)

	if config.TypesenseEnabled.GetBool() {
		InitTypesense()
		if err := CreateTypesenseCollections(); err != nil {
			b.Skipf("typesense server not available: %v", err)
		}
		if err := ReindexAllTasks(); err != nil {
			b.Skipf("typesense server not available: %v", err)
		}
	}

	// Get all projects for the user
	s := db.NewSession()
	projects, _, _, err := getRawProjectsForUser(
		s,
		&projectOptions{
			user: auth,
			page: -1,
		},
	)
	s.Close()
	if err != nil {
		b.Fatalf("get projects: %v", err)
	}

	// Create search options
	opts := &taskSearchOptions{
		search:             needle,
		page:               1,
		perPage:            50,
		filter:             "done = false",
		filterTimezone:     "UTC",
		filterIncludeNulls: false,
	}

	b.Log("Setup done, starting benchmark...")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := db.NewSession()
		resultSlice, _, _, err := getRawTasksForProjects(s, projects, auth, opts)
		if len(resultSlice) == 0 {
			b.Fatalf("no results found for needle %q", needle)
		}
		s.Close()
		if err != nil {
			b.Fatalf("search error: %v", err)
		}
	}
}
