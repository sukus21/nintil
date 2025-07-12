# PMOC

A compression scheme used by LEGO Battles, and possibly other games as well.

The format is more or less just a wrapper around the RLX format, but spit into segments.

In LEGO Battles, each segment is <= 0x4000 bytes long.
No clue if this is a limitation of the format, or that's just the size they decided on.

This implementation is lifted from [NitroPaint](https://github.com/Garhoogin/NitroPaint/blob/93a460b85f71fc46fa53f763f41ca7dd29c68699/NitroPaint/compression.c#L1557)
