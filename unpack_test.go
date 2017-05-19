package qpack

import (
	"fmt"
	"testing"
)

func TestUnpack(t *testing.T) {
	var m = make(map[interface{}]interface{})
	m["Names"] = []string{"Iris", "Sasha"}

	cases := []struct {
		in  []byte
		out string
		err error
	}{
		{[]byte{
			140, 239, 163, 159, 32, 72, 105, 32, 81, 112, 97, 99, 107},
			" Hi Qpack", nil},
		{[]byte{244, 134, 109, 121, 110, 97, 109, 101, 132, 73, 114, 105, 115},
			"map[myname:Iris]", nil},
		{[]byte{249}, "true", nil},
		{[]byte{250}, "false", nil},
		{[]byte{251}, "<nil>", nil},
		{[]byte{64}, "-1", nil},
		{[]byte{123}, "-60", nil},
		{[]byte{232, 195}, "-61", nil},
		{[]byte{0}, "0", nil},
		{[]byte{1}, "1", nil},
		{[]byte{63}, "63", nil},
		{[]byte{232, 64}, "64", nil},
		{[]byte{125}, "-1", nil},
		{[]byte{126}, "0", nil},
		{[]byte{127}, "1", nil},
		{[]byte{232, 136}, "-120", nil},
		{[]byte{233, 2, 255}, "-254", nil},
		{[]byte{234, 70, 35, 1, 255}, "-16702650", nil},
		{[]byte{235, 138, 103, 69, 35, 1, 255, 255, 255}, "-1094624909430", nil},
		{[]byte{232, 120}, "120", nil},                                         // 120
		{[]byte{233, 254, 0}, "254", nil},                                      // 0xfe
		{[]byte{234, 186, 220, 254, 0}, "16702650", nil},                       // 0xfedcba
		{[]byte{235, 118, 152, 186, 220, 254, 0, 0, 0}, "1094624909430", nil},  // 0xfedcba9876
		{[]byte{236, 135, 136, 155, 83, 201, 192, 243, 191}, "-1.234567", nil}, // -1.234567
		{[]byte{236, 83, 5, 163, 146, 58, 221, 94, 64}, "123.4567", nil},       // 123.4567
		{[]byte{
			240, 126, 236, 154, 153, 153, 153, 153, 153, 241, 63,
			236, 154, 153, 153, 153, 153, 153, 1, 64}, "[0 1.1 2.2]", nil}, // [0.0, 1.1, 2.2]
		{[]byte{242, 10, 20, 30, 40, 50}, "[10 20 30 40 50]", nil},             // [10, 20, 30, 40, 50]
		{[]byte{252, 10, 20, 30, 40, 50, 60, 254}, "[10 20 30 40 50 60]", nil}, // [10, 20, 30, 40, 50, 60]
		{[]byte{
			239, 0, 244, 133, 78, 97, 109, 101, 115, 239, 132, 73, 114, 105,
			115, 133, 83, 97, 115, 104, 97}, "[0 map[Names:[Iris Sasha]]]", nil},
	}
	for _, c := range cases {
		if c.err == nil {
			got, err := Unpack(c.in, 0)
			if err != nil {
				t.Errorf(
					"Unpack(%q, 0) returned an unexpexted error: %s", c.in, c.err)
			} else {
				if fmt.Sprint(got) != c.out {
					t.Errorf(
						"Unpack(%q, 0) expected to return '%s' but got '%s'",
						c.in, c.out, fmt.Sprint(got))
				}
			}
		} else {
			_, err := Unpack(c.in, 0)
			t.Errorf("Error: %v", err)
		}
	}
}
