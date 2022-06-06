package main

import (
    "fmt"
	"os"
	//"bytes"
	"encoding/binary"
	"model"
	"io"
	"github.com/new-world-tools/go-oodle"
)

//var header_data model.Header
var header_data model.Header
func getHeader() (ret model.Header) {
	file, err := os.Open("./test/example.pkg")
	if err != nil {
		fmt.Print(err)
	}
	stat, _ := file.Stat()
	data := make([]byte, stat.Size())
	header_length, readError := file.Read(data)
	if readError != nil {
		fmt.Println(readError)
	}
	var h model.Header
	file_data := data[:header_length]
	h.RawFile = *file
	h.HeaderBin = file_data
	h.FileSize = stat.Size()


	h.PkgID = binary.LittleEndian.Uint16(file_data[0x10:])
	h.PatchID = binary.LittleEndian.Uint16(file_data[0x30:])
	h.EntryTableOffset = binary.LittleEndian.Uint32(file_data[0x44:])
	h.EntryTableCount = binary.LittleEndian.Uint32(file_data[0x60:])
	h.BlockTableSize = binary.LittleEndian.Uint32(file_data[0x68:])
	h.BlockTableOffset = binary.LittleEndian.Uint32(file_data[0x6C:])
	header_data = h
	defer file.Close()

	return h
}

func getEntries() (ret model.Entries) {
	header := getHeader()
	var entries model.Entries;
	var decoded_entries []model.Entry
	entry_bin := header.HeaderBin[header.EntryTableOffset:header.EntryTableCount]
	for i:= 0; i < int(header.EntryTableCount) * 16; i += 16 {
		//data := make([]byte, header.FileSize)
		//header.RawFile.Seek(int64(i), 0)

		entries.A = binary.LittleEndian.Uint32(entry_bin[i:])
		entries.B = binary.LittleEndian.Uint32(entry_bin[i + 4:])
		entries.C = binary.LittleEndian.Uint32(entry_bin[i + 8:])
		entries.D = binary.LittleEndian.Uint32(entry_bin[i + 12:])
		entry := readEntry(entries)
		decoded_entries = append(decoded_entries, entry)
	}
	return decoded_entries
}

func getBlocks() (ret []model.Block) {
	header := getHeader()
	var block model.Block
	var blocks_slice []model.Block
	file, err := os.Open("./test/example.pkg")
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()
	block_table := header.BlockTableOffset + header.BlockTableSize * 48
	for i:= header.BlockTableOffset; i < block_table; i += 48 {

		//binary.Read(i, binary.LittleEndian, &blockDataStruct)
		file.Seek(int64(i), io.SeekStart)
		block.Offset = model.ReadUint32(file)
		block.Size = model.ReadUint32(file)
		block.PatchID = model.ReadUint16(file)
		block.BitFlag = model.ReadUint16(file)

		file.Seek(0x20, io.SeekCurrent)
		block.GcmTag = model.ReadGCMBuffer(file)
		/*
		gcm := make([]byte, 0x10)
		binary.LittleEndian.PutUint16(gcm,  binary.LittleEndian.Uint16(header.HeaderBin[i + 0x20:]))
		*/
		blocks_slice = append(blocks_slice, block)
	}
	return blocks_slice
}

func readEntry(raw_entry model.Entries) (ret []model.Entry) {
	header := getHeader()
	entries := getEntries()

	var entry model.Entry
	entry.Ref = entries.A
	entry.NumType = (entries.B >> 9) & 0x7F
	entry.NumSubType = (entries.B >> 6) & 0x7
	entry.StartingBlock = entries.C & 0x3FFF
	entry.StartingBlockOffset = ((entryC >> 14) & 0x3FFF) << 4
	entry.FileSize = (entryD & 0x3FFFFFF) << 4 | (entryC >> 28) & 0xF
	return entry
}

func oodleDecompressBlock(block model.Block) (ret []byte) {
	blocks := getBlocks()
	var decompressed_blocks []model.Block
	for i := 0; i < len(blocks); i += 1 {
		decompress := oodle.Decompress()
	}
}

func main() {
	//g := getHeader() //Set Package Path Here
	fmt.Print(getBlocks())
	//fmt.Printf("Package ID: %d\nPatch ID: %d\nEntry Table Offset: %d\nEntry Table Count: %d", g.PkgID, g.PatchID, g.EntryTableOffset, g.EntryTableCount)
}