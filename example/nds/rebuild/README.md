# Rebuild

Takes apart a ROM file and rebuilds it.
Not guaranteed to work for all ROM files:
* I don't have a full understanding of overlay files, I might be missing something
* If the binary contains hard-coded ROM addresses, this wont work
* If either CPU's entrypoint is within the other's binary
* If any of the ROMs building blocks overlap (FAT/FNT/banner embedded in binaries, etc.)

Tested working with `MARIO&LUIGI2_ARMP01_00.nds`.

## Usage:
`go run github.com/sukus21/nintil/example/nds/rebuild <path-to-rom>`
