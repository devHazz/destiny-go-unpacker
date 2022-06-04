package main

import (
    "fmt"
	"os"
	//"bytes"
	"encoding/binary"
)

type header struct {
	PkgID uint16
	PatchID uint16
	EntryTableOffset uint32
	EntryTableCount uint32
}

type entry struct {
	ref uint
	entryB uint 
	entryC uint 
	entryD uint 
}

func getHeader(path string) (ret header) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Print(err)
	}
	data := make([]byte, 0x16F)
	header_length, readError := file.Read(data)
	if readError != nil {
		fmt.Println(readError)
	}
	var h header
	header_data := data[:header_length]
	h.PkgID = binary.LittleEndian.Uint16(header_data[0x10:])
	h.PatchID = binary.LittleEndian.Uint16(header_data[0x30:])
	h.EntryTableOffset = binary.LittleEndian.Uint32(header_data[0x44:])
	h.EntryTableCount = binary.LittleEndian.Uint32(header_data[0x60:])
	return h
}

func main() {
    fmt.Println("Destiny 2 Go Unpacker, Written by Hazz")
	g := getHeader("./test/example.pkg") //Set Package Path Here
	fmt.Printf("Package ID: %d\nPatch ID: %d\nEntry Table Offset: %d\nEntry Table Count: %d", g.PkgID, g.PatchID, g.EntryTableOffset, g.EntryTableCount)
}