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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/crhym3/imgdiff"
)

const usageText = `Compare two images and optionally output resulting diff image.
Supported image formats: png, jpeg, gif, tiff, bmp and webp.

Exit code will be non-zero if the difference is above specified threshold.
Threshold value can also be a percentage, e.g. 0.5%.

Currently supported comparison algorithms are 'binary' and 'perceptual'.
Binary algorithm simply compares the two images' pixels as is.
Default is perceptual. Change using -a option.

Images can either be local file paths or URLs.

Output is usually a file path. Specify '-' to write to stdout instead.
Resulting image format is inferred from the output file extension
or -of argument otherwise. It defaults to png.

Examples:
  # compare two local PNG images using perceptual algorithm
  # and store the result in pdiff.png
  imgdiff -o pdiff.png image1.png image2.png

  # compare remote images w/o storing the result
  imgdiff http://example.org/image1.jpg http://example.org/image2.jpg

  # use binary comparison algorithm
  imgdiff -a binary -o bdiff.png image1.gif image2.gif

  # use threshold of 0.1%
  imgdiff -t 0.1% image1.tiff image2.tiff
`

var (
	version string // set by linker -X

	// cmd line arguments
	threshold = thresholdVar{value: 100}
	algorithm = flag.String("a", "perceptual", "diff algorithm")
	output    = flag.String("o", "", "diff output")
	outputFmt = flag.String("of", "", "output image format when -o -")
	// perceptual args
	gamma   = flag.Float64("g", 2.2, "gamma adjustment; perceptual only")
	lum     = flag.Float64("lum", 100.0, "luminance factor; perceptual only")
	fov     = flag.Float64("fov", 45.0, "field of view; perceptual only")
	cf      = flag.Float64("cf", 1.0, "color factor; perceptual only")
	nocolor = flag.Bool("nocolor", false, "don't use color during comparison; perceptual only")
)

func init() {
	flag.Var(&threshold, "t", "threshold value")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(0)
	flag.Usage = usage
	run()
}

func run() {
	flag.Parse()
	if flag.NArg() == 1 && flag.Arg(0) == "version" {
		fmt.Println(version)
		return
	}
	if flag.NArg() != 2 {
		log.Fatal("invalid number of positional arguments")
	}

	img1 := readImage(flag.Arg(0))
	img2 := readImage(flag.Arg(1))
	res, n, err := newDiffer().Compare(img1, img2)
	if err != nil {
		log.Fatal(err)
	}
	np := float64(n) / float64(res.Bounds().Dx()*res.Bounds().Dy())
	if threshold.percent && !(np > threshold.value) || !(float64(n) > threshold.value) {
		return
	}
	fmt.Printf("difference: %d pixel(s), %f%%\n", n, np)
	defer os.Exit(1)
	if *output == "" {
		return
	}
	writeImage(*output, *outputFmt, res)
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s\nUsage: imgdiff [options] image1 image2\n", usageText)
	flag.PrintDefaults()
}

func newDiffer() imgdiff.Differ {
	switch *algorithm {
	case "binary":
		return imgdiff.NewBinary()
	case "perceptual":
		return imgdiff.NewPerceptual(*gamma, *lum, *fov, *cf, *nocolor)
	}
	log.Fatalf("unsupported diff algorithm: %s", *algorithm)
	return nil
}

type thresholdVar struct {
	value   float64
	percent bool
}

func (v *thresholdVar) String() string {
	unit := ""
	if v.percent {
		unit = "%"
	}
	return fmt.Sprintf("%g%s", v.value, unit)
}

func (v *thresholdVar) Set(t string) error {
	if len(t) == 0 {
		v.value = 0
		return nil
	}
	percent := false
	if t[len(t)-1] == '%' {
		percent = true
		t = t[:len(t)-1]
	}
	val, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return err
	}
	v.percent = percent
	v.value = val
	return nil
}
