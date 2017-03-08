package qpack

import (
	"bytes"
	"testing"
)

func TestUnpack(t *testing.T) {
	var m = make(map[interface{}]interface{})
	m["Names"] = []string{"Iris", "Sasha"}

	cases := []struct {
		in  []byte
		err error
	}{
		{[]byte{
			140, 72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100},
			nil}, // Hello, world
		{[]byte{249}, nil},                                        // false
		{[]byte{250}, nil},                                        // true
		{[]byte{251}, nil},                                        // nil
		{[]byte{64}, nil},                                         // -1
		{[]byte{123}, nil},                                        // -60
		{[]byte{232, 195}, nil},                                   // -61
		{[]byte{0}, nil},                                          // 0
		{[]byte{1}, nil},                                          // 1
		{[]byte{63}, nil},                                         // 63
		{[]byte{232, 64}, nil},                                    // 6
		{[]byte{125}, nil},                                        // -1.0
		{[]byte{126}, nil},                                        // 0.0
		{[]byte{127}, nil},                                        // 1.0
		{[]byte{232, 136}, nil},                                   // -120
		{[]byte{233, 2, 255}, nil},                                // -0xfe
		{[]byte{234, 70, 35, 1, 255}, nil},                        // -0xfedcba
		{[]byte{235, 138, 103, 69, 35, 1, 255, 255, 255}, nil},    // -0xfedcba9876
		{[]byte{232, 120}, nil},                                   // 120
		{[]byte{233, 254, 0}, nil},                                // 0xfe
		{[]byte{234, 186, 220, 254, 0}, nil},                      // 0xfedcba
		{[]byte{235, 118, 152, 186, 220, 254, 0, 0, 0}, nil},      // 0xfedcba9876
		{[]byte{236, 135, 136, 155, 83, 201, 192, 243, 191}, nil}, // -1.234567
		{[]byte{236, 83, 5, 163, 146, 58, 221, 94, 64}, nil},      // 123.4567
		{[]byte{
			240, 126, 236, 154, 153, 153, 153, 153, 153, 241, 63,
			236, 154, 153, 153, 153, 153, 153, 1, 64}, nil}, // [0.0, 1.1, 2.2]
		{[]byte{242, 10, 20, 30, 40, 50}, nil},          // [10, 20, 30, 40, 50]
		{[]byte{252, 10, 20, 30, 40, 50, 60, 254}, nil}, // [10, 20, 30, 40, 50, 60]
		{[]byte{
			239, 0, 244, 133, 78, 97, 109, 101, 115, 239, 132, 73, 114, 105,
			115, 133, 83, 97, 115, 104, 97}, nil},
	}
	for _, c := range cases {
		if c.err == nil {
			got, err := Unpack(c.in)
			if err != nil {
				t.Errorf(
					"Unpack(%q) returned an unexpexted error: %s", c.in, c.err)
			} else {
				packed, err := Pack(got)
				if err != nil {
					t.Errorf("Unpack(%q) => Pack(%q) returned an error: %s", c.in, got, err)
				} else if !bytes.Equal(packed, c.in) {
					t.Errorf("Unpack(%q) => Pack(%q) == %v", c.in, got, packed)
				}
			}
		} else {
			_, err := Unpack(c.in)
			t.Errorf("Error: %v", err)
		}
	}
}
