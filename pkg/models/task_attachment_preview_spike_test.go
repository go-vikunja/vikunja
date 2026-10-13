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

// Spike only: compares JPEG, PNG and AVIF as preview formats. Not meant to be merged.

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gen2brain/gav1d/avif"
	"github.com/stretchr/testify/require"
)

const (
	spikeEnv       = "VIKUNJA_PREVIEW_SPIKE"
	spikeImagesEnv = "VIKUNJA_PREVIEW_SPIKE_IMAGES"
	spikeOutEnv    = "VIKUNJA_PREVIEW_SPIKE_OUT"
)

type spikeImage struct {
	kind string
	name string
	img  image.Image
}

// Relative to pkg/models, where go test runs.
var spikeRepoImages = []struct{ kind, name, path string }{
	{"photo", "lion", "../../frontend/tests/fixtures/image.jpg"},
	{"photo", "vicunas", "../../frontend/src/assets/no-auth-image.jpg"},
	{"screenshot", "docs", "testdata/preview-spike/vikunja-docs.png"},
	{"alpha", "icon", "../../desktop/icon.png"},
	{"alpha", "wunderlist", "../../frontend/originalMedia/images/migration/wunderlist.png"},
}

func spikeDecode(tb testing.TB, path string) image.Image {
	tb.Helper()
	f, err := os.Open(path) // #nosec G703 -- path from the developer's environment
	require.NoError(tb, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(tb, err, path)
	return img
}

// Extra photos are every image in the directory spikeImagesEnv names; they are not in the repo.
func spikeCorpus(tb testing.TB) []spikeImage {
	tb.Helper()
	corpus := make([]spikeImage, 0, len(spikeRepoImages))
	for _, src := range spikeRepoImages {
		corpus = append(corpus, spikeImage{kind: src.kind, name: src.name, img: spikeDecode(tb, src.path)})
	}

	dir := os.Getenv(spikeImagesEnv)
	if dir == "" {
		return corpus
	}
	extra, err := filepath.Glob(filepath.Join(dir, "*"))
	require.NoError(tb, err)
	sort.Strings(extra)
	for _, path := range extra {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		corpus = append(corpus, spikeImage{kind: "photo", name: name, img: spikeDecode(tb, path)})
	}
	return corpus
}

type spikeEncoder struct {
	name   string
	encode func(w io.Writer, img *image.NRGBA) error
	decode func(r io.Reader) (image.Image, error)
}

func spikeJPEG(quality int) spikeEncoder {
	return spikeEncoder{
		name: fmt.Sprintf("jpeg-q%d", quality),
		encode: func(w io.Writer, img *image.NRGBA) error {
			return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
		},
		decode: jpeg.Decode,
	}
}

func spikePNG() spikeEncoder {
	return spikeEncoder{
		name:   "png",
		encode: func(w io.Writer, img *image.NRGBA) error { return png.Encode(w, img) },
		decode: png.Decode,
	}
}

func spikeAVIF(quality, speed int) spikeEncoder {
	return spikeEncoder{
		name: fmt.Sprintf("avif-q%d-s%d", quality, speed),
		encode: func(w io.Writer, img *image.NRGBA) error {
			return avif.Encode(w, img, avif.EncodeOptions{Quality: quality, Speed: speed})
		},
		decode: func(r io.Reader) (image.Image, error) { return avif.Decode(r) },
	}
}

func spikeBenchEncoders() []spikeEncoder {
	encoders := []spikeEncoder{spikeJPEG(85), spikePNG()}
	for _, quality := range []int{50, 60, 70, 80, 90} {
		encoders = append(encoders, spikeAVIF(quality, avif.DefaultSpeed))
	}
	for _, speed := range []int{2, 4, 8, 10} {
		encoders = append(encoders, spikeAVIF(70, speed))
	}
	return encoders
}

func spikeQualityEncoders() []spikeEncoder {
	encoders := []spikeEncoder{spikeJPEG(75), spikeJPEG(85), spikeJPEG(90), spikeJPEG(95), spikePNG()}
	for quality := 30; quality <= 95; quality += 5 {
		encoders = append(encoders, spikeAVIF(quality, avif.DefaultSpeed))
	}
	for _, speed := range []int{0, 2, 4, 8, 10} {
		encoders = append(encoders, spikeAVIF(70, speed))
	}
	return encoders
}

var spikeWidths = []PreviewSize{PreviewSmall, PreviewMedium, PreviewLarge, PreviewExtraLarge}

// BenchmarkPreviewEncodeSpike measures only the encode, as the decode and the resize are the same for every format.
func BenchmarkPreviewEncodeSpike(b *testing.B) {
	for _, src := range spikeCorpus(b) {
		for _, size := range spikeWidths {
			resized := resizeImage(src.img, size.GetSize())
			for _, enc := range spikeBenchEncoders() {
				b.Run(fmt.Sprintf("%s/%s/w%d/%s", src.kind, src.name, size.GetSize(), enc.name), func(b *testing.B) {
					buf := &bytes.Buffer{}
					b.ReportAllocs()
					for b.Loop() {
						buf.Reset()
						if err := enc.encode(buf, resized); err != nil {
							b.Fatal(err)
						}
					}
					b.ReportMetric(float64(buf.Len()), "bytes")
				})
			}
		}
	}
}

// BenchmarkPreviewDecodeResizeSpike measures what a preview costs before it is encoded, whatever the format.
func BenchmarkPreviewDecodeResizeSpike(b *testing.B) {
	for _, src := range spikeRepoImages {
		content, err := os.ReadFile(src.path)
		require.NoError(b, err)
		for _, size := range spikeWidths {
			b.Run(fmt.Sprintf("%s/%s/w%d", src.kind, src.name, size.GetSize()), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					img, _, err := image.Decode(bytes.NewReader(content))
					if err != nil {
						b.Fatal(err)
					}
					resizeImage(img, size.GetSize())
				}
			})
		}
	}
}

// spikeOnWhite flattens the alpha so the color error is measured on what is displayed.
func spikeOnWhite(img image.Image) *image.NRGBA {
	flat := image.NewNRGBA(img.Bounds())
	draw.Draw(flat, flat.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(flat, flat.Bounds(), img, img.Bounds().Min, draw.Over)
	return flat
}

func spikeLuma(img *image.NRGBA) []float64 {
	luma := make([]float64, 0, len(img.Pix)/4)
	for i := 0; i < len(img.Pix); i += 4 {
		luma = append(luma, 0.299*float64(img.Pix[i])+0.587*float64(img.Pix[i+1])+0.114*float64(img.Pix[i+2]))
	}
	return luma
}

func spikePSNR(sumSquaredError float64, samples int) float64 {
	if sumSquaredError == 0 {
		return math.Inf(1)
	}
	return 10 * math.Log10(255*255/(sumSquaredError/float64(samples)))
}

func spikePSNRLuma(a, b []float64) float64 {
	var sum float64
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return spikePSNR(sum, len(a))
}

func spikePSNRRGB(a, b *image.NRGBA) float64 {
	var sum float64
	for i := 0; i < len(a.Pix); i += 4 {
		for c := range 3 {
			d := float64(a.Pix[i+c]) - float64(b.Pix[i+c])
			sum += d * d
		}
	}
	return spikePSNR(sum, len(a.Pix)/4*3)
}

func spikePSNRAlpha(a, b image.Image) float64 {
	bounds := a.Bounds()
	var sum float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alphaA := a.At(x, y).RGBA()
			_, _, _, alphaB := b.At(x-bounds.Min.X+b.Bounds().Min.X, y-bounds.Min.Y+b.Bounds().Min.Y).RGBA()
			d := float64(alphaA>>8) - float64(alphaB>>8)
			sum += d * d
		}
	}
	return spikePSNR(sum, bounds.Dx()*bounds.Dy())
}

// spikeSSIM is the mean SSIM of the luma over uniform 8x8 windows every 4 pixels, not the Gaussian-weighted original.
func spikeSSIM(a, b []float64, width, height int) float64 {
	const (
		window = 8
		step   = 4
		c1     = (0.01 * 255) * (0.01 * 255)
		c2     = (0.03 * 255) * (0.03 * 255)
	)
	var total float64
	var windows int
	for y := 0; y+window <= height; y += step {
		for x := 0; x+window <= width; x += step {
			var sumA, sumB, sumAA, sumBB, sumAB float64
			for wy := range window {
				row := (y+wy)*width + x
				for wx := range window {
					va, vb := a[row+wx], b[row+wx]
					sumA += va
					sumB += vb
					sumAA += va * va
					sumBB += vb * vb
					sumAB += va * vb
				}
			}
			const n = window * window
			meanA, meanB := sumA/n, sumB/n
			varA := sumAA/n - meanA*meanA
			varB := sumBB/n - meanB*meanB
			cov := sumAB/n - meanA*meanB
			total += ((2*meanA*meanB + c1) * (2*cov + c2)) / ((meanA*meanA + meanB*meanB + c1) * (varA + varB + c2))
			windows++
		}
	}
	return total / float64(windows)
}

// spikeEncodeMemory returns the bytes an encode allocates in total, and the highest heap growth
// seen while it ran with the GC collecting almost continuously, which approximates what it holds at once.
func spikeEncodeMemory(encode func()) (peak, allocated uint64) {
	defer debug.SetGCPercent(debug.SetGCPercent(1))
	runtime.GC()
	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	var highest atomic.Uint64
	stop := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		var stats runtime.MemStats
		for {
			runtime.ReadMemStats(&stats)
			if stats.HeapAlloc > highest.Load() {
				highest.Store(stats.HeapAlloc)
			}
			select {
			case <-stop:
				return
			case <-time.After(50 * time.Microsecond):
			}
		}
	}()

	encode()
	close(stop)
	<-stopped

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	if highest.Load() > before.HeapAlloc {
		peak = highest.Load() - before.HeapAlloc
	}
	return peak, after.TotalAlloc - before.TotalAlloc
}

func spikeWrite(t *testing.T, dir, name string, content []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), content, 0o600)) // #nosec G703 -- path from the developer's environment
}

// TestPreviewFormatSpike prints one tab-separated row per image, width and encoder.
// With spikeOutEnv set it also writes every encoded preview into that directory.
func TestPreviewFormatSpike(t *testing.T) {
	if os.Getenv(spikeEnv) == "" {
		t.Skipf("set %s=1 to run the preview format comparison", spikeEnv)
	}
	out := os.Getenv(spikeOutEnv)
	if out != "" {
		require.NoError(t, os.MkdirAll(out, 0o750)) // #nosec G703 -- path from the developer's environment
	}

	rows := []string{"kind\timage\twidth\theight\tencoder\tmime\tbytes\tpsnr_y\tpsnr_rgb\tssim_y\tpsnr_alpha\tpeak_heap\talloc_bytes\tsha256"}
	for _, src := range spikeCorpus(t) {
		for _, size := range spikeWidths {
			resized := resizeImage(src.img, size.GetSize())
			width, height := resized.Bounds().Dx(), resized.Bounds().Dy()
			reference := spikeOnWhite(resized)
			referenceLuma := spikeLuma(reference)

			for _, enc := range spikeQualityEncoders() {
				buf := &bytes.Buffer{}
				// Warm up so one-time table setup is not counted as the encode's memory.
				require.NoError(t, enc.encode(io.Discard, resized))
				peak, allocated := spikeEncodeMemory(func() {
					require.NoError(t, enc.encode(buf, resized))
				})
				encoded := buf.Bytes()

				again := &bytes.Buffer{}
				require.NoError(t, enc.encode(again, resized))
				require.Equal(t, encoded, again.Bytes(), "%s is not deterministic", enc.name)

				decoded, err := enc.decode(bytes.NewReader(encoded))
				require.NoError(t, err, "%s %s %d", src.name, enc.name, width)
				require.Equal(t, resized.Bounds().Size(), decoded.Bounds().Size())
				flat := spikeOnWhite(decoded)
				flatLuma := spikeLuma(flat)

				alpha := "-"
				if !resized.Opaque() && !strings.HasPrefix(enc.name, "jpeg") {
					alpha = fmt.Sprintf("%.2f", spikePSNRAlpha(resized, decoded))
				}

				rows = append(rows, fmt.Sprintf("%s\t%s\t%d\t%d\t%s\t%s\t%d\t%.2f\t%.2f\t%.4f\t%s\t%d\t%d\t%x",
					src.kind, src.name, width, height, enc.name, mimetype.Detect(encoded).String(), len(encoded),
					spikePSNRLuma(referenceLuma, flatLuma), spikePSNRRGB(reference, flat),
					spikeSSIM(referenceLuma, flatLuma, width, height), alpha, peak, allocated,
					sha256.Sum256(encoded)))

				if out != "" {
					ext := strings.SplitN(enc.name, "-", 2)[0]
					name := fmt.Sprintf("%s-%s-w%d-%s.%s", src.kind, src.name, width, enc.name, ext)
					spikeWrite(t, out, name, encoded)
				}
			}

			if out != "" {
				buf := &bytes.Buffer{}
				require.NoError(t, png.Encode(buf, resized))
				name := fmt.Sprintf("%s-%s-w%d-reference.png", src.kind, src.name, width)
				spikeWrite(t, out, name, buf.Bytes())
			}
		}
	}

	table := strings.Join(rows, "\n") + "\n"
	if out != "" {
		spikeWrite(t, out, "results.tsv", []byte(table))
	}
	t.Log("\n" + table)
}
