

command info (r6 struct):
0x00:   u16,    command
0x02:   u16,    variable to write result to(?)
0x04:   u16,    flags bitfield
0x06:   ???,    ???
0x08:   u32,    value_1
0x0C:   u32,    value_2
0x10:   u32,    ???
...
0x4C:   null,   end


script manager (r7 struct):
0x00:   u32,    Script program counter

r8 struct:
0x00:   ???,    ???
0x0C:   u32,    command attribute table pointer
