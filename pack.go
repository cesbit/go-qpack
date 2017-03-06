// Package qpack provides fast data (de)serializing.
// Maintainer: Jeroen van der Heijden / Transceptor Technology
package qpack

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

// Pack returns a byte array containing the serialized data for argument v.
func Pack(v interface{}) ([]byte, error) {
	b := make([]byte, 0)
	err := pack(&b, v)
	return b, err
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
