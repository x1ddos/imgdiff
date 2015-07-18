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
	"image/color"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExitCode(t *testing.T) {
	if os.Getenv("RUNME") == "1" {
		run()
		return
	}

	m := image.NewRGBA(image.Rect(0, 0, 100, 100))
	img1, err := writeTempImage(m)
	if err != nil {
		t.Fatal(err)
	}
	m.Set(0, 0, color.RGBA{0xff, 0xff, 0xff, 0xff})
	img2, err := writeTempImage(m)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(img1)
		os.Remove(img2)
	}()

	tests := []struct {
		opts string
		exit int
	}{
		{"-t 0 -a perceptual", 1},
		{"-t 0 -a binary", 1},
		{"-t 1 -a perceptual", 0},
		{"-t 1 -a binary", 0},
	}
	for i, test := range tests {
		args := append([]string{"-test.run=TestExitCode"}, strings.Split(test.opts, " ")...)
		args = append(args, img1, img2)
		cmd := exec.Command(os.Args[0], args...)
		cmd.Env = append(os.Environ(), "RUNME=1")
		out, err := cmd.CombinedOutput()
		e, ok := err.(*exec.ExitError)
		if !ok && err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		if test.exit == 0 && e == nil || test.exit != 0 && e != nil && !e.Success() {
			continue
		}
		t.Errorf("%d: err: %v; want exit code %d", i, err, test.exit)
		t.Log(string(out))
	}
}

func TestOpenURL(t *testing.T) {
	if os.Getenv("RUNME") == "1" {
		run()
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	img.Set(0, 0, color.RGBA{0xff, 0xff, 0xff, 0xff})
	imgpath, err := writeTempImage(img)
	if err != nil {
		t.Fatal(err)
	}

	fetched := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetched = true
		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	}))
	defer ts.Close()

	args := []string{"-test.run=TestOpenURL", "-a", "binary", imgpath, ts.URL}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "RUNME=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Error(err)
	}
	if !fetched {
		t.Errorf("image was never fetched from %s", ts.URL)
	}
}

func writeTempImage(m image.Image) (string, error) {
	f, err := ioutil.TempFile("", "img")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := png.Encode(f, m); err != nil {
		return "", err
	}
	return f.Name(), nil
}
