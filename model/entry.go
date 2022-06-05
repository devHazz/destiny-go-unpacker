package model

type Entries struct {
	A uint32
	B uint32
	C uint32
	D uint32
}

type Entry struct {
	Ref uint
	NumType uint
	NumSubType uint
	StartingBlock uint
	StartingBlockOffset uint
	FileSize uint
}