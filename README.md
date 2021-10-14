# pdf2cbz
[![CodeQL](https://github.com/gryffyn/pdf2cbz/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/gryffyn/pdf2cbz/actions/workflows/codeql-analysis.yml)

Converts PDF files to CBZ archives.

### A Word(s) of Warning
- Don't use the PNG option as it is right now, it tends to increase the size of the CBZ relative to the PDF by about 2-10x. So a 58MB PDF outputs a CBZ around 530MB.
- the crop feature is quite limited at the minute, it's like 1 am and my brain can't handle doing the math for percentage right now, so you'll just have to guess for the pixel dimensions to crop off. I'll fix this soon, moving to percentage off each side instead of px.

## Building
Requires CGO.

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
  pdf2cbz [OPTIONS] [<INPUT PDF>] [<OUTPUT CBZ>]

Application Options:
  -i, --images        Extracts images from PDF (default: extracts pages)
  -p, --png           Outputs pages as PNG (default: jpeg)
  -c, --crop=         Dimensions of the crop region (css shorthand, comma-delimited)
  -q, --quality=      JPEG quality (0-100) (default: 85)

Help Options:
  -h, --help          Show this help message
```
