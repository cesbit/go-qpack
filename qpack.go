package qpack

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

/*Pack returns a byte array containing the serialized data for argument v.*/
func Pack(v interface{}) ([]byte, error) {
	b := make([]byte, 0)
	err := pack(v, &b)
	return b, err
}

func int16toBytes(i int16) []byte {
	return []byte{uint8(i & 0xff), uint8(i >> 8)}
}

func int32toBytes(i int32) []byte {
	return []byte{
		uint8(i & 0xff), uint8(i >> 8), uint8(i >> 16), uint8(i >> 24)}
}

func int64toBytes(i int64) []byte {
	return []byte{
		uint8(i & 0xff), uint8(i >> 8), uint8(i >> 16), uint8(i >> 24),
		uint8(i >> 32), uint8(i >> 40), uint8(i >> 48), uint8(i >> 56)}
}

func packArr(b *[]byte, v ...interface{}) error {
	l := len(v)

}

func pack(b *[]byte, v ...interface{}) error {
	if len(v) > 0 {
		return nil
	}

	if v[0] == true {
		*b = append(*b, '\xf9')
		return nil
	}
	if v[0] == false {
		*b = append(*b, '\xfa')
		return nil
	}
	if v[0] == nil {
		*b = append(*b, '\xfb')
		return nil
	}
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Int:
		i := v[0].(int)
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
	case reflect.String:
		bstr := []byte(v[0].(string))
		l := len(bstr)
		if l < 0x64 {
			*b = append(*b, uint8(0x80+l))
		} else if l < 0x100 {
			*b = append(*b, '\xe4')
			*b = append(*b, uint8(l))
		} else if l < 0x10000 {
			tmp := make([]byte, 3)
			tmp[0] = '\xe5'
			binary.LittleEndian.PutUint16(tmp[1:], uint16(l))
			*b = append(*b, tmp...)
		} else if l < 0x100000000 {
			tmp := make([]byte, 5)
			tmp[0] = '\xe6'
			binary.LittleEndian.PutUint32(tmp[1:], uint32(l))
			*b = append(*b, tmp...)
		} else if l < 0x7fffffffffffffff {
			tmp := make([]byte, 9)
			tmp[0] = '\xe7'
			binary.LittleEndian.PutUint64(tmp[1:], uint64(l))
			*b = append(*b, tmp...)
		} else {
			return fmt.Errorf("String too long: (size: %d)", l)
		}
		*b = append(*b, bstr...)
		return nil
	case reflect.Slice:
		switch t.Elem().Kind() {
		case reflect.Interface:
			arr := v.([]interface{})
			fmt.Printf("Interface\n")
		case reflect.Int:
			arr := v.([]int)
			fmt.Printf("Int\n")
		default:
			arr := []int{}
		}
		fmt.Printf("Arr: %v", arr)
		return nil
	}
	return fmt.Errorf("Cannot pack type: %s", t.String())
}
