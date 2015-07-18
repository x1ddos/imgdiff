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
	"math"
	"sync"
)

var (
	// white values of XYZ colorspace
	whiteX, whiteY, whiteZ float64
)

const (
	// LAB colorspace
	epsilon = 216.0 / 24389.0
	kappa   = 24389.0 / 27.0
)

func init() {
	whiteX, whiteY, whiteZ = xyz(color.RGBA{0xff, 0xff, 0xff, 0xff}, 1)
}

type perceptual struct {
	gamma float64
	// luminance
	lum float64
	// test for luminance only
	nocolor bool
	// field of view
	fov float64
	// color factor
	cf float64
	// num one degree pixels
	odp float64
	// adaptation level index, starting from 0
	ai int
}

// NewPerceptual creates a new Differ based on perceptual diff algorithm.
func NewPerceptual(gamma, luminance, fov, cf float64, nocolor bool) Differ {
	d := &perceptual{
		gamma:   gamma,
		lum:     luminance,
		fov:     fov,
		cf:      cf,
		nocolor: nocolor,
		odp:     2 * math.Tan(fov*0.5*math.Pi/180) * 180 / math.Pi,
	}
	for n := 1.0; !(n > d.odp); n *= 2 {
		d.ai++
		if d.ai == lapLevels-1 {
			break
		}
	}
	return d
}

// NewDefaultPerceptual returns the result of calling NewPerceptual with:
//   gamma = 2.2
//   luminance = 100.0
//   fov = 45.0
//   cf = 1.0
//   nocolor = false
func NewDefaultPerceptual() Differ {
	return NewPerceptual(2.2, 100.0, 45.0, 1.0, false)
}

// Compare compares a and b using pdiff algorithm.
func (d *perceptual) Compare(a, b image.Image) (image.Image, int, error) {
	ab, bb := a.Bounds(), b.Bounds()
	w, h := ab.Dx(), ab.Dy()
	if w != bb.Dx() || h != bb.Dy() {
		return nil, -1, ErrSize
	}

	diff := image.NewNRGBA(image.Rect(0, 0, w, h))

	var (
		wg         sync.WaitGroup
		aLAB, bLAB [][]*labColor
		aLap, bLap [][][]float64
	)

	wg.Add(2)
	go func() {
		aLAB, aLap = labLap(a, d.gamma, d.lum)
		wg.Done()
	}()
	go func() {
		bLAB, bLap = labLap(b, d.gamma, d.lum)
		wg.Done()
	}()

	cpd := make([]float64, lapLevels) // cycles per degree
	cpd[0] = 0.5 * float64(w) / d.odp // 0.5 * pixels per degree
	for i := 1; i < lapLevels; i++ {
		cpd[i] = 0.5 * cpd[i-1]
	}
	csfMax := csf(3.248, 100.0)
	freq := make([]float64, lapLevels-2)
	for i := 0; i < lapLevels-2; i++ {
		freq[i] = csfMax / csf(cpd[i], 100.0)
	}

	wg.Wait()

	var npix int // num of diff pixels
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			adapt := math.Max(0.5*(aLap[d.ai][y][x]+bLap[d.ai][y][x]), 1e-5)
			mask := make([]float64, lapLevels-2)
			contrast := make([]float64, lapLevels-2)
			var contrastSum float64
			for i := 0; i < lapLevels-2; i++ {
				n1 := math.Abs(aLap[i][y][x] - aLap[i+1][y][x])
				n2 := math.Abs(bLap[i][y][x] - bLap[i+1][y][x])
				d1 := math.Abs(aLap[i+2][y][x])
				d2 := math.Abs(bLap[i+2][y][x])
				d := math.Max(d1, d2)
				contrast[i] = math.Max(n1, n2) / math.Max(d, 1e-5)
				mask[i] = vmask(contrast[i] * csf(cpd[i], adapt))
				contrastSum += contrast[i]
			}
			if contrastSum < 1e-5 {
				contrastSum = 1e-5
			}

			var factor float64
			for i := 0; i < lapLevels-2; i++ {
				factor += contrast[i] * freq[i] * mask[i] / contrastSum
			}
			if factor < 1 {
				factor = 1
			} else if factor > 10 {
				factor = 10
			}

			delta := math.Abs(aLap[0][y][x] - bLap[0][y][x])
			pass := true
			// pure luminance test
			if delta > factor*tvi(adapt) {
				pass = false
			} else if !d.nocolor {
				// CIE delta E test with modifications
				cf := d.cf
				// ramp down the color test in scotopic regions
				if adapt < 10.0 {
					// don't do color test at all
					cf = 0.0
				}
				da := aLAB[y][x].a - bLAB[y][x].a
				db := aLAB[y][x].b - bLAB[y][x].b
				if (da*da+db*db)*cf > factor {
					pass = false
				}
			}

			c := color.NRGBA{0, 0, 0, 0xff}
			if !pass {
				npix++
				c.R = 0xff
				//ar, ag, ab, _ := a.At(x, y).RGBA()
				//br, bg, bb, _ := b.At(x, y).RGBA()
				//c.R = uint8((math.Abs(float64(ar)-float64(br)) / 0xffff) * 0xff)
				//c.G = uint8((math.Abs(float64(ag)-float64(bg)) / 0xffff) * 0xff)
				//c.B = uint8((math.Abs(float64(ab)-float64(bb)) / 0xffff) * 0xff)
			}
			diff.Set(x, y, c)
		}
	}

	return diff, npix, nil
}

type labColor struct {
	l, a, b float64
}

func lab(x, y, z float64) *labColor {
	r := [3]float64{x / whiteX, y / whiteY, z / whiteZ}
	var f [3]float64
	for i := 0; i < 3; i++ {
		if r[i] > epsilon {
			f[i] = math.Pow(r[i], 1.0/3.0)
			continue
		}
		f[i] = (kappa*r[i] + 16.0) / 116.0
	}
	return &labColor{
		l: 116.0*f[1] - 16.0,
		a: 500.0 * (f[0] - f[1]),
		b: 200.0 * (f[1] - f[2]),
	}
}

func xyz(c color.Color, gamma float64) (float64, float64, float64) {
	r, g, b, _ := c.RGBA()
	rg := math.Pow(float64(r)/0xffff, gamma)
	gg := math.Pow(float64(g)/0xffff, gamma)
	bg := math.Pow(float64(b)/0xffff, gamma)
	x := rg*0.576700 + gg*0.185556 + bg*0.188212
	y := rg*0.297361 + gg*0.627355 + bg*0.0752847
	z := rg*0.0270328 + gg*0.0706879 + bg*0.991248
	return x, y, z
}

func labLap(m image.Image, gamma, lum float64) ([][]*labColor, [][][]float64) {
	w, h := m.Bounds().Dx(), m.Bounds().Dy()
	aLum, aLAB := make([][]float64, h), make([][]*labColor, h)
	for y := 0; y < h; y++ {
		aLum[y], aLAB[y] = make([]float64, w), make([]*labColor, w)
		for x := 0; x < w; x++ {
			cx, cy, cz := xyz(m.At(x, y), gamma)
			aLAB[y][x] = lab(cx, cy, cz)
			aLum[y][x] = cy * lum
		}
	}
	return aLAB, pyramid(aLum)
}

var (
	// max levels
	lapLevels = 8
	// filter kernel
	lapKernel = [5]float64{0.05, 0.25, 0.4, 0.25, 0.05}
)

// pyramid creates a Laplacian Pyramid out of the image m.
// The result is [level][y][x] where level ranges from 0 to lapLevels.
func pyramid(m [][]float64) [][][]float64 {
	h, w := len(m), len(m[0])
	p := make([][][]float64, lapLevels)
	for l := 0; l < lapLevels; l++ {
		p[l] = make([][]float64, h)
		// first level is a copy
		if l == 0 {
			for y := 0; y < h; y++ {
				p[l][y] = make([]float64, w)
				copy(p[l][y], m[y])
			}
			continue
		}
		// next levels are convolution of the previous one
		for y := 0; y < h; y++ {
			p[l][y] = make([]float64, w)
			for x := 0; x < w; x++ {
				for i := -2; i <= 2; i++ {
					for j := -2; j <= 2; j++ {
						ny := y + j
						if ny < 0 {
							ny = -ny
						}
						if ny >= h {
							ny = 2*h - ny - 1
						}
						nx := x + i
						if nx < 0 {
							nx = -nx
						}
						if nx >= w {
							nx = 2*w - nx - 1
						}
						p[l][y][x] += lapKernel[i+2] * lapKernel[j+2] * p[l-1][ny][nx]
					}
				}
			}
		}
	}
	return p
}

// csf computes the contrast sensitivity function (Barten SPIE 1989)
// given the cycles per degree cpd and luminance lum.
func csf(cpd, lum float64) float64 {
	a := 440.0 * math.Pow((1.0+0.7/lum), -0.2)
	b := 0.3 * math.Pow((1.0+100.0/lum), 0.15)
	return a * cpd * math.Exp(-b*cpd) * math.Sqrt(1.0+0.06*math.Exp(b*cpd))
}

// vmask is Visual Masking from Daly 1993, computed from contrast c.
func vmask(c float64) float64 {
	a := math.Pow(392.498*c, 0.7)
	b := math.Pow(0.0153*a, 4.0)
	return math.Pow(1.0+b, 0.25)
}

// tvi, Threshold vs Intensity, computes the threshold of visibility
// given the adaptation luminance al in candelas per square meter.
// It is based on Ward Larson Siggraph 1997.
func tvi(al float64) float64 {
	var r float64
	al = math.Log10(al)
	switch {
	case al < -3.94:
		r = -2.86
	case al < -1.44:
		r = math.Pow(0.405*al+1.6, 2.18) - 2.86
	case al < -0.0184:
		r = al - 0.395
	case al < 1.9:
		r = math.Pow(0.249*al+0.65, 2.7) - 0.72
	default:
		r = al - 1.255
	}

	return math.Pow(10.0, r)
}
