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

package imageutils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
)

var ErrUndecodable = errors.New("image cannot be decoded")

// Opener is called twice, so callers queued for capacity hold no storage handle or S3 body while waiting.
type Opener func() (io.ReadCloser, error)

func BytesOpener(data []byte) Opener {
	return func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
}

const maxHeaderBytes = 12 << 20

var bounded atomic.Pointer[decoder]

func init() {
	// Below MaxPixels the largest allowed image could never acquire.
	bounded.Store(newDecoder(2*MaxPixels, 30*time.Second))
}

type decoder struct {
	budget       int64
	slots        *semaphore.Weighted
	queueTimeout time.Duration
}

func newDecoder(pixelBudget int64, queueTimeout time.Duration) *decoder {
	return &decoder{
		budget:       pixelBudget,
		slots:        semaphore.NewWeighted(pixelBudget),
		queueTimeout: queueTimeout,
	}
}

// Decode holds the image's pixels against a process-wide budget until use returns,
// because the decoded image stays in memory while use resizes or encodes it.
func Decode(ctx context.Context, open Opener, use func(img image.Image, format string) error) error {
	return bounded.Load().decode(ctx, open, use)
}

func ExhaustBudgetForTests(queueTimeout time.Duration) (release func()) {
	original := bounded.Load()
	exhausted := newDecoder(original.budget, queueTimeout)
	exhausted.slots.TryAcquire(exhausted.budget)
	bounded.Store(exhausted)
	var once sync.Once
	return func() {
		once.Do(func() {
			bounded.Store(original)
			exhausted.slots.Release(exhausted.budget)
		})
	}
}

func (d *decoder) decode(ctx context.Context, open Opener, use func(img image.Image, format string) error) error {
	cfg, err := readConfig(open)
	if err != nil {
		return err
	}

	pixels := int64(cfg.Width) * int64(cfg.Height)
	if pixels > d.budget {
		return &ErrImageTooLarge{Width: cfg.Width, Height: cfg.Height}
	}
	if err := d.acquire(ctx, pixels); err != nil {
		return err
	}
	defer d.slots.Release(pixels)

	img, format, err := decodeImage(open)
	if err != nil {
		return err
	}
	return use(img, format)
}

func readConfig(open Opener) (image.Config, error) {
	header, err := open()
	if err != nil {
		return image.Config{}, err
	}
	defer header.Close()

	// DecodeConfig buffers the whole file for formats whose metadata may sit at the end, like TIFF.
	cfg, _, err := image.DecodeConfig(io.LimitReader(header, maxHeaderBytes))
	if err != nil {
		return image.Config{}, fmt.Errorf("%w: %w", ErrUndecodable, err)
	}
	if err := ValidateConfig(cfg); err != nil {
		return image.Config{}, err
	}
	return cfg, nil
}

func (d *decoder) acquire(ctx context.Context, pixels int64) error {
	queueCtx, cancel := context.WithTimeout(ctx, d.queueTimeout)
	defer cancel()

	if err := d.slots.Acquire(queueCtx, pixels); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("waiting to decode an image: %w", context.Cause(ctx))
		}
		return ErrDecodeBusy{}
	}
	return nil
}

func decodeImage(open Opener) (image.Image, string, error) {
	src, err := open()
	if err != nil {
		return nil, "", err
	}
	defer src.Close()

	img, format, err := image.Decode(src)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrUndecodable, err)
	}
	return img, format, nil
}
