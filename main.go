package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"github.com/new-world-tools/go-oodle"
	"io"
	"io/ioutil"
	"log"
	"math"
	"model"
	"os"
	"strings"
)

//var header_data model.Header
var header_data model.Header
var nonce []byte

func latestPatchId(id string) (ret string) {

	pkgPath := model.Path
	var patchId string
	largestId := "-1"
	packages, _ := ioutil.ReadDir(pkgPath)
	for _, pkg := range packages {
		if !pkg.IsDir() {
			if strings.Contains(pkg.Name(), id) {
				patchId = pkg.Name()[len(pkg.Name())-5:]
				if patchId > largestId {
					largestId = patchId
				}
				model.Name = pkg.Name()[:len(pkg.Name())-6]
			}
		}
	}
	return model.Path + "/" + model.Name + "_" + patchId
}

func getHeader() (ret model.Header) {
	file, err := os.Open(model.PackagePath)
	if err != nil {
		fmt.Print(err)
	}
	stat, _ := file.Stat()
	data := make([]byte, stat.Size())
	header_length, readError := file.Read(data)
	if readError != nil {
		log.Fatal(readError)
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
	file, err := os.Open(model.PackagePath)
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()
	var entries model.Entries
	var decoded_entries []model.Entry
	//entry_bin := header.HeaderBin[header.EntryTableOffset:header.EntryTableCount]
	for i := header.EntryTableOffset; i < header.EntryTableOffset+header.EntryTableCount*16; i += 16 {
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
	var blocksSlice []model.Block
	file, err := os.Open(model.PackagePath)
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()
	for i := header.BlockTableOffset; i < header.BlockTableOffset+header.BlockTableSize*48; i += 48 {

		//binary.Read(i, binary.LittleEndian, &blockDataStruct)
		file.Seek(int64(i), io.SeekStart)
		block.Offset = model.ReadUint32(file)
		block.Size = model.ReadUint32(file)
		block.PatchID = model.ReadUint16(file)
		block.BitFlag = model.ReadUint16(file)

		file.Seek(int64(i+0x20), io.SeekStart)
		block.GcmTag = model.ReadGCMBuffer(file)
		blocksSlice = append(blocksSlice, block)
	}
	return blocksSlice
}

func readEntry(rawEntry model.Entries) (ret model.Entry) {
	var entry model.Entry
	entry.Ref = rawEntry.A
	entry.NumType = (rawEntry.B >> 9) & 0x7F
	entry.NumSubType = (rawEntry.B >> 6) & 0x7
	entry.StartingBlock = rawEntry.C & 0x3FFF
	entry.StartingBlockOffset = ((rawEntry.C >> 14) & 0x3FFF) << 4
	entry.FileSize = (rawEntry.D&0x3FFFFFF)<<4 | (rawEntry.C>>28)&0xF

	//hex := fmt.Sprintf("0x%x", entry.NumType)
	//fmt.Println(hex)
	return entry
}

func oodleDecompressBlock(block model.Block, block_bin []byte) (ret []byte) {
	if !oodle.IsDllExist() {
		log.Fatal("DLL not here mate")
	}
	decompress, err := oodle.Decompress(block_bin, 0x40000)
	if err != nil {
		log.Fatal(err)
	}
	return decompress
}

func changeNonce() {
	model.Nonce[0] ^= byte(header_data.PkgID >> 8 & 0xFF)
	model.Nonce[11] ^= byte(header_data.PkgID & 0xFF)
}

func decrypt(block model.Block, blockBin []byte) (ret []byte) {
	var key []byte
	if block.BitFlag&0x4 != 0 {
		key = model.AES_2
	} else {
		key = model.AES_1
	}
	blockBin = append(model.Nonce, blockBin...)
	blockBin = append(blockBin, block.GcmTag...)

	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		log.Panic("Cipher error")
	}

	aesgcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		log.Panic("New GCM Error")
	}
	//fmt.Println(aesgcm.NonceSize())
	//fmt.Printf("tag Size: {%d}\n", aesgcm.Overhead())
	//gcmNonce := changeNonce()
	pr, err := aesgcm.Open(nil, blockBin[:aesgcm.NonceSize()], blockBin[12:], nil)
	if err != nil {
		log.Panic(err)
	}
	blockBin = pr
	return blockBin
}

func extract() {
	entries := getEntries()
	blocks := getBlocks()
	file, err := os.Open(model.PackagePath)
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()
	for i := 0; i <= int(header_data.PatchID); i++ {
		patchPath := model.PackagePath
		patchPath = patchPath[:len(patchPath)-5] + string(rune(i+48))
		model.PatchStreamPath = append(model.PatchStreamPath, patchPath)
		//fmt.Println(model.PatchStreamPath)
	}
	changeNonce()
	for i := 0; i < len(entries); i++ {
		currentEntry := entries[i]
		blockIndex := currentEntry.StartingBlock
		blockStartingIndex := currentEntry.StartingBlockOffset
		blockCount := math.Floor((float64(blockStartingIndex) + float64(currentEntry.FileSize) - 1) / 262144)
		fileBuffer := make([]byte, currentEntry.FileSize)
		if currentEntry.FileSize == 0 {
			blockCount = 0
		}
		lastBlockIndex := blockIndex + uint32(blockCount)
		var curBufferOffset int = 0
		for blockIndex <= lastBlockIndex {
			currentBlock := blocks[blockIndex]
			file, err := os.Open(model.PatchStreamPath[currentBlock.PatchID] + ".pkg")
			if err != nil {
				log.Fatal(err)
			}
			file.Seek(int64(currentBlock.Offset), io.SeekStart)
			blockBuffer := make([]byte, currentBlock.Size)
			n, _ := file.Read(blockBuffer)

			if uint32(n) != currentBlock.Size {
				log.Fatal("Reading Error")
			}
			blockBuffer = blockBuffer[:n]
			decryptBuffer := make([]byte, currentBlock.Size)
			decompBuffer := make([]byte, 262144)

			if currentBlock.BitFlag&0x2 != 0 {
				decryptBuffer = decrypt(currentBlock, blockBuffer)
			} else {
				decryptBuffer = blockBuffer
			}
			if currentBlock.BitFlag&0x1 != 0 {
				decompBuffer = oodleDecompressBlock(currentBlock, decryptBuffer)
			} else {
				decompBuffer = decryptBuffer
			}
			if blockIndex == currentEntry.StartingBlock {
				size := 0
				if blockIndex == lastBlockIndex {
					size = int(currentEntry.FileSize)
				} else {
					size = int(262144 - currentEntry.StartingBlockOffset)
				}

				copy(fileBuffer[0:size], decompBuffer[currentEntry.StartingBlockOffset:currentEntry.StartingBlockOffset+uint32(size)]) //append(fileBuffer[:size], decompBuffer[currentEntry.StartingBlockOffset:currentEntry.StartingBlockOffset+uint32(size)]...)
				curBufferOffset += size
			} else if blockIndex == lastBlockIndex {
				copy(fileBuffer[curBufferOffset:], decompBuffer[:currentEntry.FileSize-uint32(curBufferOffset)])
			} else {
				copy(fileBuffer[curBufferOffset:(curBufferOffset+262144)], decompBuffer[0:262144])
				curBufferOffset += 262144
			}
			defer file.Close()
			blockIndex += 1
			decompBuffer = nil
		}
		path := model.OutputPath + "/" + fmt.Sprintf("%x", header_data.PkgID) + "-" + fmt.Sprintf("%x", i) + ".bin"
		err := os.WriteFile(path, fileBuffer, 0666)
		if err != nil {
			log.Fatal(err)
		}

	}
	fmt.Println("Wrote .bin files to output path")
}

func main() {
	model.Path = "D:/SteamLibrary/steamapps/common/Destiny 2/packages"
	model.OutputPath = "./output"
	model.PackagePath = latestPatchId("018a")
	//fmt.Println(model.PackagePath)
	extract()
	//fmt.Printf("Package ID: %d\nPatch ID: %d\nEntry Table Offset: %d\nEntry Table Count: %d", g.PkgID, g.PatchID, g.EntryTableOffset, g.EntryTableCount)
}
