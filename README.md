**imgdiff** is an image comparison command line tool written in [Go](http://golang.org).
It can compare two images using a simple binary (bitmap) and perceptual algorithms.

```
$ imgdiff -h
Compare two images and optionally output resulting diff image.
Supported image formats: png, jpeg, gif, tiff, bmp and webp.

Exit code will be non-zero if the difference is above specified threshold.
Threshold value can also be a percentage, e.g. 0.5%.

Currently supported comparison algorithms are 'binary' and 'perceptual'.
Binary algorithm simply compares the two images' pixels as is.

image1 and image2 can be either local file paths or URLs.

Output is usually a file path. Specify '-' to write to stdout instead.
Resulting image format is inferred from the output file extension
or -of argument otherwise. It defaults to png.

Usage: imgdiff [options] image1 image2
  -a="perceptual": diff algorithm
  -o="": diff output
  -of="": output image format
  -t=0: threshold value
```


## Get the binary

Navigate to the [releases page](releases/latest) and download the executable
compatible with your OS and ARCH: darwin/linux/windows and 386/amd64 respectively.


## Compile

If you have Go installed, you can just do

```
go get github.com/crhym3/imgdiff/cmd/imgdiff
```


## License

(c) Google, 2015. Licensed under an [Apache-2](LICENSE) license.
