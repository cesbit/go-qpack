package qpack

import (
	"encoding/binary"
	"math"
)

func int16FromBytes(b []byte) int16 {
	return int16(b[0]) + int16(b[1])<<8
}

func int16toBytes(i int16) []byte {
	return []byte{uint8(i & 0xff), uint8(i >> 8)}
}

func int32FromBytes(b []byte) int32 {
	return int32(b[0]) + int32(b[1])<<8 + int32(b[2])<<16 + int32(b[3])<<24
}

func int32toBytes(i int32) []byte {
	return []byte{
		uint8(i & 0xff), uint8(i >> 8), uint8(i >> 16), uint8(i >> 24)}
}

func int64FromBytes(b []byte) int64 {
	return int64(b[0]) + int64(b[1])<<8 + int64(b[2])<<16 + int64(b[3])<<24 +
		int64(b[4])<<32 + int64(b[5])<<40 + int64(b[6])<<48 + int64(b[7])<<56
}

func int64toBytes(i int64) []byte {
	return []byte{
		uint8(i & 0xff), uint8(i >> 8), uint8(i >> 16), uint8(i >> 24),
		uint8(i >> 32), uint8(i >> 40), uint8(i >> 48), uint8(i >> 56)}
}

func float64FromBytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func float64toBytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
