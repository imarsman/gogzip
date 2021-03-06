# gogzip
An implementation of gzip in Go.

gzip was first publicly released in  31 October 1992. It was developed as part
as free software under the GNU licence.

The algorithm was written to allow for decompression while streaming. I have not
been able to do streaming (buffered) compression. It may be possible but I have
not yet found the answer to this issue.

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
* `gogzip -k <file1> <file2>` - do not delete original file
* `gogzip -c` - gzip standard input and send to standard output - with no files
  specified and with uncompressed stdin the `-c` flag is set to true
  * `gogzip --stdout` - gzip standard input and send to standard output
* `gogzip -d <file1> <file2>` - decompress instead of gzip
* `gogzip -l <file1> <file2>` - use compression level `0-9` with 0 being no compression
* `gogzip -q <file1> <file2>` - quiet output (no complaints for invalid input)
* `gogzip -L <file1> <file2>` - give summary information for listed files
* `gogzip -t <file1> <file2>` - test file list integrity
* `gogzip -f <file1> <file2>` - force overwrite etc. - implemented when file
  would be overriden

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

# Simple tests

As currently implemented, BSD gzip is faster.

```
$: time for i in {1..1000}; do ./gogzip -k sample/1.txt; done

real	0m4.684s
user	0m2.125s
sys	0m2.014s
```

```
$: time for i in {1..1000}; do gzip -k -f sample/1.txt; done

real	0m2.616s
user	0m0.943s
sys	0m1.517s
```

## Notes

The argument parsing library used here does not deal with arguments such as -1,
-2, -, etc. It may be that an argument will need to have a different identifier to
work around this.

  -- Ian Marsman
  