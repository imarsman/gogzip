package main

import (
	// "exec"
	"bytes"
	"os/exec"
	"testing"

	"github.com/matryer/is"
)

//                Tests and benchmarks
// -----------------------------------------------------
// benchmark
//   go test -run=XXX -bench=. -benchmem
// Get allocation information and pipe to less
//   go build -gcflags '-m -m' ./*.go 2>&1 |less
// Run all tests
//   go test -v
// Run one test and do allocation profiling
//   go test -run=XXX -bench=IterativeISOTimestampLong -gcflags '-m' 2>&1 |less
// Run a specific test by function name pattern
//  go test -run=TestParsISOTimestamp
//
//  go test -run=XXX -bench=.
//  go test -bench=. -benchmem -memprofile memprofile.out -cpuprofile cpuprofile.out
//  go tool pprof -http=:8080 memprofile.out
//  go tool pprof -http=:8080 cpuprofile.out

func runCmd(command string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return err
}

func cleanup() error {
	err := runCmd("script/reset.sh")

	return err
}

func TestStart(t *testing.T) {
	cleanup()
}
func TestRunCmd(t *testing.T) {
	is := is.New(t)

	err := runCmd("script/reset.sh")
	is.NoErr(err)
}

func TestRunCheckFile(t *testing.T) {
	is := is.New(t)

	err := checkPath("sample/1.txt")
	is.NoErr(err)
}

func TestRunGetFile(t *testing.T) {
	is := is.New(t)

	file, err := openFile("sample/1.txt")
	is.NoErr(err)
	is.True(file != nil)
	t.Log(file.Name())
}

func TestCreateFile(t *testing.T) {
	is := is.New(t)

	file, err := createFile("sample/1.txt.gz")
	is.NoErr(err)
	is.True(file != nil)
	t.Log(file.Name())
}

func TestEnd(t *testing.T) {
	cleanup()
}
