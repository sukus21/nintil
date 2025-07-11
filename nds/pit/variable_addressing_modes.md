# Variable addressing modes

When reading a variable from memory, the top 4 bits of the variable are not used for the address, but for addressing mode.
Here is a list of known addressing modes:

## `0x0000`: 0
This addressing mode always returns 0.

## `0x1000`: value from manager
Reads a value from the script manager.
Maybe from the stack? not sure...
`return (u32*)(scriptManager)[1 + variable];`

## `0x2000`: boolean flag
Reads a single bit from a bitlist.
Almost identical to `0xE000` + `0xF000`, but only has 12 bits for bit index.
The bitlist read from is also not the same.

## `0x3000`: field specific
Calls to a function that is only valid when overlay 0 (field program) is loaded.
If jumped to in battle, will probably crash the game completely.
Unknown exactly what happens and what is returned.

## `0x4000`: battle specific
Same as `0x3000`, but for battles.
Unknown what is returned from this method (more research needed).



## `0xE000` and `0xF000`: boolean flag
Reads a single bit from a bitlist.
The purpose of this bitlist is currently unknown.
The value in the lower 13 (yes 13, not 12. That's why both E and F) is the bit index.
