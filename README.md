# rdiff-collision-maker
This is a simple tool to generate collisions for rdiff/librsync using a
birthday attack -- this doesn't try to do anything clever with md4 at
all - any hash that's truncated to 8 bytes will work just the same.

## Configuration
Edit config.go to tweak parameters - the ones in there now work fine for
8-byte (64-bit) output.

Currently the prefixes are hard-coded into main.go - these must have the
same rsync checksum, and be blocksize-20 in size (2028 for rdiff - other
utilities such as duplicity use different block sizes).
