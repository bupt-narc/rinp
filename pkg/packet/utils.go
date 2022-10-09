package packet

func uint16ToByteSlice(in uint16) []byte {
	return []byte{byte(in), byte(in >> 8)}
}

func byteSliceToUint16(in []byte) uint16 {
	return uint16(in[0]) + uint16(in[1])<<8
}

func appendSlices[T any](in ...[]T) []T {
	var ret []T
	for _, i := range in {
		ret = append(ret, i...)
	}
	return ret
}
