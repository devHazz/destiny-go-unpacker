package model

type Entries struct {
	a uint32
	b uint32
	c uint32
	d uint32
}

type Entry struct {
	ref uint
	numType uint
	numSubType uint
	startingBlock uint
	startingBlockOffset uint
	fileSize uint
}