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

// BenchmarkGetAvatar_Sequential tests memory usage for sequential avatar generation
func BenchmarkGetAvatar_Sequential(b *testing.B) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}
	testUser := &user.User{
		ID:       1,
		Name:     "Test User",
		Username: "testuser",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := provider.GetAvatar(testUser, 64)
		if err != nil {
			b.Fatalf("GetAvatar failed: %v", err)
		}
	}
}

// BenchmarkGetAvatar_DifferentSizes tests memory usage for different avatar sizes
func BenchmarkGetAvatar_DifferentSizes(b *testing.B) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}
	sizes := []int64{32, 64, 128, 256, 512}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			testUser := &user.User{
				ID:       1,
				Name:     "Test User",
				Username: "testuser",
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _, err := provider.GetAvatar(testUser, size)
				if err != nil {
					b.Fatalf("GetAvatar failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkGetAvatar_Concurrent tests memory usage under concurrent load
func BenchmarkGetAvatar_Concurrent(b *testing.B) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}
	concurrencyLevels := []int{10, 50, 100, 200}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrent_%d", concurrency), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				// Each goroutine gets a different user ID
				userID := int64(0)
				for pb.Next() {
					userID++
					testUser := &user.User{
						ID:       userID,
						Name:     fmt.Sprintf("User %d", userID),
						Username: fmt.Sprintf("user%d", userID),
					}

					_, _, err := provider.GetAvatar(testUser, 64)
					if err != nil {
						b.Fatalf("GetAvatar failed: %v", err)
					}
				}
			})
		})
	}
}

// TestMemoryUsage_Concurrent measures actual memory consumption during concurrent avatar generation
func TestMemoryUsage_Concurrent(t *testing.T) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}

	// Force garbage collection and get baseline memory
	runtime.GC()
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	concurrency := 10
	iterations := 10

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				userID := int64(workerID*iterations + j)
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
	t.Logf("  Iterations per worker: %d", iterations)
	t.Logf("  Total avatars generated: %d", concurrency*iterations)
	t.Logf("  Current allocated memory: %.2f MB", allocatedMB)
	t.Logf("  Total allocated memory: %.2f MB", totalAllocMB)
	t.Logf("  Heap allocated memory: %.2f MB", heapAllocMB)
	t.Logf("  Memory per avatar (avg): %.2f KB", totalAllocMB*1024/float64(concurrency*iterations))

	// Alert if memory usage is excessive (adjust threshold as needed)
	if heapAllocMB > 500 {
		t.Logf("WARNING: High heap memory usage detected (%.2f MB). This may lead to OOM in production.", heapAllocMB)
	}
}

// TestMemoryUsage_WithoutCache tests memory usage when cache is disabled
func TestMemoryUsage_WithoutCache(t *testing.T) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}

	// Force garbage collection and get baseline memory
	runtime.GC()
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Generate avatars for the same user repeatedly (simulating cache misses)
	iterations := 10
	testUser := &user.User{
		ID:       99999,
		Name:     "Cache Test User",
		Username: "cachetest",
	}

	for i := 0; i < iterations; i++ {
		// Clear cache before each generation to simulate worst case
		_ = provider.FlushCache(testUser)

		_, _, err := provider.GetAvatar(testUser, 64)
		if err != nil {
			t.Fatalf("GetAvatar failed: %v", err)
		}
	}

	// Force garbage collection and get final memory
	runtime.GC()
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Calculate memory usage
	totalAllocMB := float64(memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc) / 1024 / 1024

	t.Logf("Memory Usage Without Cache:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total allocated memory: %.2f MB", totalAllocMB)
	t.Logf("  Memory per uncached avatar: %.2f KB", totalAllocMB*1024/float64(iterations))
}

// TestMemoryLeak_RepeatedGeneration checks for memory leaks with repeated generation
func TestMemoryLeak_RepeatedGeneration(t *testing.T) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}

	measurements := 5
	iterationsPerMeasurement := 50

	var memoryReadings []uint64

	for m := 0; m < measurements; m++ {
		runtime.GC()
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		memoryReadings = append(memoryReadings, memStats.HeapAlloc)

		// Generate many avatars
		for i := 0; i < iterationsPerMeasurement; i++ {
			testUser := &user.User{
				ID:       int64(m*iterationsPerMeasurement + i),
				Name:     fmt.Sprintf("User %d", m*iterationsPerMeasurement+i),
				Username: fmt.Sprintf("user%d", m*iterationsPerMeasurement+i),
			}

			_, _, err := provider.GetAvatar(testUser, 64)
			if err != nil {
				t.Fatalf("GetAvatar failed: %v", err)
			}
		}
	}

	// Analyze memory trend
	t.Logf("Memory Leak Analysis:")
	for i, mem := range memoryReadings {
		t.Logf("  Measurement %d: %.2f MB", i+1, float64(mem)/1024/1024)
	}

	// Check if memory is growing linearly (potential leak indicator)
	if len(memoryReadings) >= 3 {
		firstThird := memoryReadings[0]
		lastThird := memoryReadings[len(memoryReadings)-1]
		growth := float64(lastThird-firstThird) / float64(firstThird) * 100

		t.Logf("  Memory growth: %.2f%%", growth)

		if growth > 100 {
			t.Logf("WARNING: Significant memory growth detected (%.2f%%). Possible memory leak.", growth)
		}
	}
}

// BenchmarkDrawImage tests the core image drawing performance and memory
func BenchmarkDrawImage(b *testing.B) {
	bg := avatarBgColors[0]
	testRune := 'A'

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := drawImage(testRune, bg)
		if err != nil {
			b.Fatalf("drawImage failed: %v", err)
		}
	}
}
