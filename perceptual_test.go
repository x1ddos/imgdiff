// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package imgdiff

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"

	_ "golang.org/x/image/tiff"
)

func TestCompare(t *testing.T) {
	pdiff := NewPerceptual(2.2, 100.0, 45.0, 1.0, false)
	bdiff := NewBinary()
	tests := []struct {
		img1, img2 string
		d          Differ
		npix       int
	}{
		{"aqsis_vase_ref.png", "aqsis_vase.png", pdiff, 108},
		{"bug1102605_ref.tif", "bug1102605.tif", pdiff, 2198},
		{"bug1471457_ref.tif", "bug1471457.tif", pdiff, 7},
		{"cam_mb_ref.tif", "cam_mb.tif", pdiff, 0},
		{"fish1.png", "fish2.png", pdiff, 27324},
		{"aqsis_vase_ref.png", "aqsis_vase.png", bdiff, 14796},
		{"bug1102605_ref.tif", "bug1102605.tif", bdiff, 3207},
		{"bug1471457_ref.tif", "bug1471457.tif", bdiff, 35},
		{"cam_mb_ref.tif", "cam_mb.tif", bdiff, 7},
		{"fish1.png", "fish2.png", bdiff, 137671},
	}
	for i, test := range tests {
		a, err := readTestImage(test.img1)
		if err != nil {
			t.Errorf("(%d) %s: %v", i, test.img1, err)
			continue
		}
		b, err := readTestImage(test.img2)
		if err != nil {
			t.Errorf("(%d) %s: %v", i, test.img2, err)
			continue
		}
		_, n, err := test.d.Compare(a, b)
		if err != nil {
			t.Errorf("(%d) %s: %v", i, test.img1, err)
			continue
		}
		if n > test.npix {
			t.Errorf("(%d) %s: n=%d; want n <= %d", i, test.img1, n, test.npix)
		}
	}
}

func BenchmarkPCompare(b *testing.B) {
	m1 := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	m2 := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	d := NewDefaultPerceptual()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d.Compare(m1, m2)
		}
	})
}

func BenchmarkPyramid(b *testing.B) {
	m := make([][]float64, 100)
	for i := 0; i < len(m); i++ {
		m[i] = make([]float64, 100)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pyramid(m)
	}
}

func BenchmarkLAB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lab(1, 1, 1)
	}
}

func BenchmarkXYZ(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xyz(color.RGBA{10, 20, 30, 255}, 1)
	}
}

func BenchmarkCSF(b *testing.B) {
	for i := 0; i < b.N; i++ {
		csf(1.5, 100.0)
	}
}

func BenchmarkVmask(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vmask(2.5)
	}
}

func readTestImage(p string) (image.Image, error) {
	f, err := os.Open(filepath.Join("testdata", p))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}
