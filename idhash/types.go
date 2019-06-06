package idhash

var types = [...]string{"invalid", "event"}

func typeToInt(typ string) int64 {
	for index, str := range types {
		if str == typ {
			return int64(index)
		}
	}
	return 0
}

func intToType(in int64) string {
	if int(in) >= len(types) {
		return ""
	}
	return types[in]
}
