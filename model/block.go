package model

type Block struct {
	Offset uint32
	Size uint32
	PatchID uint16
	BitFlag uint16
	GcmTag []byte
}