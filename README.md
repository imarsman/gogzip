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
* `gogzip -c` - gzip standard input and send to standard output
* `gogzip --stdout` - gzip standard input and send to standard output
* `gogzip -d <file1> <file2>` - decompress instead of gzip - not implemented yet
* `gogzip -k <file1> <file2>` - do not delete original file
* `gogzip -l <file1> <file2>` - use compression level 1-9
* `gogzip -q <file1> <file2>` - quiet output (no complaints for invalid input)
* `gogzip -t <file1> <file2>` - test compressed file integrity - not implemented
  yet

  -- Ian Marsman