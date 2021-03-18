package main

import (
	// "embed"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// "text/template"
	"time"

	"github.com/discord/lilliput"
)

var EncodeOptions = map[string]map[int]int{
	".jpeg": {lilliput.JpegQuality: 85},
	// ".png":  map[int]int{lilliput.PngCompression: 7},
	// ".webp": map[int]int{lilliput.WebpQuality: 85},
}

// content holds our static web server content.
// go:embed img/*
// var content embed.FS

func main() {
	http.HandleFunc("/", receiveImage)
	http.HandleFunc("/status", health)
	fs := http.FileServer(http.Dir("/Users/scott/dev/guppy/img"))
	http.Handle("/img/", http.StripPrefix("/img", fs))

	// http.Handle("/img/", http.StripPrefix("/img", http.FileServer(http.FS(content))))

	// template.ParseFS(content, "*")

	// fmt.Printf("%v", content)

	err := http.ListenAndServe("0.0.0.0:3000", nil)
	if err != nil {
		fmt.Println("error")
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	codeParams, ok := r.URL.Query()["code"]
	if ok && len(codeParams) > 0 {
		statusCode, _ := strconv.Atoi(codeParams[0])
		if statusCode >= 200 && statusCode < 600 {
			w.WriteHeader(statusCode)
		}
	}
	fmt.Fprintf(w, "Glub Glub")
}

func receiveImage(w http.ResponseWriter, r *http.Request) {
	err := uploadImage(r)
	if err != nil {
		http.Error(w, "Invalid Data", http.StatusBadRequest)
		return
	}
}

func uploadImage(r *http.Request) error {
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := os.OpenFile("/Users/scott/dev/guppy/img/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(f, file)

	compress("/Users/scott/dev/guppy/img/"+handler.Filename, "4k.jpg", 3840, 2160)
	compress("/Users/scott/dev/guppy/img/"+handler.Filename, "1920.jpg", 1920, 1080)
	compress("/Users/scott/dev/guppy/img/"+handler.Filename, "1280.jpg", 1280, 720)

	return nil
}

func compress(inputFilename string, outputFilename string, outputWidth int, outputHeight int) {
	// var inputFilename string
	// var outputWidth int
	// var outputHeight int
	// var outputFilename string
	// var stretch bool

	// flag.StringVar(&inputFilename, "input", "", "name of input file to resize/transcode")
	// flag.StringVar(&outputFilename, "output", "", "name of output file, also determines output type")
	// flag.IntVar(&outputWidth, "width", 0, "width of output file")
	// flag.IntVar(&outputHeight, "height", 0, "height of output file")
	// flag.BoolVar(&stretch, "stretch", false, "perform stretching resize instead of cropping")
	// flag.Parse()

	if inputFilename == "" {
		fmt.Printf("No input filename provided, quitting.\n")
		flag.Usage()
		os.Exit(1)
	}

	// decoder wants []byte, so read the whole file into a buffer
	inputBuf, err := ioutil.ReadFile(inputFilename)
	if err != nil {
		fmt.Printf("failed to read input file, %s\n", err)
		os.Exit(1)
	}

	decoder, err := lilliput.NewDecoder(inputBuf)
	// this error reflects very basic checks,
	// mostly just for the magic bytes of the file to match known image formats
	if err != nil {
		fmt.Printf("error decoding image, %s\n", err)
		os.Exit(1)
	}
	defer decoder.Close()

	header, err := decoder.Header()
	// this error is much more comprehensive and reflects
	// format errors
	if err != nil {
		fmt.Printf("error reading image header, %s\n", err)
		os.Exit(1)
	}

	// print some basic info about the image
	fmt.Printf("file type: %s\n", decoder.Description())
	fmt.Printf("%dpx x %dpx\n", header.Width(), header.Height())

	if decoder.Duration() != 0 {
		fmt.Printf("duration: %.2f s\n", float64(decoder.Duration())/float64(time.Second))
	}

	// get ready to resize image,
	// using 8192x8192 maximum resize buffer size
	ops := lilliput.NewImageOps(8192)
	defer ops.Close()

	// create a buffer to store the output image, 50MB in this case
	outputImg := make([]byte, 50*1024*1024)

	// use user supplied filename to guess output type if provided
	// otherwise don't transcode (use existing type)
	outputType := "." + strings.ToLower(decoder.Description())
	if outputFilename != "" {
		outputType = filepath.Ext(outputFilename)
	}

	if outputWidth == 0 {
		outputWidth = header.Width()
	}

	if outputHeight == 0 {
		outputHeight = header.Height()
	}

	resizeMethod := lilliput.ImageOpsFit
	// if stretch {
	// 	resizeMethod = lilliput.ImageOpsResize
	// }

	if outputWidth == header.Width() && outputHeight == header.Height() {
		resizeMethod = lilliput.ImageOpsNoResize
	}

	opts := &lilliput.ImageOptions{
		FileType:             outputType,
		Width:                outputWidth,
		Height:               outputHeight,
		ResizeMethod:         resizeMethod,
		NormalizeOrientation: true,
		EncodeOptions:        EncodeOptions[outputType],
	}

	// resize and transcode image
	outputImg, err = ops.Transform(decoder, opts, outputImg)
	if err != nil {
		fmt.Printf("error transforming image, %s\n", err)
		os.Exit(1)
	}

	// image has been resized, now write file out
	if outputFilename == "" {
		outputFilename = "resized" + filepath.Ext(inputFilename)
	}

	inputType := filepath.Ext(inputFilename)
	splitInput := strings.Split(inputFilename, inputType)

	builtOutputName := splitInput[0] + "-" + outputFilename

	if _, err := os.Stat(builtOutputName); !os.IsNotExist(err) {
		fmt.Printf("output filename %s exists, quitting\n", builtOutputName)
		os.Exit(1)
	}

	err = ioutil.WriteFile(builtOutputName, outputImg, 0644)
	if err != nil {
		fmt.Printf("error writing out resized image, %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("image written to %s\n", builtOutputName)
}

func compressAndResize() {

}
