package main

import (
	"crypto/aes"
	"fmt"
	"os"
	"crypto/cipher"
	"encoding/binary"
	"io"
	"model"
	"github.com/new-world-tools/go-oodle"
)

//var header_data model.Header
var header_data model.Header
var nonce [12]byte


func getHeader() (ret model.Header) {
	file, err := os.Open(model.Path)
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

func getEntries() (ret []model.Entry) {
	header := getHeader()
	file, err := os.Open(model.Path)
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()
	var entries model.Entries;
	var decoded_entries []model.Entry
	//entry_bin := header.HeaderBin[header.EntryTableOffset:header.EntryTableCount]
	for i:= 0; i < int(header.EntryTableCount) * 16; i += 16 {
		//data := make([]byte, header.FileSize)
		//header.RawFile.Seek(int64(i), 0)
		file.Seek(int64(i), io.SeekStart)
		entries.A = model.ReadUint32(file)
		entries.B = model.ReadUint32(file)
		entries.C = model.ReadUint32(file)
		entries.D = model.ReadUint32(file)

		entry := readEntry(entries)
		decoded_entries = append(decoded_entries, entry)
	}
	return decoded_entries
}

func getBlocks() (ret []model.Block) {
	header := getHeader()
	var block model.Block
	var blocks_slice []model.Block
	file, err := os.Open(model.Path)
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

func readEntry(raw_entry model.Entries) (ret model.Entry) {
	var entry model.Entry
	entry.Ref = raw_entry.A
	entry.NumType = (raw_entry.B >> 9) & 0x7F
	entry.NumSubType = (raw_entry.B >> 6) & 0x7
	entry.StartingBlock = raw_entry.C & 0x3FFF
	entry.StartingBlockOffset = ((raw_entry.C >> 14) & 0x3FFF) << 4
	entry.FileSize = (raw_entry.D & 0x3FFFFFF) << 4 | (raw_entry.C >> 28) & 0xF
	return entry
}

func oodleDecompressBlock(block model.Block, block_bin []byte) (ret []byte) {
		decompress, err := oodle.Decompress(block_bin, 0x40000)
		if err != nil { fmt.Println(err) }
		return decompress
}

func changeNonce() (ret [0xC]byte) {
	nonce = model.Nonce
	nonce[0] ^= byte((header_data.PkgID >> 8))
	nonce[11] ^= byte(header_data.PkgID)
	return nonce
}

func decrypt(block model.Block, block_bin []byte) (ret []byte) {
	var key []byte
	if block.BitFlag & 4 != 0 {
		key = model.AES_2
	} else {
		key = model.AES_1
	}
cipherBlock, err := aes.NewCipher([]byte(key))
if err != nil {
    fmt.Println(err)
}

aesgcm, err := cipher.NewGCM(cipherBlock)
if err != nil {
    fmt.Println(err)
}
nonce := changeNonce()
pt, err := aesgcm.Open(nil, nonce[:], block_bin, nil)
if err != nil {
	fmt.Println(err)
}
return pt
}

func extract() {
	entries := getEntries()
	blocks := getBlocks()
	for i := 0; i < len(entries); i++ {
		current_block := header_data.HeaderBin[blocks[i].Offset:blocks[i].Offset + blocks[i].Size]
		if blocks[i].BitFlag & 0x2 == 1 {
			//AES Decryption needed
			current_block = decrypt(blocks[i], current_block)
		}

		if blocks[i].BitFlag & 0x1 == 1 {
			//Oodle decompression needed
		}
		fmt.Println(current_block)
	}
}

func main() {
	model.Path = "./test/example.pkg"
	extract()
	//fmt.Printf("Package ID: %d\nPatch ID: %d\nEntry Table Offset: %d\nEntry Table Count: %d", g.PkgID, g.PatchID, g.EntryTableOffset, g.EntryTableCount)
}