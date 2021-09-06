# gogzip
An implementation of gzip in Go

This is not finished. My goal with this is to improve my knowledge of i/o,
readers, writers, files, and compression.

I have tried to implement buffered gzipping but so far the output produced by
this has been invalid. This means that each file is read in fully then written
out. 

This implementation will handle compression and decompression. I will have to
look closely at the original specification and decide what would make sense to
implement. 

## Usage

* `gogzip -h` - print usage
* `gogzip <file1> <file2>` - gzip files and delete originals
* `gogzip -c` - gzip standard input and send to standard output - with no files
  specified and with uncompressed stdin the `-c` flag is set to true
* `gogzip --stdout` - gzip standard input and send to standard output
* `gogzip -d <file1> <file2>` - decompress instead of gzip - not implemented yet
* `gogzip -k <file1> <file2>` - do not delete original file
* `gogzip -l <file1> <file2>` - use compression level `0-9` with 0 being no compression
* `gogzip -q <file1> <file2>` - quiet output (no complaints for invalid input)
* `gogzip -L <file1> <file2>` - give summary information for listed files
* `gogzip -t <file1> <file2>` - test file list integrity

## Notes

BSD gzip will forward to stdout if there is non-gzipped stdin. GNU gzip will
not. I have implemented gogzip to behave like BSD in this way.

## Building and Running

The app can be built by typing the command below. If you have an older version
of Go installed you can change the version number in go.mod if there is a
complaint on trying to compile. This should be compatible with earlier versions
of Go like 1.14 and 1.15 though I have not checked. This app does not use
embedding, which appeared in Go 1.16.

`go build .`

The app can be run without building by typing

`go run gogzip.go`

There are some tests which can be run by typing

`go test -v .`

  -- Ian Marsman