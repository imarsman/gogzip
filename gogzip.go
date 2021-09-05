package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
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

func isGzipped(in *os.File) (bool, error) {
	defer in.Close()
	buff := make([]byte, 512)

	// why 512 bytes ? see http://golang.org/pkg/net/http/#DetectContentType

	_, err := in.Seek(0, io.SeekStart)
	if err != nil {
		return false, err
	}
	_, err = in.Read(buff)
	if err != nil {
		return false, err
	}
	_, err = in.Seek(0, io.SeekStart)
	if err != nil {
		return false, err
	}

	filetype := http.DetectContentType(buff)

	switch filetype {
	case "application/x-gzip", "application/zip":
		return true, nil
	default:
		return false, nil
	}
}

func gUnzip(in *os.File, out *os.File, level int) (int, error) {
	var written int
	buf := make([]byte, 2048)
	var readWriter *bufio.ReadWriter

	gzipReader, err := gzip.NewReader(in)
	if err != nil {
		return 0, err
	}
	br := bufio.NewReader(gzipReader)

	bw := bufio.NewWriter(out)
	readWriter = bufio.NewReadWriter(br, bw)

	defer gzipReader.Close()

	buf = make([]byte, 2048)

	for {
		n, err := readWriter.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, err.Error())
			break
		}
		if n == 0 && err == io.EOF {
			break
		}
		n, err = readWriter.Write(buf[0:n])
		if err != nil {
			return 0, err
		}
		written += n

		// The write method for fileWriter.write does flush.
		readWriter.Flush()
	}

	_, err = in.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return written, nil
}

func gZip(in *os.File, out *os.File, level int) (int, error) {
	var written int

	buf := make([]byte, 2048)
	var readWriter *bufio.ReadWriter

	br := bufio.NewReader(in)

	gzipWriter := gzip.NewWriter(out)
	gzipWriter, err := gzip.NewWriterLevel(gzipWriter, level)
	if err != nil {
		return 0, err
	}
	bw := bufio.NewWriter(gzipWriter)

	readWriter = bufio.NewReadWriter(br, bw)

	defer gzipWriter.Close()

	buf = make([]byte, 2048)

	for {
		n, err := readWriter.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, err.Error())
			break
		}
		if n == 0 && err == io.EOF {
			break
		}
		n, err = readWriter.Write(buf[0:n])
		written += n
		// The write method for fileWriter.write does flush.
		readWriter.Flush()
	}

	_, err = in.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return written, nil
}

func colour(colour int, input ...string) string {
	str := fmt.Sprint(strings.Join(input, " "))
	str = strings.Replace(str, "  ", " ", -1)

	if !useColour {
		return str
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

	flag.IntVar(&level, "l", 1, "compression level")

	flag.BoolVar(&decompress, "c", false, "decompress input")

	flag.BoolVar(&list, "L", false, "list compression information")

	flag.Parse()

}
