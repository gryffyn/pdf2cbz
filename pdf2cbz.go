package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/gen2brain/go-fitz"
	"github.com/jessevdk/go-flags"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

const tmpl = `{{ string . "prefix" | green}} {{counters . }} {{ bar . ("[" | green) ("=" | green) (">" | green) ("." | red) ("]" | green)}} {{percent . }} {{ string . "suffix" | green}}`
var info bool

func main() {
	type Opts struct {
		Images         bool   `short:"i" long:"images" description:"Extracts images from PDF (default: extracts pages)"`
		PNG            bool   `short:"p" long:"png" description:"Outputs pages as PNG (default: jpeg)"`
		CropDimensions string `short:"c" long:"crop" description:"Dimensions of the crop region (css shorthand, comma-delimited) in percent. Ex. '10,15,20,25' means 10% off the top, 15% off the right, so on"`
		JPEGQuality    int    `short:"q" long:"quality" description:"JPEG quality (0-100)" default:"85"`
		Debug          bool   `long:"debug" description:"enable debug printing"`
		Positional     struct {
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

	// set up logging
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if opts.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if !flags.WroteHelp(err) {
		if opts.Images {
			err = extractImages(opts.Positional.PDF, opts.Positional.CBZ)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			err = extractPages(opts.Positional.PDF, opts.Positional.CBZ, opts.PNG, opts.JPEGQuality, opts.CropDimensions)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func extractPages(pdf, cbz string, usePNG bool, jpegQuality int, crop string) error {
	doc, err := fitz.New(pdf)
	if err != nil {
		return err
	}

	defer doc.Close()

	cFile, err := os.Create(cbz)
	if err != nil {
		return err
	}

	zWriter := zip.NewWriter(cFile)

	bar := pb.ProgressBarTemplate(tmpl).Start(doc.NumPage()-1)

	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			log.Println(err)
		}

		filetype := "jpg"
		if usePNG {
			filetype = "png"
		}

		h := &zip.FileHeader{Name: fmt.Sprintf("page%04d.%s", n, filetype), Method: zip.Deflate, Modified: time.Now()}
		f, err := zWriter.CreateHeader(h)
		if err != nil {
			return err
		}

		if crop != "" {
			simg := cropImage(img, crop)
			if usePNG {
				err = png.Encode(f, simg)
				if err != nil {
					log.Println(err)
				}
			} else {
				err = jpeg.Encode(f, simg, &jpeg.Options{Quality: jpegQuality})
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			if usePNG {
				err = png.Encode(f, img)
				if err != nil {
					log.Println(err)
				}
			} else {
				err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpegQuality})
				if err != nil {
					log.Println(err)
				}
			}
		}

		bar.Set("suffix", path.Base(cbz)).Increment()
	}

	zWriter.Close()
	cFile.Close()

	return nil
}

func extractImages(pdf, cbz string) error {
	cFile, err := os.Create(cbz)
	if err != nil {
		return err
	}
	zWriter := zip.NewWriter(cFile)

	tmp, err := os.MkdirTemp("", "pdf2cbz")
	if err != nil {
		return err
	}

	err = api.ExtractImagesFile(pdf, tmp, nil, pdfcpu.NewDefaultConfiguration())
	if err != nil { return err }

	files, err := os.ReadDir(tmp)
	if err != nil { return err }

	bar := pb.ProgressBarTemplate(tmpl).Start(len(files)-1)

	for _, file := range files {
		f, err := os.Open(tmp + file.Name())
		if err != nil {
			return err
		}

		info, err := file.Info()
		if err != nil { return err }
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
		if err != nil { return err }

		f.Close()

		bar.Set("suffix", path.Base(cbz)).Increment()
	}

	zWriter.Close()
	cFile.Close()

	return nil
}

func cropImage(img image.Image, dims string) image.Image {
	dimsSlice := strSliceToFloat(strings.Split(dims, ","))

	flWidth := float64(img.Bounds().Dx())
	flHeight := float64(img.Bounds().Dy())
	
	topPer := int(math.Round(flHeight * (dimsSlice[0] / 100)))
	rightPer := int(math.Round(flWidth - (flWidth * (dimsSlice[1] / 100))))
	bottomPer := int(math.Round(flHeight - (flHeight * (dimsSlice[2] / 100))))
	leftPer := int(math.Round(flWidth * (dimsSlice[3] / 100)))

	minx := img.Bounds().Min.X
	miny := img.Bounds().Min.Y
	maxx := img.Bounds().Max.X
	maxy := img.Bounds().Max.Y

	cropRegion := image.Rect(leftPer, topPer, rightPer, bottomPer)

	if getDebug() && !info {
		zlog.Debug().Msgf("ORIGDIMS: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		zlog.Debug().Msgf("NEWDIMS: %dx%d", cropRegion.Dx(), cropRegion.Dy())
		zlog.Debug().Msgf("CROP PERCENT: T: %d, R: %d, B: %d, L: %d", topPer, rightPer, bottomPer, leftPer)
		zlog.Debug().Msgf("ORIGINAL: (%d,%d) x (%d,%d)", minx, miny, maxx, maxy)
		zlog.Debug().Msgf("NEWCROP: (%d,%d) x (%d,%d)\n",
			leftPer, topPer, rightPer, bottomPer)
		info = true
	}

	crop := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRegion)

	return crop
}

func strSliceToFloat(strs []string) []float64 {
	var fl []float64
	for _,v := range strs {
		i, _ := strconv.Atoi(v)
		fl = append(fl, float64(i))
	}
	return fl
}

func getDebug() bool {
	e := zlog.Debug()
	return e.Enabled()
}
