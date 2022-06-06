package model

type Entries struct {
	A uint32
	B uint32
	C uint32
	D uint32
}

type Entry struct {
	Ref uint32
	NumType uint32
	NumSubType uint32
	StartingBlock uint32
	StartingBlockOffset uint32
	FileSize uint32
}