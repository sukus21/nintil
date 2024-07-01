# RLZ

A custom(?) compression scheme, used on some files in Partners in Time.
Combines the strengths of RLE compression with LZ-style compression.
The format also encodes the total length of the decompressed data.
This allows programmers to pre-allocate memory before decompressing.

The rewind buffer is still 0x1000 bytes long, as far as I can tell.
In LZ mode, only 2-17 bytes can be read from the rewind buffer at a time.
In RLE mode, a byte can be repeated between 2 and 257 times total.

RLZ is not the official name for this algorithm, it is just a name I made up.
If anyone knows the algorithms real name, please open an issue and ley me know!
