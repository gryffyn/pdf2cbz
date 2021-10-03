# pdf2cbz
Converts PDF files to CBZ archives.

### A Word of Warning
Don't use the PNG option as it is right now, it tends to increase the size of the CBZ relative to the PDF by about 10x. So a 58MB PDF outputs a CBZ around 530MB.

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
  -q, --quality=      JPEG quality (0-100) (default: 85)

Help Options:
  -h, --help          Show this help message
  
```
