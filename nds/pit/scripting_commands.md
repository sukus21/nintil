# AlphaDream VM documentation

This is documentation for the specific VM used in Partners in Time.
The VMs used in the rest of the M&L series are similar, but not identical.

Huge thanks to the people at [BIS-docs](https://mnl-modding.github.io/BIS-docs/) for helping me figure this stuff out ❤️

# Command parsing
Before parsing even begins, the VM has a table with command opcode attributes.
It is not yet know exactly what these attributes look like, but they play a critical role in command parsing.
The table is constant and never changes.

Note in the following pseudo-code, `vm.read()` consumes 16 bits of data from the program.
See [variable addressing modes](#variable-addressing-modes) for information on `vm.readVariables()`.
[Command structs](#command-struct) are roughly read like this:
```go
// Read command metadata
command.opcode = vm.read()
attributes := vm.getAttributes(opcode)
if attributes.hasResult {
    command.result = vm.read()
}
if attributes.hasFlags {
    command.flags = cm.read()
}

// Read arguments
for i := 0; i < attributes.numArguments; i++ {
    if command.flags & (1 << i) == 0 {
        command.arguments[i] = vm.read()
    } else {
        command.arguments[i] = vm.readVariable(vm.read());
    }
}
```

## Variable addressing modes
When reading a VM variable, the top 4 bits of the variable are not used for the address, but for addressing mode.
Here is a list of known addressing modes:

### `0x0000`: 0
This addressing mode always returns 0.

### `0x1000`: value from manager
Reads a value from the script manager.
Maybe from the stack? not sure...
`return (u32*)(scriptManager)[1 + variable];`

### `0x2000`: boolean flag
Reads a single bit from a bitlist.
Almost identical to `0xE000` + `0xF000`, but only has 12 bits for bit index.
The bitlist read from is also not the same.

### `0x3000`: field specific
Calls to a function that is only valid when overlay 0 (field program) is loaded.
If jumped to in battle, will probably crash the game completely.
Unknown exactly what happens and what is returned.

### `0x4000`: battle specific
Same as `0x3000`, but for battles.
Unknown what is returned from this method (more research needed).

### `0x5000` through `0xD000`: TBD
**More research needed!**

## `0xE000` and `0xF000`: boolean flag
Reads a single bit from a bitlist.
The purpose of this bitlist is currently unknown.
The value in the lower 13 (yes 13, not 12. That's why both E and F) is the bit index.

## Data types

### Command struct
A command in the PiT VM looks like this:
```C
struct VmCommand {
    0x00:   u16,        opcode      // The command opcode
    0x02:   u16,        result      // The variable to store the result at
    0x04:   u16,        flags       // Determines if an argument is an immediate or not
    0x06:   u16,        _           // unknown
    0x08:   [16]i32,    arguments   // Command arguments
    0x48:   u16,        _           // unknown
}
```

# Command opcodes

All instructions return a status of `0` to the interpreter, unless otherwise noted.

## Pseudocode notes
data types:
* sXX = signed XX-bit int
* uXX = unsigned XX-bit int
* iXX = XX-bit int, can be either signed or unsigned
* fx32 = signed 32-bit fixed point number with 12-bit fractional part

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

### `0x000D`: CM_MOV_OFFSET_IDX
`CM_MOV_OFFSET_IDX(offset: s32, index: s32): s32`

Reads an immediate 32-bit value from the program code.
The value fetched is fetched using `ldr`, but since the program buffer is only guaranteed to be 16-bit aligned, watch out for misaligned reads.

The pointer to read data from is generated like so:
`manager.PC + offset + (index * 2) + 2`

### `0x000E`: CM_MOV_OFFSET
`CM_MOV_OFFSET(offset: s32): s32`

Reads an immediate 32-bit value from the program code.
The value fetched is fetched using `ldr`, but since the program buffer is only guaranteed to be 16-bit aligned, watch out for misaligned reads.

The pointer to read data from is generated like so:
`manager.PC + offset`

### `0x000F`: CM_MOV_INT
`CM_MOV_INT(value: i32): i32`

Returns the input value.
Might seem useless, except arguments don't work the way you'd expect in the VM.
This is the most direct way to move a value from one variable to another.

### `0x0010`: CM_ADD_INT
`CM_ADD_INT(a: i32, b: i32): i32`

returns `a + b`.

### `0x0011`: CM_SUB_INT
`CM_SUB_INT(a: i32, b: i32): i32`

returns `a - b`.

### `0x0012`: CM_MUL_INT
`CM_MUL_INT(a: i32, b: i32): i32`

returns `a * b`.

### `0x0013`: CM_DIV_INT
`CM_DIV_INT(a: s32, b: s32): s32`

Divides using software division.
returns `a / b`.

### `0x0014`: CM_MOD_INT
`CM_MOD_INT(a: s32, b: s32): s32`

Find the remainder using software division.
Returns `a % b`.

### `0x0015`: CM_LSL
`CM_LSL(a: i32, b: i8): i32`

Shifts `a` left by the value in `b`.
Returns `a << b`.

### `0x0016`: CM_ASR
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

### `0x001C`: CM_SQRT_INT
`CM_SQRT_INT(a: u32): u32`

Calculates the square root using the ARM9 hardware square root math registers.
returns `sqrt(a)`.

### `0x001D`: CM_INVSQRT_INT
`CM_INVSQRT_INT(a: s32): i32`

Calculates the inverse square root using the ARM9 hardware math registers.

**More research needed!**
Not exactly sure what the purpose of this is.
The inverse square root of any integer will always be 1, or less than 1.
In cases less than one, it is rounded down to an integer, and the result is 0...
So the result of this command is 1 if the input is 1, otherwise 0...?

returns `1 / sqrt/(a)`.

### `0x001E`: CM_INVERT_INT
`CM_INVERT_INT(a: s32): s32`

Calculates the inverse of a number using ARM9 hardware division math registers.
returns `1 / a`.

### `0x001F`: CM_COS_INT
`CM_COS_INT(a: s32): s16`

Returns the cosine of `s32`.
The value is fetched from a lookup table in the ARM9 binary, which contains `0x1000` sine and cosine pairs.
Not exactly sure what the range of values are, more research needed.

Returns `cos(a * ?)`.

### `0x0020`: CM_SIN_INT
`CM_SIN_INT(a: s32): s16`

Returns the sine of `s32`.
The value is fetched from a lookup table in the ARM9 binary, which contains `0x1000` sine and cosine pairs.
Not exactly sure what the range of values are, more research needed.

Returns `sin(a * ?)`.

### `0x0021`: CM_ATAN_INT
`CM_ATAN_INT(a: s32): s32`

### `0x0022`: CM_ATAN2_INT
`CM_ATAN2_INT(y: s32, x: s32): s32`

### `0x0023`: CM_RANDOM
`CM_RANDOM(max: u16): u16`

Returns a random value: `0 <= X < max`.

### `0x0024`: CM_MOV_FIXED
`CM_MOV_FIXED(value: fx32): fx32`

Same as `CM_MOV_INT`, but takes [fixed point parameters](#fixed-point-parameter).

### `0x0025`: CM_TRUNC_FIXED
`CM_TRUNC_FIXED(value: fx32): s32`

Uses [fixed point parameters](#fixed-point-parameter).
Removes the fraction part and returns an integer.

### `0x0026`: CM_FLOOR_FIXED
`CM_FLOOR_FIXED(value: fx32): fx32`

Uses [fixed point parameters](#fixed-point-parameter).
Clears the fraction bits.

### `0x0027`: CM_ADD_FIXED
`CM_ADD_FIXED(a: fx32, b: fx32): fx32`

Same as `CM_ADD_INT`, but uses [fixed point parameters](#fixed-point-parameter).

### `0x0028`: CM_SUB_FIXED
`CM_SUB_FIXED(a: fx32, b: fx32): fx32`

Same as `CM_SUB_INT`, but uses [fixed point parameters](#fixed-point-parameter).

### `0x0029`: CM_MUL_FIXED
`CM_MUL_FIXED(a: fx32, b: fx32): fx32`

Same as `CM_MUL_INT`, but uses [fixed point parameters](#fixed-point-parameter).

### `0x002A`: CM_DIV_FIXED
`CM_DIV_FIXED(a: fx32, b: fx32): fx32`

Divides using hardware division.
Uses [fixed point parameters](#fixed-point-parameter).
returns `a / b`.

### `0x002B`: CM_MOD_FIXED
`CM_MOD_FIXED(a: fx32, b: fx32): fx32`

Find the remainder using hardware division.
Uses [fixed point parameters](#fixed-point-parameter).
Returns `a % b`.

### `0x002C`: CM_SQRT_FIXED
`CM_SQRT_FIXED(a: fx32): fx32`

Calculates the square root using the ARM9 hardware square root math registers.
Uses [fixed point parameters](#fixed-point-parameter).
returns `sqrt(a)`.

### `0x002D`: CM_INVSQRT_FIXED
`CM_INVSQRT_FIXED(a: fx32): fx32`

Calculates the inverse square root using the ARM9 hardware math registers.
Uses [fixed point parameters](#fixed-point-parameter).
returns `1.0 / sqrt/(a)`.

### `0x002E`: CM_INVERT_FIXED
`CM_INVERT_FIXED(a: fx32): fx32`

Calculates the inverse of a number using ARM9 hardware division math registers.
Uses [fixed point parameters](#fixed-point-parameter).
returns `1.0 / a`.

### `0x002F`: CM_COS_FIXED
`CM_COS_FIXED(a: fx32): fx32`

Uses [fixed point parameters](#fixed-point-parameter).
The value is fetched from a lookup table in the ARM9 binary, which contains `0x1000` sine and cosine pairs.

Returns `cos(a)`.

### `0x0030`: CM_SIN_FIXED
`CM_SIN_FIXED(a: fx32): fx32`

Uses [fixed point parameters](#fixed-point-parameter).
The value is fetched from a lookup table in the ARM9 binary, which contains `0x1000` sine and cosine pairs.

Returns `sin(a)`.

### `0x0031`: CM_ATAN_FIXED
`CM_ATAN_FIXED(a: fx32): fx32`

Same as `CM_ATAN_INT`, but uses [fixed point parameters](#fixed-point-parameter).

### `0x0032`: CM_ATAN2_INT
`CM_ATAN2_INT(y: fx32, x: fx32): s32`

Same as `CM_ATAN2_INT`, but uses [fixed point parameters](#fixed-point-parameter).



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

### Fixed point parameter
If a command takes `fx32` as a parameter, and the parameter is an immediate, two arguments are consumed to form a single `f32`.

For example, `CM_ADD_FIXED` takes 2 `fx32` arguments.
```go
var fx32Args [2]fx32;

// Get fx32 argument 0
fx32Args[0] = command.arguments[0]
if command.isImmediate(0) {
    fx32Args[0] = command.arguments[0] | (command.arguments[1] << 16)
}

// Get fx32 argument 1
fx32Args[1] = command.arguments[2]
if command.isImmediate(2) {
    fx32Args[1] = command.arguments[2] | (command.arguments[3] << 16)
}
```
