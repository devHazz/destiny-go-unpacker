package model
import (
	"os"
)
type Header struct {
	RawFile os.File
	HeaderBin []byte
	FileSize int64
	PkgID uint16
	PatchID uint16
	EntryTableOffset uint32
	EntryTableCount uint32
	BlockTableSize uint32
	BlockTableOffset uint32
	Hash64TableSize uint32
	Hash64TableOffset uint32
}