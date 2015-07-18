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

// Package imgdiff provides image comparison using simple
// and perceptual diff algorithms.
package imgdiff

import (
	"errors"
	"image"
)

// ErrSize is used when the two images under comparison have different sizes.
var ErrSize = errors.New("images have different sizes")

// Differ is the image comparison interface.
// All supported algorithms implement it.
type Differ interface {
	// Compare compares images a and b, returning the resulting
	// difference image and the number of pixels that are different
	// according to an algorithm.
	//
	// It returns ErrSize if images have their width or height
	// do not match.
	Compare(a, b image.Image) (image.Image, int, error)
}
