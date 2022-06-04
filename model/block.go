package model

type Block struct {
	offset uint32
	size uint32
	patchID uint16
	bitFlag uint16
	gcmTag [0x10]byte
}