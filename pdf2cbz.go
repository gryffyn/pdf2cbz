package main

import (
	"archive/zip"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/gen2brain/go-fitz"
	"github.com/jessevdk/go-flags"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/schollz/progressbar"
)

func main() {
	type Opts struct {
		Images      bool `short:"i" long:"images" description:"Extracts images from PDF (default: extracts pages)"`
		PNG         bool `short:"p" long:"png" description:"Outputs pages as PNG (default: jpeg)"`
		JPEGQuality int  `short:"q" long:"quality" description:"JPEG quality (0-100)" default:"85"`
		Positional  struct {
			PDF string `positional-arg-name:"<INPUT PDF>"`
			CBZ string `positional-arg-name:"<OUTPUT CBZ>"`
		} `positional-args:"yes"`
	}

	var opts Opts

	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()

	if opts.Positional.PDF == "" || opts.Positional.CBZ == "" {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	if !flags.WroteHelp(err) {
		if opts.Images {
			err = extractImages(opts.Positional.PDF, opts.Positional.CBZ)
		} else {
			err = extractPages(opts.Positional.PDF, opts.Positional.CBZ, opts.PNG, opts.JPEGQuality)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func extractPages(pdf,cbz string, usePNG bool, jpegQuality int) error {
	doc, err := fitz.New(pdf)
	if err != nil { return err }

	defer doc.Close()

	cFile, err := os.Create(cbz)
	if err != nil { return err }

	zWriter := zip.NewWriter(cFile)

	bar := progressbar.New(doc.NumPage())

	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil { return err }

		filetype := "jpg"
		if usePNG {
			filetype = "png"
		}

		h := &zip.FileHeader{Name: fmt.Sprintf("page%04d.%s", n, filetype), Method: zip.Deflate}
		f, err := zWriter.CreateHeader(h)
		if err != nil { return err }

		if usePNG {
			err = png.Encode(f, img)
			if err != nil { return err }
		} else {
			err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpegQuality})
			if err != nil { return err }
		}

		bar.Add(1)
	}

	zWriter.Close()
	cFile.Close()

	return nil
}

func extractImages(pdf,cbz string) error {
	cFile, err := os.Create(cbz)
	if err != nil { return err }
	zWriter := zip.NewWriter(cFile)

	tmp, err := os.MkdirTemp("", "pdf2cbz")
	if err != nil { return err }

	err = api.ExtractImagesFile(pdf, tmp, nil, pdfcpu.NewDefaultConfiguration())

	files, err := os.ReadDir(tmp)

	bar := progressbar.New(len(files))

	for _,file := range files {
		f, err := os.Open(tmp + file.Name())
		if err != nil { return err }

		info, err := file.Info()
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate

		writer, err := zWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, f)
		return err

		f.Close()

		bar.Add(1)
	}

	zWriter.Close()
	cFile.Close()

	return nil
}