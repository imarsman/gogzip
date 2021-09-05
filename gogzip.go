package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jwalton/gchalk"
)

var useColour bool = true
var stdoutFlag bool

const (
	brightGreen = iota
	brightYellow
	brightBlue
	brightRed
	noColour // Can use to default to no colour output
)

func decompress(in *os.File, out *os.File, level int) error {
	buf := make([]byte, 2048)
	var readWriter *bufio.ReadWriter

	gzipReader, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	br := bufio.NewReader(gzipReader)

	// gzipWriter := gzip.NewWriter(out)
	// gzipWriter, err = gzip.NewWriterLevel(gzipWriter, level)
	// if err != nil {
	// 	return err
	// }
	bw := bufio.NewWriter(out)
	readWriter = bufio.NewReadWriter(br, bw)

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
		readWriter.Write(buf[0:n])
		// The write method for fileWriter.write does flush.
		readWriter.Flush()
	}

	return nil
}

func compress(in *os.File, out *os.File, level int) error {
	buf := make([]byte, 2048)
	var readWriter *bufio.ReadWriter

	// gzipReader, err := gzip.NewReader(in)
	// if err != nil {
	// 	return err
	// }
	br := bufio.NewReader(in)

	gzipWriter := gzip.NewWriter(out)
	gzipWriter, err := gzip.NewWriterLevel(gzipWriter, level)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(gzipWriter)
	readWriter = bufio.NewReadWriter(br, bw)

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
		readWriter.Write(buf[0:n])
		// The write method for fileWriter.write does flush.
		readWriter.Flush()
	}

	return nil
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

	flag.BoolVar(&stdoutFlag, "S", false, "do not forward standard input to standard output")

	flag.Parse()

}
