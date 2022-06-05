package model 

import (
 "io"
 "encoding/binary"
)

func ReadGCMBuffer(r io.Reader) ([]byte) {
	buf := make([]byte, 16)
	io.ReadFull(r, buf)
	return buf
}

func ReadUint16(r io.Reader) (uint16) {
	buf := make([]byte, 2)
	io.ReadFull(r, buf)
	return binary.LittleEndian.Uint16(buf)
}

func ReadUint32(r io.Reader) (uint32) {
	buf := make([]byte, 4)
	io.ReadFull(r, buf)
	return binary.LittleEndian.Uint32(buf)
}