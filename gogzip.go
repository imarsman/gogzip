package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
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

func isGzipped(in *os.File) (bool, error) {
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

func gZip(in *os.File, out *os.File, level int) (int, error) {
	defer in.Close()
	defer out.Close()

	br := bufio.NewReader(in)

	// Find out if reading into a buffer then incrementally writing would work
	data, err := ioutil.ReadAll(br)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 0, err
	}

	gzipWriter := gzip.NewWriter(out)
	defer gzipWriter.Close()

	n, err := gzipWriter.Write(data)

	return n, nil
}

func gUnzip(in *os.File) ([]byte, int, error) {
	gzipped, err := isGzipped(in)
	if err != nil {
		return []byte{}, 0, err
	}
	if gzipped == false {
		return []byte{}, 0, fmt.Errorf("file not gzipped %s", in.Name())
	}

	gzipReader, err := gzip.NewReader(in)
	defer gzipReader.Close()

	// Find out if reading into a buffer then incrementally writing would work
	data, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return []byte{}, 0, err
	}

	buf := new(bytes.Buffer)
	n, err := buf.Write(data)

	return buf.Bytes(), n, nil
}

func gUnzipToFile(in *os.File, out *os.File) (int, error) {
	gzipped, err := isGzipped(in)
	if err != nil {
		return 0, err
	}
	if gzipped == false {
		return 0, fmt.Errorf("file not gzipped %s", in.Name())
	}

	data, n, err := gUnzip(in)
	if err != nil {
		return 0, err
	}
	bw := bufio.NewWriter(out)
	bw.Write(data)
	bw.Flush()

	return n, nil
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
			fmt.Println(err)
			skip = true
		} else if os.IsNotExist(err) {
			fmt.Println(err)
			skip = true
		} else {
			fmt.Println(err)
			skip = true
		}
		if skip {
			continue
		}
		goodPaths = append(goodPaths, path)

		fmt.Println(paths)
	}

	// files := make([]*os.File, 0, len(paths))
	for _, p := range goodPaths {
		var fname string = p + ".gz"
		file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(file.Name())
	}

	// Use stdin if available, otherwise exit.
	stat, _ := os.Stdin.Stat()
	if (stat.Mode()&os.ModeCharDevice) == 0 && 1 == 2 {
		br := bufio.NewReader(os.Stdin)

		bb := new(bytes.Buffer)
		bw := bufio.NewWriter(bb)

		rw := bufio.NewReadWriter(br, bw)

		nBytes, nChunks := int64(0), int64(0)
		buf := make([]byte, 0, 4*1024)
		for {
			n, err := rw.Read(buf[:cap(buf)])
			buf = buf[:n]

			if n == 0 {
				if err == nil {
					continue
				}
				if err == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}

			rw.Write(buf[:n])
			bw.Flush()

			nChunks++
			nBytes += int64(len(buf))

			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
		io.Copy(gzip.NewWriter(os.Stdout), bb)
	} else if 1 == 2 {
	}
}
