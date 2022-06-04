package model

type Header struct {
	PkgID uint16
	PatchID uint16
	EntryTableOffset uint32
	EntryTableCount uint32
}