package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/jwalton/gchalk"
)

var useColour bool = true
var stdoutFlag bool
var level int = 6
var list bool = false

var decompress bool = false

const (
	brightGreen = iota
	brightYellow
	brightBlue
	brightRed
	noColour // Can use to default to no colour output
)

func isGzipped(in *os.File) (gzipped bool, err error) {
	// why 512 bytes ? see http://golang.org/pkg/net/http/#DetectContentType
	buff := make([]byte, 512)

	_, err = in.Seek(0, io.SeekStart)
	if err != nil {
		return
	}
	_, err = in.Read(buff)
	if err != nil {
		return
	}
	_, err = in.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	filetype := http.DetectContentType(buff)

	switch filetype {
	case "application/x-gzip", "application/zip":
		return true, err
	default:
		return false, err
	}
}

func gZipData(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err = gz.Write(data)
	if err != nil {
		return
	}

	if err = gz.Flush(); err != nil {
		return
	}

	if err = gz.Close(); err != nil {
		return
	}

	compressedData = b.Bytes()

	return
}

func gZipFromFile(in *os.File, level int) (compressedData []byte, count int, err error) {
	br := bufio.NewReader(in)

	// Find out if reading into a buffer then incrementally writing would work
	data, err := ioutil.ReadAll(br)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	bb := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(bb)
	defer gzipWriter.Close()

	count, err = gzipWriter.Write(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	gzipWriter.Flush()

	compressedData = bb.Bytes()

	return
}

func gZipToFile(in *os.File, out *os.File, level int) (count int, err error) {
	br := bufio.NewReader(in)

	// Find out if reading into a buffer then incrementally writing would work
	data, err := ioutil.ReadAll(br)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 0, err
	}

	gzipWriter := gzip.NewWriter(out)
	defer gzipWriter.Close()

	count, err = gzipWriter.Write(data)

	return
}

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func gUnzipFromFile(in *os.File) (resData []byte, count int, err error) {
	gzipped, err := isGzipped(in)
	if err != nil {
		return []byte{}, 0, err
	}
	if gzipped == false {
		return resData, 0, fmt.Errorf("file not gzipped %s", in.Name())
	}

	// buf := new(bytes.Buffer)
	// reader, _ := gzip.NewReader(buf)
	gzipReader, err := gzip.NewReader(in)
	defer gzipReader.Close()

	// Find out if reading into a buffer then incrementally writing would work
	data, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return resData, 0, err
	}

	buf := new(bytes.Buffer)
	count, err = buf.Write(data)
	if err != nil {
		return
	}

	resData = buf.Bytes()

	return
}

func gUnzipToFile(in *os.File, out *os.File) (count int, err error) {
	gzipped, err := isGzipped(in)
	if err != nil {
		return
	}
	if gzipped == false {
		return 0, fmt.Errorf("file not gzipped %s", in.Name())
	}

	data, count, err := gUnzipFromFile(in)
	if err != nil {
		return
	}
	bw := bufio.NewWriter(out)
	bw.Write(data)
	bw.Flush()

	return
}

func colour(colour int, input ...string) (str string) {
	str = fmt.Sprint(strings.Join(input, " "))
	str = strings.Replace(str, "  ", " ", -1)

	if !useColour {
		return
	}

	// Choose colour for output or none
	switch colour {
	case brightGreen:
		return gchalk.BrightGreen(str)
	case brightYellow:
		return gchalk.BrightYellow(str)
	case brightBlue:
		return gchalk.BrightBlue(str)
	case brightRed:
		return gchalk.BrightRed(str)
	default:
		return str
	}
}

func checkPath(path string) error {
	var err error
	if _, err = os.Stat(path); err != nil {
		return err
	} else if os.IsNotExist(err) {
		return err
	} else if err != nil {
		return err
	}

	return err
}

func openFile(path string) (*os.File, error) {
	err := checkPath(path)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func createFile(path string) (*os.File, error) {
	// file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return nil, err
	// }
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// printHelp print out simple help output
func printHelp(out *os.File) {
	fmt.Fprintln(out, colour(brightGreen, os.Args[0], "- a Go version of gzip"))
	fmt.Fprintln(out, "Usage")
	// Prints to stdout
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	var helpFlag bool
	flag.BoolVar(&helpFlag, "h", false, "print usage")

	flag.BoolVar(&stdoutFlag, "c", false, "send to standard out")
	flag.BoolVar(&stdoutFlag, "stdout", false, "send to standard out")

	flag.IntVar(&level, "l", 6, "compression level")

	var test bool
	flag.BoolVar(&test, "t", false, "test compressed file integrity")

	var keep bool
	flag.BoolVar(&keep, "k", false, "keep original file")

	flag.BoolVar(&decompress, "d", false, "decompress input")
	flag.BoolVar(&decompress, "decompress", false, "decompress input")

	flag.BoolVar(&list, "L", false, "list compression information")

	flag.Parse()

	paths := flag.Args()

	if helpFlag {
		printHelp(os.Stdout)
	}

	var goodPaths = make([]string, 0, len(paths))

	for _, path := range paths {
		var skip bool
		if _, err := os.Stat(path); err != nil {
			fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			skip = true
		} else if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			skip = true
		} else {
			fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			skip = true
		}
		if skip {
			continue
		}
		goodPaths = append(goodPaths, path)

		fmt.Println(paths)
	}

	// There are files to compress
	if !stdoutFlag && len(goodPaths) > 0 {
		// files := make([]*os.File, 0, len(paths))
		for _, p := range goodPaths {
			var fname string = p + ".gz"
			file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
				continue
			}
			fmt.Println(file.Name())
		}
	}

	if stdoutFlag && len(goodPaths) > 0 {
		fmt.Fprintln(os.Stderr, colour(brightRed, errors.New("files specified along with stdout").Error()))
		os.Exit(1)
	} else if stdoutFlag {

		// Use stdin if available, otherwise exit.
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, _, err := gZipFromFile(os.Stdin, level)
			if err != nil {
				fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
				os.Exit(1)
			}

			reader := bytes.NewReader(data)
			io.CopyBuffer(os.Stdout, reader, data)
		}
	}
}
