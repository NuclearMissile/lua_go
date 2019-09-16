package vm

func FloatByte2Int(i int) int {
	if i < 8 {
		return i
	} else {
		return ((i & 7) + 8) << uint((i>>3)-1)
	}
}

func Int2FloatByte(i int) int {
	e := 0
	if i < 8 {
		return i
	}
	for i >= (8 << 4) {
		i = (i + 0xF) >> 4
		e += 4
	}
	for i >= (8 << 1) {
		i = (i + 1) >> 1
		e++
	}
	return ((e + 1) << 3) | (i - 8)
}
