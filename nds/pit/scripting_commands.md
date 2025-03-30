# VM commands

All instructions return a status of `0` to the interpreter, unless otherwise noted.

## Pseudocode notes
data types:
* sXX = signed XX-bit int
* uXX = unsigned XX-bit int
* iXX = XX-bit int, can be either signed or unsigned

## Common commands (CM_*)

### `0x0000`: CM_TERMINATE
`CM_TERMINATE(): void`

Terminates a script completely.
Sets the script program counter to `0`.
Always returns status `1` to the interpreter.

### `0x0001`: CM_RETURN
`CM_RETURN(): void`

Pops a value from the stack, and sets the program counter to that value.
If the stack is empty, returns status `2` to the interpreter.

### `0x0002`: CM_BRANCH
`CM_BRANCH(withLink: bool, offset: s32): void`

If the `withLink` flag is `1`, the program counter is pushed to the stack BEFORE branching.
Adds `offset` to the program counter.

### `0x0003`: CM_WAIT
`CM_WAIT(numFrames: u16): void`

Sets the VM's wait counter to the specified value.
Always returns status `3` to the interpreter.

### `0x0004`: CM_JUMP_COND
`CM_JUMP_COND(conditionType: u8, a: s32, b: s32, expectedResult: bool, offset: s32): void`

Adds `offset` to the program counter, if the specified criteria are met:
```GO
if VM_checkJumpCondition(conditionType, a, b) == expectedResult {
    pc += offset
}
```

See [this list](#jump-conditions) for details on the different jump conditions.

### `0x0005`: CM_UNUSED_0005
Does nothing

### `0x0006`: CM_UNUSED_0006
Does nothing

### `0x0007`: CM_UNUSED_0007
Does nothing

### `0x0008`: CM_UNUSED_0008
Does nothing

### `0x0009`: CM_UNKNOWN_0009
`CM_UNKNOWN_0009(unknown: u32): void`

Not currently known what this does.
It reads a function pointer from an address in the ARM9 binary, and calls it twice with different inputs.
The function it calls does nothing (it just returns), and the main field / battle overlays don't reference the pointer address.

It also does a lot of pointless arithmetic that doesn't amount to anything...
Maybe it was used for debugging during development?

### `0x000A`: CM_PUSH
`CM_PUSH(value: u32): void`

Pushes the given value onto the stack.

### `0x000B`: CM_POP
`CM_POP(): u32`

Pops the top value off the stack and returns it.

### `0x000C`: CM_STACK_JUMP_COND
`CM_STACK_JUMP_COND(options: u8, conditionType: u8, a: s32, offset: s32): void`

Works a lot like `CM_JUMP_COND`, except there is no way to invert the condition, and the `b` value is the top element of the stack. 

The `options` argument is a bitfield.
The lowest 3 bits determine how the value should be altered:
* 0x01: Increment after reading
* 0x02: Decrement after reading
* 0x03: Increment before reading
* 0x04: Decrement before reading

Bits 3 and 4 determine if the value should be popped from the stack, or just read:
* 0x02: Pop value when condition passes
* 0x03: Pop value when condition fails

Cases not listed in either bitfield do nothing.

### `0x000D`: CM_SETIMM_IDX
`CM_SETIMM_IDX(offset: s32, index: s32): s32`

Reads an immediate 32-bit value from the program code.
The value fetched is fetched using `ldr`, but since the program buffer is only guaranteed to be 16-bit aligned, watch out for misaligned reads.

The pointer to read data from is generated like so:
`manager.PC + offset + (index * 4) + 2`

### `0x000E`: CM_SETIMM
`CM_SETIMM_IDX(offset: s32): s32`

Reads an immediate 32-bit value from the program code.
The value fetched is fetched using `ldr`, but since the program buffer is only guaranteed to be 16-bit aligned, watch out for misaligned reads.

The pointer to read data from is generated like so:
`manager.PC + offset`

### `0x000F`: CM_SET
`CM_SET(value: i32): i32`

Returns the input value.
Might seem useless, except arguments don't work the way you'd expect in the VM.
This is the most direct way to move a value from one variable to another.

### `0x0010`: CM_ADD
`CM_ADD(a: i32, b: i32): i32`

returns `a + b`.

### `0x0011`: CM_SUB
`CM_SUB(a: i32, b: i32): i32`

returns `a - b`.

### `0x0012`: CM_MUL
`CM_MUL(a: i32, b: i32): i32`

returns `a * b`.

### `0x0013`: CM_DIV_SOFT
`CM_DIV_SOFT(a: s32, b: s32): s32`

Divides using software division.
returns `a / b`.

### `0x0014`: CM_MOD_SOFT
`CM_MOD_SOFT(a: s32, b: s32): s32`

Find the remainder using software division.
Returns `a % b`.

### `0x0015`: CM_LSL
`CM_LSL(a: i32, b: i8): i32`

Shifts `a` left by the value in `b`.
Returns `a << b`.

### `0x0016`: CM_MOD_SOFT
`CM_ASR(a: s32, b: i8): s32`

Shifts `a` right by the value in `b`.
Note that this is a signed shift, meaning the bits shifted in will have the same value as the old bit 31 of `a`.
Returns `a >> b`.

### `0x0017`: CM_AND
`CM_AND(a: i32, b: i32): i32`

returns `a & b`.

### `0x0018`: CM_ORR
`CM_ORR(a: i32, b: i32): i32`

returns `a | b`.

### `0x0019`: CM_EOR
`CM_EOR(a: i32, b: i32): i32`

returns `a ^ b`.

### `0x001A`: CM_NOT
`CM_NOT(a: bool): bool`

returns `a == 0`.

### `0x001B`: CM_NEG
`CM_NEG(a: i32): i32`

returns `a ^ -1`.

### `0x001C`: CM_SQRT_HARD
`CM_SQRT_HARD(a: u32): u32`

Calculates the square root using the ARM9 hardware square root math registers.
returns `sqrt(a)`.

### `0x001D`: CM_CURT_HARD
`CM_CURT_HARD(a: u32): u32`

Calculates the cubic root using the ARM9 hardware square root and division math registers.
Actually, I'm not totally sure of this one.
The documentation on [Yoshi's Lighthouse](https://www.tapatalk.com/groups/lighthouse_of_yoshi/scripting-t372.html) claims this MIGHT be cubic root, but I can't tell from just looking at the code.

returns `curt(a)`.

### `0x001E`: CM_DIV_HARD
`CM_DIV_HARD(a: s32, b: s32): s32`

Divides using hardware division.
returns `a / b`.

### `0x001F`: CM_COS
`CM_COS(a: s32): s16`

Returns the cosine of `s32`.
Not entirely sure what the range of values are...
Also not sure of this even calculates the cosine, or just uses the lookup table.

More research needed!

### `0x0020`: CM_SIN
`CM_SIN(a: s32): s16`

Returns the sine of `s32`.
Not entirely sure what the range of values are...
Also not sure of this even calculates the sine, or just uses the lookup table.

More research needed!

## Data types

### Jump conditions
```C
0x00: a == b
0x01: a != b
0x02: a < b
0x03: a > b
0x04: a <= b
0x05: a >= b
0x06: (a & b) != 0
0x07: (a | b) != 0
0x08: (a ^ b) != 0
0x09: a == 0
0x0a: a != -1 // 0xFFFFFFFF
```
