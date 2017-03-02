package qpack

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

// Pack returns a byte array containing the serialized data for argument v.
func Pack(v interface{}) ([]byte, error) {
	b := make([]byte, 0)
	err := pack(&b, v)
	return b, err
}

// Unpack return an interface containing deserialized data.
func Unpack(b []byte) (interface{}, error) {
	var v interface{}
	pos := 0
	err := unpack(&b, &v, &pos, len(b))
	return v, err
}

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

func pack(b *[]byte, v interface{}) error {

	if v == true {
		*b = append(*b, '\xf9')
		return nil
	}
	if v == false {
		*b = append(*b, '\xfa')
		return nil
	}
	if v == nil {
		*b = append(*b, '\xfb')
		return nil
	}
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Int:
		i := v.(int)
		if i >= 0 && i < 64 {
			*b = append(*b, uint8(i))
		} else if i >= -60 && i < 0 {
			*b = append(*b, uint8(63-i))
		} else if int(int8(i)) == i {
			*b = append(*b, '\xe8')
			*b = append(*b, byte(i))
		} else if int(int16(i)) == i {
			*b = append(*b, '\xe9')
			*b = append(*b, int16toBytes(int16(i))...)
		} else if int(int32(i)) == i {
			*b = append(*b, '\xea')
			*b = append(*b, int32toBytes(int32(i))...)
		} else if int(int64(i)) == i {
			*b = append(*b, '\xeb')
			*b = append(*b, int64toBytes(int64(i))...)
		} else {
			return fmt.Errorf("Cannot int: %d", i)
		}
		return nil
	case reflect.Float64:
		f := v.(float64)
		if f == 0.0 {
			*b = append(*b, '\x7e')
		} else if f == -1.0 {
			*b = append(*b, '\x7d')
		} else if f == 1.0 {
			*b = append(*b, '\x7f')
		} else {
			*b = append(*b, '\xec')
			*b = append(*b, float64toBytes(f)...)
		}
		return nil
	case reflect.String:
		bstr := []byte(v.(string))
		l := len(bstr)
		switch {
		case l < 0x64:
			*b = append(*b, uint8(0x80+l))
		case l < 0x100:
			*b = append(*b, '\xe4')
			*b = append(*b, uint8(l))
		case l < 0x10000:
			tmp := make([]byte, 3)
			tmp[0] = '\xe5'
			binary.LittleEndian.PutUint16(tmp[1:], uint16(l))
			*b = append(*b, tmp...)
		case l < 0x100000000:
			tmp := make([]byte, 5)
			tmp[0] = '\xe6'
			binary.LittleEndian.PutUint32(tmp[1:], uint32(l))
			*b = append(*b, tmp...)
		case l < 0x7fffffffffffffff:
			tmp := make([]byte, 9)
			tmp[0] = '\xe7'
			binary.LittleEndian.PutUint64(tmp[1:], uint64(l))
			*b = append(*b, tmp...)
		default:
			return fmt.Errorf("String too long: (size: %d)", l)
		}
		*b = append(*b, bstr...)
		return nil
	case reflect.Slice:
		slice := reflect.ValueOf(v)
		n := slice.Len()
		if n >= 6 {
			*b = append(*b, '\xfc') // Open array
		} else {
			*b = append(*b, uint8('\xed'+n)) // Fixed size array
		}
		for i := 0; i < n; i++ {
			err := pack(b, slice.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		if n >= 6 {
			*b = append(*b, '\xfe') // Close array
		}
		return nil
	case reflect.Map:
		m := reflect.ValueOf(v)
		n := m.Len()
		if n >= 6 {
			*b = append(*b, '\xfd') // Open map
		} else {
			*b = append(*b, uint8('\xf3'+n)) // Fixed size map
		}
		for _, key := range m.MapKeys() {
			err := pack(b, key.Interface())
			if err != nil {
				return err
			}
			val := m.MapIndex(key)
			err = pack(b, val.Interface())
			if err != nil {
				return err
			}
		}
		if n >= 6 {
			*b = append(*b, '\xff') // Close map
		}
		return nil
	}
	return fmt.Errorf("Cannot pack type: %s", t.String())
}

func unpack(b *[]byte, v *interface{}, pos *int, end int) error {
	if *pos >= end {
		return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
	}
	tp := (*b)[*pos]
	*pos++
	switch tp {
	case '\x00', '\x01', '\x02', '\x03', '\x04', '\x05', '\x06', '\x07', '\x08',
		'\x09', '\x0a', '\x0b', '\x0c', '\x0d', '\x0e', '\x0f', '\x10', '\x11',
		'\x12', '\x13', '\x14', '\x15', '\x16', '\x17', '\x18', '\x19', '\x1a',
		'\x1b', '\x1c', '\x1d', '\x1e', '\x1f', '\x20', '\x21', '\x22', '\x23',
		'\x24', '\x25', '\x26', '\x27', '\x28', '\x29', '\x2a', '\x2b', '\x2c',
		'\x2d', '\x2e', '\x2f', '\x30', '\x31', '\x32', '\x33', '\x34', '\x35',
		'\x36', '\x37', '\x38', '\x39', '\x3a', '\x3b', '\x3c', '\x3d', '\x3e',
		'\x3f':
		*v = int(tp)
	case '\x40', '\x41', '\x42', '\x43', '\x44', '\x45', '\x46', '\x47', '\x48',
		'\x49', '\x4a', '\x4b', '\x4c', '\x4d', '\x4e', '\x4f', '\x50', '\x51',
		'\x52', '\x53', '\x54', '\x55', '\x56', '\x57', '\x58', '\x59', '\x5a',
		'\x5b', '\x5c', '\x5d', '\x5e', '\x5f', '\x60', '\x61', '\x62', '\x63',
		'\x64', '\x65', '\x66', '\x67', '\x68', '\x69', '\x6a', '\x6b', '\x6c',
		'\x6d', '\x6e', '\x6f', '\x70', '\x71', '\x72', '\x73', '\x74', '\x75',
		'\x76', '\x77', '\x78', '\x79', '\x7a', '\x7b':
		*v = 63 - int(tp)
	case '\x7c':
		*v = nil
	case '\x7d':
		*v = -1.0
	case '\x7e':
		*v = 0.0
	case '\x7f':
		*v = 1.0
	case '\x80', '\x81', '\x82', '\x83', '\x84', '\x85', '\x86', '\x87', '\x88',
		'\x89', '\x8a', '\x8b', '\x8c', '\x8d', '\x8e', '\x8f', '\x90', '\x91',
		'\x92', '\x93', '\x94', '\x95', '\x96', '\x97', '\x98', '\x99', '\x9a',
		'\x9b', '\x9c', '\x9d', '\x9e', '\x9f', '\xa0', '\xa1', '\xa2', '\xa3',
		'\xa4', '\xa5', '\xa6', '\xa7', '\xa8', '\xa9', '\xaa', '\xab', '\xac',
		'\xad', '\xae', '\xaf', '\xb0', '\xb1', '\xb2', '\xb3', '\xb4', '\xb5',
		'\xb6', '\xb7', '\xb8', '\xb9', '\xba', '\xbb', '\xbc', '\xbd', '\xbe',
		'\xbf', '\xc0', '\xc1', '\xc2', '\xc3', '\xc4', '\xc5', '\xc6', '\xc7',
		'\xc8', '\xc9', '\xca', '\xcb', '\xcc', '\xcd', '\xce', '\xcf', '\xd0',
		'\xd1', '\xd2', '\xd3', '\xd4', '\xd5', '\xd6', '\xd7', '\xd8', '\xd9',
		'\xda', '\xdb', '\xdc', '\xdd', '\xde', '\xdf', '\xe0', '\xe1', '\xe2',
		'\xe3':
		n := int(tp) - 128
		*pos += n
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = string((*b)[*pos-n : *pos])
	case '\xe4':
		if *pos >= end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		n := int((*b)[*pos])
		*pos += n + 1
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = string((*b)[*pos-n : *pos])
	case '\xe5':
		*pos += 2
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		n := int(binary.LittleEndian.Uint16((*b)[*pos-2 : *pos]))
		*pos += n
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = string((*b)[*pos-n : *pos])
	case '\xe6':
		*pos += 4
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		n := int(binary.LittleEndian.Uint32((*b)[*pos-4 : *pos]))
		*pos += n
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = string((*b)[*pos-n : *pos])
	case '\xe7':
		*pos += 8
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		n := int(binary.LittleEndian.Uint16((*b)[*pos-8 : *pos]))
		*pos += n
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = string((*b)[*pos-n : *pos])
	case '\xe8':
		if *pos >= end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = int(int8((*b)[*pos]))
		*pos++
	case '\xe9':
		*pos += 2
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = int(int16FromBytes((*b)[*pos-2 : *pos]))
	case '\xea':
		*pos += 4
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = int(int32FromBytes((*b)[*pos-4 : *pos]))
	case '\xeb':
		*pos += 8
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = int(int64FromBytes((*b)[*pos-8 : *pos]))
	case '\xec':
		*pos += 8
		if *pos > end {
			return fmt.Errorf("Unpack() is missing data at position: %d", *pos)
		}
		*v = float64FromBytes((*b)[*pos-8 : *pos])
	case '\xed', '\xee', '\xef', '\xf0', '\xf1', '\xf2':
		n := int(tp) - 237
		slice := make([]interface{}, n)
		for i := 0; i < n; i++ {
			err := unpack(b, &slice[i], pos, end)
			if err != nil {
				return err
			}
		}
		*v = slice
	case '\xf3', '\xf4', '\xf5', '\xf6', '\xf7', '\xf8':
		n := int(tp) - 243
		m := make(map[interface{}]interface{})
		var key, val interface{}
		for i := 0; i < n; i++ {
			err := unpack(b, &key, pos, end)
			if err != nil {
				return err
			}
			err = unpack(b, &val, pos, end)
			if err != nil {
				return err
			}
			m[key] = val
		}
		*v = m
	case '\xf9':
		*v = true
	case '\xfa':
		*v = false
	case '\xfb':
		*v = nil
	case '\xfc':
		slice := make([]interface{}, 0)
		var val interface{}
		for *pos < end && (*b)[*pos] != '\xfe' {
			err := unpack(b, &val, pos, end)
			if err != nil {
				return err
			} else if val == '\xff' {
				return fmt.Errorf(
					"Unpack() got an unexpected close map at position: %d",
					*pos)
			}
			slice = append(slice, val)
		}
		*v = slice
	case '\xfd':
		m := make(map[interface{}]interface{})
		var key, val interface{}
		for *pos < end && (*b)[*pos] != '\xff' {
			err := unpack(b, &key, pos, end)
			if err != nil {
				return err
			} else if key == '\xfe' {
				return fmt.Errorf(
					"Unpack() got an unexpected close array at position: %d",
					*pos)
			}
			err = unpack(b, &val, pos, end)
			if err != nil {
				return err
			} else if val == '\xfe' {
				return fmt.Errorf(
					"Unpack() got an unexpected close array at position: %d",
					*pos)
			}
			m[key] = val
		}
		*v = m
	case '\xfe', '\xff':
		*v = tp
	default:
		return fmt.Errorf(
			"Unpack() got an unexpected type %d at position: %d", tp, *pos)
	}

	return nil
}
