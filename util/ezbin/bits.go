package ezbin

func Bitget[OUT Integer, IN Integer](bitlist IN, bits int, pos int) OUT {
	mask := IN(1<<bits) - 1
	oval := (bitlist >> pos) & mask
	return OUT(oval)
}

func BitgetSigned[OUT Integer, IN Integer](bitlist IN, bits int, pos int) OUT {
	mask := IN(1<<bits) - 1
	oval := (bitlist >> pos) & mask
	if oval&(1<<bits-1) != 0 {
		oval |= ^mask
	}
	return OUT(oval)
}

func BitgetFlag[IN Integer](bitlist IN, pos int) bool {
	return bitlist&(1<<pos) != 0
}

func Bitset[IN Integer, VAL Integer](bitlist IN, val VAL, bits int, pos int) IN {
	mask := IN(1<<bits) - 1
	clean := bitlist & ^(mask << pos)
	return clean | ((IN(val) & mask) << pos)
}

func BitsetFlag[IN Integer](bitlist IN, flag bool, pos int) IN {
	clean := bitlist & ^(1 << pos)
	if flag {
		clean |= 1 << pos
	}
	return clean
}
