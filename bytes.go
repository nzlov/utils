package utils

import "encoding/binary"

func Uint16ToBytes(in uint16) (out []byte) {
	out = make([]byte, 2)
	binary.BigEndian.PutUint16(out, in)
	return
}

func BytesToUint16(in []byte) uint16 {
	return binary.BigEndian.Uint16(in)
}

func BytesToUint16s(in []byte) (out []uint16) {
	for i := 0; i < len(in); i += 2 {
		out = append(out, BytesToUint16(in[i:i+2]))
	}
	return
}

func Uint32ToBytes(in uint32) (out []byte) {
	out = make([]byte, 4)
	binary.BigEndian.PutUint32(out, in)
	return
}

func BytesToUint32(in []byte) uint32 {
	return binary.BigEndian.Uint32(in)
}

func BytesToUint32s(in []byte) (out []uint32) {
	for i := 0; i < len(in); i += 4 {
		out = append(out, BytesToUint32(in[i:i+4]))
	}
	return
}
