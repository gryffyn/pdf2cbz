# pdf2cbz
[![CodeQL](https://github.com/gryffyn/pdf2cbz/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/gryffyn/pdf2cbz/actions/workflows/codeql-analysis.yml)
[![Build Status](https://ci.gryffyn.io/api/badges/gryffyn/pdf2cbz/status.svg?ref=refs/heads/main)](https://ci.gryffyn.io/gryffyn/pdf2cbz)

Converts PDF files to CBZ archives.

### A Word(s) of Warning
- Don't use the PNG option as it is right now, it tends to increase the size of the CBZ relative to the PDF by about 2-10x. So a 58MB PDF outputs a CBZ around 530MB.

## Building
Requires CGO, and uses the MuPDF Fitz library.

```
git clone https://github.com/gryffyn/pdf2cbz
cd pdf2cbz
go build
```
alternatively,

`go install github.com/gryffyn/pdf2cbz`

## Usage
```
Usage:
  pdf2cbz.linux.x64 [OPTIONS] [<INPUT PDF>] [<OUTPUT CBZ>]

Application Options:
  -i, --images        Extracts images from PDF (default: extracts pages)
  -p, --png           Outputs pages as PNG (default: jpeg)
  -c, --crop=         Dimensions of the crop region (css shorthand, comma-delimited) in percent. Ex. '10,15,20,25' means 10% off the top, 15% off the right, so on
  -q, --quality=      JPEG quality (0-100) (default: 85)
      --debug         enable debug printing

Help Options:
  -h, --help          Show this help message
```
