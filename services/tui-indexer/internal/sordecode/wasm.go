package sordecode

import (
	"bytes"
	"fmt"
)

const (
	wasmMagic               = "\x00asm"
	contractSpecSectionName = "contractspecv0"
)

// ExtractSpecFromWASM reads the contractspecv0 custom section from a Soroban WASM module.
func ExtractSpecFromWASM(wasmBytes []byte) ([]byte, error) {
	if len(wasmBytes) < 8 || string(wasmBytes[:4]) != wasmMagic {
		return nil, fmt.Errorf("invalid wasm magic")
	}

	reader := bytes.NewReader(wasmBytes[8:])
	for reader.Len() > 0 {
		sectionType, err := reader.ReadByte()
		if err != nil {
			break
		}
		sectionSize, err := readULEB128(reader)
		if err != nil {
			break
		}
		if sectionSize > uint64(reader.Len()) {
			break
		}
		sectionData := make([]byte, sectionSize)
		if _, err := reader.Read(sectionData); err != nil {
			break
		}
		if sectionType != 0 {
			continue
		}

		sectionReader := bytes.NewReader(sectionData)
		nameLen, err := readULEB128(sectionReader)
		if err != nil || nameLen > uint64(sectionReader.Len()) {
			continue
		}
		name := make([]byte, nameLen)
		if _, err := sectionReader.Read(name); err != nil {
			continue
		}
		if string(name) != contractSpecSectionName {
			continue
		}

		specData := make([]byte, sectionReader.Len())
		if _, err := sectionReader.Read(specData); err != nil {
			continue
		}
		return specData, nil
	}

	return nil, fmt.Errorf("contractspecv0 section not found")
}

func readULEB128(reader *bytes.Reader) (uint64, error) {
	var result uint64
	var shift uint
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
		if shift >= 64 {
			return 0, fmt.Errorf("LEB128 overflow")
		}
	}
	return result, nil
}
