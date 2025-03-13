# PiT file documentation
Contains documentation for all (known) Partners in Time file formats.

## Table of contents
* BAI
    * BAI_iwasaki.dat
    * BAI_mon_0_hn.dat
    * BAI_mon_1_hn.dat
    * BAI_mon_2_hn.dat
    * BAI_min_3_hn.dat
    * BAI_min_4_hn.dat
    * BAI_mon_ji.dat
    * BAI_scn_0_hn.dat
    * BAI_scn_1_hn.dat
    * BAI_scn_2_hn.dat
    * BAI_scn_3_hn.dat
    * BAI_scn_4_hn.dat
    * BAI_scn_ji.dat
    * BAI_sugiyama.dat
    * BMes.dat
* BData
    * [BDataMon.dat](#bdatamondat)
    * [mfset_AItmC.dat](#mfset_aitmcdat)
    * [mfset_AItmE.dat](#mfset_aitmedat)
    * [mfset_AItmE2.dat](#mfset_aitme2dat)
    * [mfset_AItmN.dat](#mfset_aitmndat)
    * [mfset_BadgeE.dat](#mfset_badgeedat)
    * [mfset_BadgeN.dat](#mfset_badgendat)
    * [mfset_BonusE.dat](#mfset_bonusedat)
    * [mfset_Help.dat](#mfset_helpdat)
    * [mfset_MonN.dat](#mfset_monndat)
    * [mfset_UItmE.dat](#mfset_uitmedat)
    * [mfset_UItmE2.dat](#mfset_uitme2dat)
    * [mfset_UItmN.dat](#mfset_uitmndat)
    * [mfset_WearE.dat](#mfset_wearedat)
    * [mfset_WearN.dat](#mfset_wearndat)
* BFx
    * dfx_00.dat
    * rfx_00.dat
* BMap
    * BMap.dat
* BObjMon
    * BObjMon.dat
* BObjPc
    * BObjPc.dat
* BObjUI
    * BObjUI.dat
* Etc
    * Haraki
        * HarakiTestData.dat
    * Sasaki
        * TestMapData.dat
    * Sugiyama
        * Test.dat
    * Uchida
        * UchidaTestDat.dat
* FEvent
    * FEvData.dat
* FieldFx
    * FieldFxData.dat
* FMap
    * [FMapData.dat](#fmapdatadat)
* FObj
    * FObj.dat
* FObjMon
    * FObjMon.dat
* FObjPc
    * FObjPc.dat
* Font
    * StatFontSet.dat
* Menu
    * Menu.dat
    * [MenuDat.dat](#menudatdat)
    * [MenuDat](#menudat)
* MenuAI
    * BAI_iwasaki.dat
    * MAI_fujioka.dat
    * MAI_uchida.dat
    * [mfset_menu_mes.dat](#mfset_menu_mesdat)
    * [mfset_Mes_AreaName_out.dat](#mfset_mes_areaname_outdat)
    * [mfset_Mes_LoadSave_out.dat](#mfset_mes_loadsave_outdat)
    * mfset_Mes_MenuAI_out.dat
    * [mfset_Mes_Outline_out.dat](#mfset_mes_outline_outdat)
    * [mfset_option_mes.dat](#mfset_option_mesdat)
    * [mfset_shop_mes.dat](#mfset_shop_mesdat)
* SavePoint
    * [SavePhoto.dat](#savephotodat)
* Sound
    * sound_data.sdat
* Title
    * TitleBG.dat
* Treasure
    * [TreasureInfo.dat](#treasureinfodat)

## Lanugage order
Languages always come in this order:
* Japanese
* English
* French
* German
* Italian
* Spanish

## Pseudo-code
Throughout this document, I am going to be using a type of pseudo-code, to define structures:
```C
// Define a structure
struct structName {
    // Offset   type    name
    0x00:       u32,    number
    0x04:       u32,    number_2
}
```

## Common types
### dat
A `.dat` file is a file containing multiple entries.
The file begins with a table of offsets.
The offsets 
Every offset is a 32-bit unsigned integer.
Each offset must be a bigger number than the previous offset, and must not exceed the bounds of the file.
The end of the offset table has been reached, when an offset is equal to the total length of the `.dat` file.

The `.dat` file contains `n-1` entries, where N is the number of offsets, including the terminating offset (the one whose length is equal to the file size).
The data of each entry goes from its beginning offset, to the beginning of the next entry.

### Strings
Strings are null-terminated. 
Escape character is 0xFF.
Known escape sequences:
- 0x00: Newline
- 0x20: Color (default)
- 0x27: Color (green)
- 0x2D: Color (red)
- 0x35: Center string?

### MFset
I don't know what Mfset means, but I know what it contains.
A MFset is a list of an unknown number of [strings](#strings).

The beginning of the MFset structure contains a list of pointers (relative to the beginning of the MFset structure) to [strings](#strings).
From what I can tell, there is no way to tell how many entries an MFset structure contains, as there is no terminator in the array list (afaik).

## File documentation

### BDataMon.dat
Despite the filename, this is NOT a `.dat` file.
It contains the stats for all enemy types encountered in the game.

The file is an array of the following structure:
```C
struct enemy {
    0x00: u16, name ID in mfset_MonN.dat
    0x02-0x05: ???
    0x06: u16, HP
    0x08: u16, POW
    0x0A: u16, DEF
    0x0C: u16, SPD
    0x0E-0x1F: ???
    0x20: EXP from defeat
    0x22: Coins from defeat
    0x24-0x2B: ???
}
```

Enemies are listed in this order:
```
Baby Bowser (1)
Junior Shrooboid
Shroob
Shrooblet
...
```

### mfset_AItmC.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

All of the strings are just the word "DUMMY" followed by a newline...
I don't think this is used?

### mfset_AItmE.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains Bros. Attack descriptions.
Each string describes one page of a Bros. Attack demo description.
The strings are stored in a somewhat arbitrary order:
```
Cannonballers description 1
Green Shell description 1
Green Shell description 2
Dummy (unsued)
Red Shell description 1
Red Shell description 2
Bro Flower description 1
Bro Flower description 2
Smash Egg description 1
Smash Egg description 1
Mix Flower description 1
Mix Flower description 2
Trampoline description 1
Ice Flower description 1
Pocket Chomp description 1
Copy Flower description 1
Cannonballers description 2
Ice Flower description 2
Pocket Chomp description 2
Trampoline description 2
Solo Green Shell description
Solo Red Shell description
Solo Bro Flower description
Solo Smash Egg description
Solo Ice Flower description
Solo Pocket Chomp description
Copy Flower desctiption 2
```

### mfset_AItmE2.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

This file contains target/status modifiers for each [Bros. Attack](#bros-attacks-by-id).

### mfset_AItmN.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains name(s) of [Bros. Attacks](#bros-attacks-by-id).

The strings are grouped as follows for every entry:
```C
struct ItemNames {
    string, Singular         // Shown when player only has one
    string, NamePlural       // Shown when player has multiple
    string, NameStorageFull  // Shown when trying to obtain more at max capacity
}
```

It contains one of these structs for every Bros. attack.

#### Bros. Attacks by ID
```
Trampoline
Red Shell
Dummy (unused)
Green Shell
Bro Flower
Smash Egg
Mix Flower
Cannonballer
Ice Flower
Pocket Chomp
Copy Flower
```

#### Badges by ID
```
Nothing
Shroom Badge A
Shroom Badge
Coin Badge A
Coin Badge
EXP Badge A
EXP Badge
Treasure Badge
Big-POW Badge
Big-DEF Badge
Cure Badge A
Cure Badge
Drain Badge A
Drain Badge
Hit-POW Badge
Hit-Free Badge
Pep Badge
Dire-POW Badge
Dire-Free Badge
Stomp Badge
Pummel Badge
Counter Badge
Risk Badge
Training Badge
Easy Badge
Dynamic Badge
Dynamic Badge A
Rough Badge
Safety Badge
Salvage Badge A
Salvage Badge
Simple Badge
1-Change Badge
Item-Fan Badge
Wallet Badge
POW-Peak Badge
DEF-Peak Badge
Lucky Badge A
Lucky Badge
Ulti-Free Badge
Cash-Back Badge
```

### mfset_BadgeE.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains badge descriptions, one for every(?) badge.
I assume descriptions can be looked up using the same IDs from [mfset_BadgeN.dat](#mfset_badgendat), but I haven't double-checked.

### mfset_BadgeN.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).

Contains names of badges.
Structurally identical to [mfset_AItmN.dat](#mfset_aitmndat), just replace "Bros. Attack" with "Badge".

### mfset_BonusE.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Unsure what this file is for...
It contains the string '---\n' 22 times for each (non-japanese) language.
It MIGHT be the cost of Bros. Attacks when using the Ulti-Free Badge? But there are only 11 [Bros. Attack IDs](#bros-attacks-by-id), why would they need to be stored twice?

### mfset_Help.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains all the in-battle help.
Action-command names, instructions for how to select things, etc.

### mfset_MonN.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains a list of names of all enemies in the game.

#### Enemy names
```
Boo Guy
Elasto-Pihranha
Scoot Bloop
Koopeleon
Lakitufo
Spiny Shroopa
Swiggler
Hammer Bro
Kamek
Coconutter
Gnarantula
Dry Bones
Bully
Dr. Shroob
Junior Shrooboid
Goomba
Baby Bowser
Pihranha Planet
Handfake
Thwack
Snoozorb
Tanoomba
Mrs. Thwomp
Blazing Shroob
Thwack Totem
Bowser
Shroid
Love Bubble
Fly Guy
Skellokey
Elder Shrooboid
Boom Guy
Boo
Bob-omb
RC Shroober
Shrooba Diver
Pidgit
Pokey
Dark Boo
Snifaro
Shrooboid Brat
Petey Pihranha
Princess Shroob
Sunnycide
Shroob-omb
Commander Shroob
Shroob
Support Shroob
Shrooblet
Elite Boom Guy
Red Coconutter
Gold Koopeleon
Guardian Shroob
Soul Bubble
Wonder Thwack
Shroobsworth
Shroob Rex
Shrowser
Tashrooba
Ghoul Guy
Lethal Bob-omb
Crown
Tentacle
Foot
Intern Shroob
```


### mfset_UItmE.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains consumable item descriptions.
There is one for every(?) [Item](#item-ids).
I assume descriptions can be looked up using the same IDs from [mfset_UItmN.dat](#mfset_uitmndat), but I haven't double-checked.

### mfset_UItmE2.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains the string "DUMMY" for all languages except japanese.
Likely not used at all.

### mfset_UItmN.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains name(s) of consumable items.

Structurally identical to [mfset_AItmN.dat](#mfset_aitmndat), just replace "Bros. Attack" with "item".

#### Item IDs
```
Mushroom
Super Mushroom
1-Up Mushroom
Ultra Mushroom
Max Mushroom
1-Up Super
Refreshing Herb
Red Pepper
Green Pepper
Blue Pepper
Mushroom Drop
Super Drop
Ultra Drop
Bean
```

### mfset_WearE.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains wear descriptions.
There is one string for every(?) piece of [Wear](#wear-ids).
I assume descriptions can be looked up using the same IDs from [mfset_WearN.dat](#mfset_wearndat), but I haven't double-checked.

### mfset_WearN.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains names of wear items (slacks and pants).
Structurally identical to [mfset_AItmN.dat](#mfset_aitmndat), just replace "Bros. Attack" with "Wear".

#### Wear IDs
```
Wafer Slacks
Patched Slacks
Wild Trousers
Branded Slacks
Puffy Trousers
Shell Slacks
Adult Trousers
Muscle Slacks
Svelte Slacks
Block Trousers
Shroom Slacks
Star Trousers
Para Slacks
Royal Trousers
Space Trousers
Supreme Slacks
Silky Pants
Starchy Jeans
Unarmed Jeans
Preferred Pants
Heart Pants
Egg Pants
Secret Jeans
Thrilling Pants
Tissue Pants
Golden Pants
Mushroom Jeans
Stardust Pants
Stache Jeans
Royal Pants
Rocket Jeans
100-Point Pants
Nothing
```

### FMapData.dat
This is a [.dat file](#dat).
Contains overworld tilemaps.
Basically contains every room you explore in the game.

Every entry in this file looks to be RLZ compressed, except a few.
Their IDs are in a [table below](#non-rlz-compressed-files).

#### Not known
While we can render MOST maps with no problems, we are still far from figuring this all out.
Here is a list of things we haven't figured out yet:
* NPC data
* Map collision data
* Tilemap animations
* Tilemap layer transparency
  * Orange tint in Baby Bowser's Castle
  * Beams of light in Gritzy Desert Caves
  * Center mushroom before Hollijolli Village
  * Various light sources

#### FMapInfo lookup table
To decode the tilemaps of a FMap, you need 4 files:
* Layer 0 tileset
* Layer 1 tileset
* Layer 2 tileset
* Bundle file (tilemap, palettes, bounds, etc.)
* Treasure info

The files in `FMapData.dat` are stored in no particular order, but a lookup table of file pairings is embedded somewhere within the ARM9 binary (yes, really).
Where exactly it is located differs from region to region, and probably bewteen versions as well.
To find it, open the ARM9 binary in a hex editor, and search for the following sequence of bytes:
```
00 00 00 00   01 00 00 00   02 00 00 00   03 00 00 00
```
Those are the bytes at the start of the table.

The lookup itself has 638 entries, and each entry looks like this:
```C
struct FMapInfo {
    // IDs of tileset files.
    // Files are located in FMapData.dat.
    // A value of 0xFFFF_FFFF means layer is unused.
    0x00:   [3]u32, tileset_file_id

    // ID of bundle file.
    // File is located in FMapData.dat.
    // This value is never 0xFFFF_FFFF.
    0x0C:   u32,    bundle_file_id

    // ID for treasure file.
    // File is located in TreasureInfo.dat.
    // A value of 0xFFFF_FFFF means no treasure in this room.
    0x10:   u32,    treasure_file_id
}
```

The optional treasure file is documented [here](#treasureinfodat).

#### FMap bundle file
The bundle file is itself a RLZ compressed [.dat file](#dat).
Bundle files are made up of 13 (uncompressed) files:
* 0-2 are tilemaps, one for each layer (may be empty)
* 3-5 are palettes, one for each layer (may be empty)
* 6 contains metadata:
```C
struct FMapMetadata {
    // Width and height in tiles
    0x00:   u16,    map_width
    0x02:   u16,    map_height

    // unknown, but usually 0xFF
    0x04:   u8,     _

    // Lower 3 bits determine bit-depth
    // of each layers tileset
    // (0 = 4BPP, 1 = 8BPP)
    // The rest are unknown
    0x05:   u8,     layer_info  
    
    // Unknown (usually all 0)
    0x06:   [6]u8,  _
}
```
* 7-9 might be tilemap animation? (may be empty)
* 10-12 are unknown

#### non-RLZ compressed files
It is unknown exactly what is in these files.
```
87
119
186
417
422
693
853
860
863
865
869
874
938
947
980
985
993
1061
1110
1127
1192
1446
1456
1464
1466
1476
1502
1669
1735
1780
1813
1814
1821
1972
2851
```



### MenuDat.dat
This is a [.dat file](#dat).
It contains various graphics data for the games menu's.
There is no system to any of the files here.
I am just going to list what I have been able to make sense of.

Every entry in the file is LZ10 compressed, except for the palettes, which are uncompressed.
All tilesets are 8bpp, unless otherwise noted.

#### Full-screen tilemaps
| Tileset | Tilemap | Palette | Description |
|-|-|-|-|
| 0 | 1 | 2 | Suitcase, lid opening |
| 3 | 4 | 7 | Suitcase bottom-screen |
| 3 | 5 | 7 | Suitcase bottom-screen, w/o bros & cobalt star |
| 3 | 6 | 7 | Suitcase bottom-screen, w/o cobalt star |
| 8 | 9 | 12 | Suitcase bottom-screen, blurred |
| 8 | 10 | 12 | Suitcase bottom-screen, blurred, w/o bros & cobalt star |
| 8 | 11 | 12 | Suitcase bottom-screen, blurred, w/o cobalt star |
| 13 | 14 | 15 | Suitcase top-screen |
| 16 | 17 | 18 | Suitcase top-screen, grayed out baby slots mask? (4bpp) |
| 20 | 21 | 19 | Suitcase bottom-screen, items |
| 20 | 22 | 19 | Suitcase bottom-screen, key items |
| 23 | 24 | 19 | Suitcase bottom-screen, blurred, items |
| 25 | 26 | 19 | Suitcase bottom-screen, blurred, key items |
| 28 | 29 | 27 | Suitcase bottom-screen, gear w/o badges |
| 28 | 30 | 27 | Suitcase bottom-screen, gear |
| 31 | 32 | 27 | Suitcase bottom-screen, blurred, gear w/o badges |
| 31 | 32 | 27 | Suitcase bottom-screen, blurred, gear |
| 35 | 36 | 34 | Suitcase bottom-screen, badges |
| 37 | 38 | 34 | Suitcase bottom-screen, blurred, badges |
| 40 | 41 | 39 | Suitcase bottom-screen, stats w/o badges |
| 40 | 42 | 39 | Suitcase bottom-screen, stats |
| 43 | 44 | 39 | Suitcase bottom-screen, blurred, stats w/o badges |
| 43 | 45 | 39 | Suitcase bottom-screen, blurred, stats |
| 47 | 48 | 46 | Suitcase bottom-screen, bros. attacks |
| 49 | 50 | 46 | Suitcase bottom-screen, blurred, bros. attacks |
| 52 | 53 | 51 | Suitcase bottom-screen, cobalt star |
| 54 | 55 | 51 | Suitcase bottom-screen, cobalt-star (fade-in) |
| 72 | 73 | 74 | Suitcase top-screen, empty? (unused?) |
| 75 | 76 | 77 | Suitcase top-screen, all bros locked? (unused?) |
| 81 | 82 | 83 | Suitcase top-screen, empty, checkered at bottom? (unused?) |
| 84 | 86 | 85 | Save menu-looking popup |
| 84 | 87 | 85 | End of battle menu |
| 84 | 88 | 85 | End of battle, item rewards (?) |
| 84 | 89 | 85 | Shop, buy menu (coins) |
| 84 | 90 | 85 | Shop, category selector |
| 84 | 91 | 85 | Shop, buy menu (beans) |
| 92 | 93 | 94 | Peach's castle shop |
| 95 | 96 | 97 | Glitzy Caves shop |
| 98 | 99 | 100 | Toad Town shop |
| 101 | 102 | 103 | Fawful shop |
| 116 | 117 | 118 | Save menu, bottom-screen background |
| 119 | 120 | 121 | Save menu, initializing save data popup (4bpp) |
| 122 | 123 | 124 | Save menu, top-screen background |
| 136 | 137 | 138 | "Nintendo" logo (4bpp) |
| 139 | 140 | 138 | "Nintendo" logo again\*? (4bpp) |
| 142 | 143 | 144 | "Alphadream Corporation" logo (4bpp) |

\* Tilemap 137 and 140 are identical, but tileset 136 and 139 are not?
I cannot find a difference though.

#### Tile graphics without tilemap data
| Tilemap | Palette | Description |
|-|-|-|
| 56 | 57 | Suitcase opening, bottom screen (?) |
| 58 | 64 | Level-up menu text (japanese) |
| 59 | 64 | Level-up menu text (english) |
| 60 | 64 | Level-up menu text (french) |
| 61 | 64 | Level-up menu text (german) |
| 62 | 64 | Level-up menu text (italian) |
| 63 | 64 | Level-up menu text (spanish) |
| 65 | 66 | Don't know, looks like snow? (4bpp) |
| 67 | 64 | Equipment stats font (4bpp) |
| 68 | 64 | Equipment stats font (alternate) (4bpp) |
| 69 | 70 | Small (8x8) item icons |
| 104 | 106 | Equipment stats font (4bpp) (again?) |
| 105 | 106 | Equipment stats font (4bpp) (alternate) |
| 107 | 113 | Level-up stats (japanese) |
| 108 | 113 | Level-up stats (english) |
| 109 | 113 | Level-up stats (french) |
| 110 | 113 | Level-up stats (german) |
| 111 | 113 | Level-up stats (italian) |
| 112 | 113 | Level-up stats (spanish) |
| 114 | 115 | Battle HUD, character heads |

#### Bitmap graphics
| Bitmap | Palette | Description |
|-|-|-|
| 127 | 133 | Save-profile, character stats (4bpp) (japanese) |
| 128 | 133 | Save-profile, character stats (4bpp) (english) |
| 129 | 133 | Save-profile, character stats (4bpp) (french) |
| 130 | 133 | Save-profile, character stats (4bpp) (german) |
| 131 | 133 | Save-profile, character stats (4bpp) (italian) |
| 132 | 133 | Save-profile, character stats (4bpp) (spanish) |
| 134 | 135 | Level-up/stats font (4bpp) |
| 145 | 146 | Credits image 1 |
| 147 | 148 | Credits image 2 |
| 149 | 150 | Credits image 3 |
| 151 | 152 | Credits image 4 |
| 153 | 154 | Credits image 5 |
| 155 | 156 | Credits image 6 |
| 157 | 158 | Credits image 7 |
| 159 | 160 | Credits image 8 |
| 161 | 162 | Credits image 9 |
| 163 | 164 | Credits image 10 |
| 165 | 166 | Credits image 11 |
| 167 | 168 | Credits image 12 |
| 169 | 170 | Credits image 13 |
| 171 | 172 | Credits image 14 |
| 173 | 174 | Credits image 15 |
| 175 | 176 | Credits image 16 |
| 177 | ??? | Clouds, don't know where they are used (4bpp) |
| 179 | 180 | "The End" text (4bpp) (japanese) |
| 181 | 182 | "The End" text (4bpp) (english) |
| 183 | 184 | "The End" text (4bpp) (french) |
| 185 | 186 | "The End" text (4bpp) (german) |
| 187 | 188 | "The End" text (4bpp) (italian) |
| 189 | 190 | "The End" text (4bpp) (spanish) |

#### Unknown
I have no idea what these are used for:
* 71, palette
* 78, tileset
* 79, tileset
* 80, palette
* 125, tilemap (?)
* 126, palette
* 141, palette
* 178, palette
* 191, bitmap
* 192, palette
* 193, palette

### MenuDat
A small part of a C header file.
Contains a couple of `#define` statements, and nothing else.
How it ended up in the games filesystem is a mystery...

### mfset_menu_mes.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

This contains the names of the categories in the menu.
It also contains the names and descriptions of key items for some reason...

### mfset_Mes_AreaName_out.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains the names of every area the player can save in.
This is used on the save screen.

The file contains 46 entries, but only 24 are used.
The rest just say "TEMP".

### mfset_Mes_LoadSave_out.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains a majority of the text seen on the save screen.
Stuff like the "save & continue", "continue from Peach's Castle", and more.
They accounted for a lot more errors than I imagined with the savegames.

### mfset_Mes_Outline_out.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains the plot synopsis you see on the save screen.
Laid out in Cronological order (I think) from entry 0 to 64.

### mfset_option_mes.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains text relevant to the Rumble Pak option (singular).

### mfset_shop_mes.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains the text shown in all shops.
Also contains the dialouge for shopkeepers, spoken on the shopping menu.

### SavePhoto.dat
This is a [.dat file](#dat).

This file contains the images seen in the save album.
Each image is 256x192 pixels.
The same resolution as a DS screen.
There is a lot of empty space around each image.

There are 33 total images in the file, the last 5 being unused placeholders.
The entries in the file are grouped together in groups of 3:
* Tileset
* Tilemap
* Palette

#### Tileset
The tileset file is LZ10 compressed.
It contains 8x8 tiles, at 8BPP.

#### Tilemap
The tilemap file is LZ10 compressed.
It contains 32*24 = 768 16-bit integers, which is interpreted as follows:
* Bits 0-9 contain the tile ID (from the tileset data).
* Bit 10 toggles horizontal mirroring of the tile.
* Bit 11 toggles vertical mirroring of the tile.
* bits 12-15 are unused(?)

The tilemap goes from left-to-right, row by row.

#### Palette
The palette is not compressed.
The colors are stored in RGB555 format.
The first color of every palette is ignored, as it is treated as transparent.

### TreasureInfo.dat
This is a [.dat file](#dat).

Each file contains an array of "treasure" to be collected on a map.
Known types of treasure are beans and item blocks with coins, items and gear.
It is unknown how many types of treasure there are, and their exact data.

To determine which map a treasure file belongs to, see the [FMapInfoInfo lookup table](#fmapinfo-lookup-table).

Each treasure looks like this:
```C
struct TreasureInfo {
    0x00:   u8,     type
    0x01:   u8,     subtype
    0x02:   u16,    contents    // Depends on the type
    0x04:   u16,    id          // Unique treasure ID
    0x06:   u16,    pos_x
    0x08:   u16,    pos_y
    0x0A:   u16,    pos_z       // This is the up/down axis
}
```

Currently not known if there's a setting in the [FMap bundle](#fmap-bundle-file) determining how many treasures are on a map.
Maybe it just looks at the size of the treasure file and divides by `0x0A` to get the number of entries?

#### Treasure types
This list has a long way to go.
The current documentation process involves going into a hex editor, changing some values, reloading the ROM and observing the result.
Before I document more of this, I really need to develop a tool to re-compile maps.

* **type 0x0A:** Mario question block with items
* **type 0x0B:** Mario question block with coins
* **type 0x0C:** Luigi question block with items
* **type 0x0D:** Luigi question block with coins
* **type 0x0E:** question block with item
* **type 0x0F:** question block with coins. The content is always `0xFFFF`, all other values seem to crash the game. The top 3 bits of the subtype are ignored. The coin in the box is worth the value of the subtype field (after clearing the top 3 bits) plus one, with these exceptions:
  * **subtype 0x01:** single 10-coin
  * **subtype 0x02:** single 20-coin
