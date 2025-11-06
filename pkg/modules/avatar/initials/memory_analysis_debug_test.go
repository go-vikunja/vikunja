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
	"testing"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
)

// TestSingleAvatarMemory measures memory for generating a single avatar
func TestSingleAvatarMemory(t *testing.T) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}

	testUser := &user.User{
		ID:       12345,
		Name:     "Test User",
		Username: "testuser",
	}

	// Clear any existing cache
	_ = provider.FlushCache(testUser)

	// Force GC and measure before
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Generate one avatar
	avatar, mimeType, err := provider.GetAvatar(testUser, 64)
	if err != nil {
		t.Fatalf("GetAvatar failed: %v", err)
	}

	// Force GC and measure after
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	t.Logf("Single Avatar Generation:")
	t.Logf("  Avatar size (returned bytes): %d bytes (%.2f KB)", len(avatar), float64(len(avatar))/1024)
	t.Logf("  Mime type: %s", mimeType)
	t.Logf("  Heap allocated delta: %.2f MB", float64(memAfter.HeapAlloc-memBefore.HeapAlloc)/1024/1024)
	t.Logf("  Total allocated delta: %.2f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)
	t.Logf("  Mallocs: %d", memAfter.Mallocs-memBefore.Mallocs)
	t.Logf("  Frees: %d", memAfter.Frees-memBefore.Frees)
	t.Logf("  Live objects: %d", (memAfter.Mallocs-memAfter.Frees)-(memBefore.Mallocs-memBefore.Frees))
}

// TestFullSizeImageMemory measures memory for the full-size image generation
func TestFullSizeImageMemory(t *testing.T) {
	log.InitLogger()
	keyvalue.InitStorage()

	testUser := &user.User{
		ID:       54321,
		Name:     "Full Size Test",
		Username: "fullsize",
	}

	// Clear cache
	cacheKey := getCacheKey("full", testUser.ID)
	_ = keyvalue.Del(cacheKey)

	// Force GC and measure before
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Generate full size avatar
	fullAvatar, err := getAvatarForUser(testUser)
	if err != nil {
		t.Fatalf("getAvatarForUser failed: %v", err)
	}

	// Force GC and measure after
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate theoretical image size
	bounds := fullAvatar.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	theoreticalSize := width * height * 8 // RGBA64 = 8 bytes per pixel

	t.Logf("Full Size Image Generation:")
	t.Logf("  Image dimensions: %dx%d", width, height)
	t.Logf("  Theoretical size (RGBA64): %.2f MB", float64(theoreticalSize)/1024/1024)
	t.Logf("  Heap allocated delta: %.2f MB", float64(memAfter.HeapAlloc-memBefore.HeapAlloc)/1024/1024)
	t.Logf("  Total allocated delta: %.2f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)
}

// TestDrawImageMemory measures the drawImage function specifically
func TestDrawImageMemory(t *testing.T) {
	// Force GC and measure before
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	bg := avatarBgColors[0]
	testRune := 'A'

	img, err := drawImage(testRune, bg)
	if err != nil {
		t.Fatalf("drawImage failed: %v", err)
	}

	// Force GC and measure after
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	theoreticalSize := width * height * 8

	t.Logf("drawImage Function:")
	t.Logf("  Image dimensions: %dx%d", width, height)
	t.Logf("  Theoretical size: %.2f MB", float64(theoreticalSize)/1024/1024)
	t.Logf("  Heap allocated delta: %.2f MB", float64(memAfter.HeapAlloc-memBefore.HeapAlloc)/1024/1024)
	t.Logf("  Total allocated delta: %.2f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)
}

// TestSequentialVsConcurrentMemory compares sequential vs concurrent memory usage
func TestSequentialVsConcurrentMemory(t *testing.T) {
	log.InitLogger()
	keyvalue.InitStorage()

	provider := &Provider{}
	count := 10

	// Test sequential
	t.Run("Sequential", func(t *testing.T) {
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		for i := 0; i < count; i++ {
			testUser := &user.User{
				ID:       int64(10000 + i),
				Name:     fmt.Sprintf("User %d", i),
				Username: fmt.Sprintf("user%d", i),
			}
			_, _, err := provider.GetAvatar(testUser, 64)
			if err != nil {
				t.Fatalf("GetAvatar failed: %v", err)
			}
		}

		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		t.Logf("Sequential generation of %d avatars:", count)
		t.Logf("  Heap delta: %.2f MB", float64(memAfter.HeapAlloc-memBefore.HeapAlloc)/1024/1024)
		t.Logf("  Total alloc: %.2f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)
		t.Logf("  Per avatar: %.2f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024/float64(count))
	})

	// Test concurrent (would show higher peak memory)
	t.Run("Concurrent", func(t *testing.T) {
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		done := make(chan bool, count)
		for i := 0; i < count; i++ {
			go func(id int) {
				testUser := &user.User{
					ID:       int64(20000 + id),
					Name:     fmt.Sprintf("User %d", id),
					Username: fmt.Sprintf("user%d", id),
				}
				_, _, err := provider.GetAvatar(testUser, 64)
				if err != nil {
					t.Errorf("GetAvatar failed: %v", err)
				}
				done <- true
			}(i)
		}

		for i := 0; i < count; i++ {
			<-done
		}

		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		t.Logf("Concurrent generation of %d avatars:", count)
		t.Logf("  Heap delta: %.2f MB", float64(memAfter.HeapAlloc-memBefore.HeapAlloc)/1024/1024)
		t.Logf("  Total alloc: %.2f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024)
		t.Logf("  Per avatar: %.2f MB", float64(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024/float64(count))
	})
}
