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

package initials

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
)

// TestMemoryUsage_Concurrent measures memory consumption during concurrent avatar generation
func TestMemoryUsage_Concurrent(t *testing.T) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}

	// Force garbage collection and get baseline memory
	runtime.GC()
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	concurrency := 10
	totalAvatars := 100
	iterationsPerWorker := totalAvatars / concurrency

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterationsPerWorker; j++ {
				userID := int64(workerID*iterationsPerWorker + j)
				testUser := &user.User{
					ID:       userID,
					Name:     fmt.Sprintf("User %d", userID),
					Username: fmt.Sprintf("user%d", userID),
				}

				_, _, err := provider.GetAvatar(testUser, 64)
				if err != nil {
					t.Errorf("GetAvatar failed for user %d: %v", userID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Force garbage collection and get final memory
	runtime.GC()
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory usage
	allocatedMB := float64(memStatsAfter.Alloc-memStatsBefore.Alloc) / 1024 / 1024
	totalAllocMB := float64(memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc) / 1024 / 1024
	heapAllocMB := float64(memStatsAfter.HeapAlloc-memStatsBefore.HeapAlloc) / 1024 / 1024

	t.Logf("Memory Usage Statistics:")
	t.Logf("  Concurrent workers: %d", concurrency)
	t.Logf("  Total avatars generated: %d", totalAvatars)
	t.Logf("  Current allocated memory: %.2f MB", allocatedMB)
	t.Logf("  Total allocated memory: %.2f MB", totalAllocMB)
	t.Logf("  Heap allocated memory: %.2f MB", heapAllocMB)
	t.Logf("  Memory per avatar (avg): %.2f KB", totalAllocMB*1024/float64(totalAvatars))

	// Alert if memory usage is excessive (adjust threshold as needed)
	if heapAllocMB > 500 {
		t.Logf("WARNING: High heap memory usage detected (%.2f MB). This may lead to OOM in production.", heapAllocMB)
	}
}
