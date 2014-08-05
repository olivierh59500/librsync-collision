# rdiff-collision-maker
This is a simple tool to generate collisions for rdiff/librsync using a
birthday attack -- this doesn't try to do anything clever with md4 at
all - any hash that's truncated to 8 bytes will work just the same.

## WARNING
The default settings use nearly 96G of RAM (achievement unlocked - 't'
in the RES field in top).  You may need to close Firefox before
running, or possibly tune it down a bit (eg. SPACE_WF=(1<<24) and
HASH_TRUNC=6).

## Configuration
Edit the constants in config.go - for example if you don't have enough
RAM you can decrease SPACE_WF and increates TIME_WF - this will make it
slow, but fit in RAM.  Do not increase SPACE_WF to more than half of
the bitlength -- for HASH_TRUNC=8, don't do more than (1<<32).

TIME_WF should be at-least the total bitlength of the hash, divided by
SPACE_WF, but add an extra bit to make sure - sometimes it's just
unlucky.

STORE_PROCS originated from the fact that I couldn't make a hashtable of
size (1<<31) in golang, so I did memcached-style multiple threads, with
the first byte of the hash deciding which one to go to.  This can't be
more than 256 without some changes, but I found 4 gave a good spread of
CPU load - so maybe just stick with that.

Currently the prefixes are hard-coded into main.go - these must have the
same rsync checksum, and be blocksize-20 in size (2028 for rdiff - other
utilities such as duplicity use different block sizes).

