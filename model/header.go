package model

type header struct {
	PkgID uint16
	PatchID uint16
	EntryTableOffset uint32
	EntryTableCount uint32
}

func headerStruct() (ret header) {
	return header{}
}