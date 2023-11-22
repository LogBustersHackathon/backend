package utils

import "unsafe"

func ByteToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func StringToByte(s string) (b []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
