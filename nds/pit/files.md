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
    * [mfset_AltmC.dat](#mfset_altmcdat)
    * [mfset_AltmE.dat](#mfset_altmedat)
    * [mfset_AltmE2.dat](#mfset_altme2dat)
    * [mfset_AltmN.dat](#mfset_altmndat)
    * [mfset_BadgeE.dat](#mfset_badgeedat)
    * [mfset_BadgeN.dat](#mfset_badgendat)
    * [mfset_BonusE.dat](#mfset_bonusedat)
    * [mfset_Help.dat](#mfset_helpdat)
    * [mfset_MonN.dat](#mfset_monndat)
    * [mfset_UltmE.dat](#mfset_ultmedat)
    * [mfset_UltmE2.dat](#mfset_ultme2dat)
    * [mfset_UltmN.dat](#mfset_ultmndat)
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
    * FMapData.dat
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
    * MenuDat.dat
    * MenuDat
* MenuAI
    * BAI_iwasaki.dat
    * MAI_fujioka.dat
    * MAI_uchida.dat
    * mfset_menu_mes.dat
    * mfset_Mes_AreaName_out.dat
    * mfset_Mes_LoadSave_out.dat
    * mfset_Mes_MenuAI_out.dat
    * mfset_Mes_Outline_out.dat
    * mfset_option_mes.dat
    * mfset_shop_mes.dat
* SavePoint
    * SavePhoto.dat
* Sound
    * sound_data.sdat
* Title
    * TitleBG.dat
* Treasure
    * TreasureInfo.dat

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
 - 0x00: Split string?
 - 0x0A: Newline
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

### mfset_AltmC.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

All of the strings are just the word "DUMMY" followed by a newline...
I don't think this is used?

### mfset_AltmE.dat
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

### mfset_AltmE2.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

This file contains target/status modifiers for each [Bros. Attack](#bros-attacks-by-id).

### mfset_AltmN.dat
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
Structurally identical to [mfset_AltmN.dat](#mfset_altmndat), just replace "Bros. Attack" with "Badge".

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


### mfset_UltmE.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains consumable item descriptions.
There is one for every(?) [Item](#item-ids).
I assume descriptions can be looked up using the same IDs from [mfset_UltmN.dat](#mfset_ultmndat), but I haven't double-checked.

### mfset_UltmE2.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains the string "DUMMY" for all languages except japanese.
Likely not used at all.

### mfset_UltmN.dat
This is a [.dat file](#dat).
Each entry corrosponds to a [language](#lanugage-order).
Inside every entry is a [MFset](#mfset) structure.

Contains name(s) of consumable items.

Structurally identical to [mfset_AltmN.dat](#mfset_altmndat), just replace "Bros. Attack" with "item".

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
Structurally identical to [mfset_AltmN.dat](#mfset_altmndat), just replace "Bros. Attack" with "Wear".

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
