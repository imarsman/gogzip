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

// needed when reading stdin which doesn't allow seek
func isGzippedFromBytes(data []byte) (gzipped bool, err error) {
	// 	filetype = http.DetectContentType(data[0 : len(data)-1])
	if len(data) < 2 {
		return gzipped, errors.New("not enough bytes to detect")
	}

	// Look at two first bytes
	if data[0] == 0x1F && data[1] == 0x8B {
		gzipped = true
	}

	return
}

// Look in file for gzip magic header
func isGzipped(in *os.File, seek bool) (gzipped bool, err error) {
	data := make([]byte, 512)

	if seek {
		_, err = in.Seek(0, io.SeekStart)
		if err != nil {
			return
		}
	}
	_, err = in.Read(data)
	if err != nil {
		return
	}
	if seek {
		_, err = in.Seek(0, io.SeekStart)
		if err != nil {
			return
		}
	}

	gzipped, err = isGzippedFromBytes(data)

	return
}

// Look in file for gzip magic header
func testGzipped(in *os.File, seek bool) (gzipped bool, err error) {
	gzipped, err = isGzipped(in, true)
	if err != nil {
		return false, err
	}
	if !gzipped {
		return false, fmt.Errorf("file %s not gzipped", in.Name())
	}

	_, _, err = gUnzipFromFile(in)
	if err != nil {
		return false, err
	}

	return
}

// https://gist.github.com/alex-ant/aeaaf497055590dacba760af24839b8d
func gZipData(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gzipWriter := gzip.NewWriter(&b)

	_, err = gzipWriter.Write(data)
	if err != nil {
		return
	}

	if err = gzipWriter.Flush(); err != nil {
		return
	}

	if err = gzipWriter.Close(); err != nil {
		return
	}

	compressedData = b.Bytes()

	return
}

// read from file and return gzipped byte slice
func gZipFromFile(in *os.File, level int) (compressedData []byte, count int, err error) {
	br := bufio.NewReader(in)

	// Find out if reading into a buffer then incrementally writing would work
	data, err := ioutil.ReadAll(br)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	compressedData, err = gZipData(data)
	if err != nil {
		return
	}
	count = len(compressedData)

	return
}

// read from file and write gzipped to output file
func gZipToFile(in *os.File, out *os.File, level int) (count int, err error) {
	// gzipped, err := isGzipped(in, true)
	// if err != nil {
	// 	return count, err
	// }
	// if gzipped == true {
	// 	return count, fmt.Errorf("file already gzipped %s", in.Name())
	// }

	br := bufio.NewReader(in)

	// Find out if reading into a buffer then incrementally writing would work
	data, err := ioutil.ReadAll(br)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 0, err
	}

	compressedData, err := gZipData(data)
	if err != nil {
		return
	}

	bw := bufio.NewWriter(out)
	count, err = bw.Write(compressedData)
	if err != nil {
		return
	}
	bw.Flush()

	return
}

// unzip a slice of bytes
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

// read from file and gunzip to byte slice
func gUnzipFromFile(in *os.File) (resData []byte, count int, err error) {
	// gzipped, err := isGzipped(in, true)
	// if err != nil {
	// 	return []byte{}, 0, err
	// }
	// if gzipped == false {
	// 	return resData, 0, fmt.Errorf("file not gzipped %s", in.Name())
	// }

	br := bufio.NewReader(in)
	compressedData, err := ioutil.ReadAll(br)
	if err != nil {
		return
	}

	resData, err = gUnzipData(compressedData)
	if err != nil {
		return
	}
	count = len(resData)

	return
}

func gUnzipToFile(in *os.File, out *os.File) (count int, err error) {
	// gzipped, err := isGzipped(in, true)
	// if err != nil {
	// 	return
	// }
	// if gzipped == false {
	// 	return 0, fmt.Errorf("file not gzipped %s", in.Name())
	// }

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
		str = gchalk.BrightGreen(str)
	case brightYellow:
		str = gchalk.BrightYellow(str)
	case brightBlue:
		str = gchalk.BrightBlue(str)
	case brightRed:
		str = gchalk.BrightRed(str)
	}

	return
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

func openFile(path string) (file *os.File, err error) {
	err = checkPath(path)
	if err != nil {
		return
	}
	file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return
	}

	return
}

func createFile(path string) (file *os.File, err error) {
	file, err = os.Create(path)
	if err != nil {
		return
	}

	return
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

	var quietFlag bool
	flag.BoolVar(&quietFlag, "q", false, "quiet output")

	flag.BoolVar(&stdoutFlag, "c", false, "send to standard out")
	flag.BoolVar(&stdoutFlag, "stdout", false, "send to standard out")

	flag.IntVar(&level, "l", 6, "compression level (0-9 with 0 being no compression)")

	var test bool
	flag.BoolVar(&test, "t", false, "test compressed file integrity")

	var keepFlag bool
	flag.BoolVar(&keepFlag, "k", false, "keep original file")

	flag.BoolVar(&decompress, "d", false, "decompress input")
	flag.BoolVar(&decompress, "decompress", false, "decompress input")

	flag.BoolVar(&list, "L", false, "list compression information (not implemented yet)")

	flag.Parse()

	// override invalid level
	if level < 0 || level > 9 {
		if !quietFlag {
			fmt.Fprintln(os.Stderr, colour(brightRed, fmt.Sprintf(
				"invalid compression level %d - setting to 6", level),
			))
		}
		level = 6
	}

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
		} else if err != nil {
			fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			skip = true
		}
		if skip {
			continue
		}
		goodPaths = append(goodPaths, path)
	}

	// Test a file at a path
	// allows defer to work in case of error
	var testFile = func(path string) {
		inFile, err := openFile(path)
		defer inFile.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			return
		}
		_, err = testGzipped(inFile, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			return
		}
	}

	// Implement the -t flag to test all input files
	if test {
		if len(goodPaths) > 0 {
			for _, path := range goodPaths {
				testFile(path)
			}
		}
		os.Exit(0)
	}

	if list {
		if len(goodPaths) > 0 {
			var totalCompressed int64 = 0
			totalUncompressed := 0
			var totalRatio float64 = 0
			fmt.Fprintf(os.Stdout,
				"  %10s %12s %8s   %17s\n",
				"compressed",
				"uncompressed",
				"ratio",
				"uncompressed_name",
			)

			count := 0
			for _, path := range goodPaths {
				inFile, err := openFile(path)
				defer inFile.Close()
				if err != nil {
					if !quietFlag {
						fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
					}
					continue
				}

				// check if gzipped and skip if error or if gzipped
				gzipped, err := isGzipped(inFile, true)
				if err != nil {
					if !quietFlag {
						fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
					}
					continue
				}
				if !gzipped {
					// if !quietFlag {
					// 	fmt.Fprintln(os.Stderr, colour(brightRed, fmt.Sprintf("file not gzipped %s", path)))
					// }
					continue
				}

				s, err := os.Stat(path)
				if err != nil {
					if !quietFlag {
						fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
					}
					return
				}
				compressedCount := s.Size()
				_, uncompressedCount, err := gUnzipFromFile(inFile)
				if err != nil {
					if !quietFlag {
						fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
					}
					return
				}
				ratio := float64(compressedCount) / float64(uncompressedCount)

				name := path
				if strings.HasSuffix(path, ".gz") {
					name = name[:len(path)-3]
				}
				ratioStr := fmt.Sprintf("%.2f", ratio*100)
				fmt.Fprintf(os.Stdout,
					"  %10d %12d %7s%%   %-17s\n",
					compressedCount,
					uncompressedCount,
					ratioStr,
					name,
				)
				totalCompressed += compressedCount
				totalUncompressed += uncompressedCount
				totalRatio = totalRatio + ratio

				count++
			}
			if count > 1 {
				ratioStr := fmt.Sprintf("%.2f", totalRatio*100/float64(len(goodPaths)))
				name := "(totals)"
				fmt.Fprintf(os.Stdout,
					"  %10d %12d %7s%%   %-17s\n",
					totalCompressed,
					totalUncompressed,
					ratioStr,
					name,
				)
			}

			os.Exit(0) // exit because we dealt with named files
		}

	}

	// Do this in separate function to have file closed defer to operate
	var process = func(path string) {
		// open the current file
		inFile, err := openFile(path)
		defer inFile.Close()
		if err != nil {
			if !quietFlag {
				fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			}
			return
		}

		// check if gzipped and skip if error or if gzipped
		gzipped, err := isGzipped(inFile, true)
		if err != nil {
			if !quietFlag {
				fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			}
			return
		}
		if gzipped {
			if !quietFlag {
				fmt.Fprintln(os.Stderr, colour(brightRed, fmt.Sprintf("file already gzipped %s", path)))
			}
			return
		}

		// set gzipped name
		var gzipFName string = path + ".gz"

		// open file and show error and skip if error
		gzipFile, err := os.OpenFile(gzipFName, os.O_CREATE|os.O_WRONLY, 0644)
		defer gzipFile.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			return
		}

		// gzip output file from input file at level
		_, err = gZipToFile(inFile, gzipFile, level)
		if err != nil {
			if !quietFlag {
				fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
			}
			return
		}
		// If keep fla false remove start file
		if !keepFlag {
			err = os.Remove(path)
			if err != nil {
				if !quietFlag {
					fmt.Fprintln(os.Stderr, colour(brightRed, err.Error()))
				}
				return
			}
		}
	}

	// MacOS (BSD) gzip sends to stdout if there is stdin but GNU does not
	if !stdoutFlag && len(goodPaths) == 0 {
		stdoutFlag = true
	}

	// There are files to compress
	if !stdoutFlag && len(goodPaths) > 0 {
		// files := make([]*os.File, 0, len(paths))
		for _, path := range goodPaths {
			process(path)
		}
		os.Exit(0) // exit because we dealt with named files
	}

	if stdoutFlag {
		if !quietFlag {
			fmt.Fprintln(os.Stderr, colour(brightRed, errors.New("files specified along with stdout").Error()))
		}
		os.Exit(1)
		// MacOS gzip will forward when stdin available and no -c flag set
	} else if stdoutFlag || len(goodPaths) == 0 {

		// Use stdin if available, otherwise exit.
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			br := bufio.NewReader(os.Stdin)

			// Find out if reading into a buffer then incrementally writing would work
			data, err := ioutil.ReadAll(br)
			if err != nil {
				if !quietFlag {
					fmt.Fprintln(os.Stderr, err.Error())
				}
				return
			}

			gzipped, err := isGzippedFromBytes(data)
			if err != nil {
				if !quietFlag {
					fmt.Fprintln(os.Stderr, err.Error())
				}
				return
			}
			// Don't re-gzip but just send the gzipped data to stdout
			if gzipped {
				reader := bytes.NewReader(data)
				io.CopyBuffer(os.Stdout, reader, data)
				os.Exit(0)
			}
			compressedData, err := gZipData(data)
			if err != nil {
				if !quietFlag {
					fmt.Fprintln(os.Stderr, err.Error())
				}
				return
			}
			reader := bytes.NewReader(compressedData)
			io.CopyBuffer(os.Stdout, reader, compressedData)
			os.Exit(0)
		}
	}
}
