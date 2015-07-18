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
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func open(p string) io.ReadCloser {
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		res, err := http.Get(p)
		if err != nil {
			log.Fatal(err)
		}
		return res.Body
	}
	f, err := os.Open(p)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func readImage(p string) image.Image {
	r := open(p)
	defer r.Close()
	img, _, err := image.Decode(r)
	if err != nil {
		log.Fatalf("%s: %v", p, err)
	}
	return img
}

func writeImage(dst string, mf string, m image.Image) {
	var err error
	w := os.Stdout
	if dst != "-" {
		w, err = os.Create(dst)
		if err != nil {
			log.Fatal(err)
		}
	}
	if ext := filepath.Ext(dst); mf == "" && ext != "" {
		mf = ext[1:]
	}
	switch mf {
	default:
		err = png.Encode(w, m)
	case "jpg", "jpeg":
		err = jpeg.Encode(w, m, nil)
	case "gif":
		err = gif.Encode(w, m, nil)
	case "tif", "tiff":
		err = tiff.Encode(w, m, nil)
	case "bmp":
		err = bmp.Encode(w, m)
	}
	if err != nil {
		log.Fatal(err)
	}
}
