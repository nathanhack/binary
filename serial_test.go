package binary

import (
	"reflect"
	"strconv"
	"testing"
)

type Thing struct {
	V1 byte   `endian:"little"`
	V2 []byte `size:"4"`
}

func TestEnc2(t *testing.T) {
	input := Thing{
		V1: 0,
		V2: []byte{1, 2, 3, 4},
	}
	expected := []byte{0, 1, 2, 3, 4}

	actual, err := Encode(&input)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected \n%v\n but found \n%v\n", expected, actual)
	}
}

func TestEnc(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected []byte
	}{
		0: {
			struct {
				V0 byte   `bits:"1"`
				V1 uint8  `bits:"1"`
				V2 uint16 `bits:"1"`
				V3 uint32 `bits:"1"`
				V4 uint64 `bits:"1"`
				V5 struct {
					I0 byte `bits:"1"`
				}
				V6  [1]byte   `bits:"1"`
				V7  [1]uint8  `bits:"1"`
				V8  [1]uint16 `bits:"1"`
				V9  [1]uint32 `bits:"1"`
				V10 [1]uint64 `bits:"1"`
				v11 [1]struct {
					I0 byte `bits:"1"`
				}
				V12 []byte   `size:"1" bits:"1"`
				V13 []uint8  `size:"1" bits:"1"`
				V14 []uint16 `size:"1" bits:"1"`
				V15 []uint32 `size:"1" bits:"1"`
				V16 []uint64 `size:"1" bits:"1"`
				V17 []struct {
					I0 byte `bits:"1"`
				} `size:"1"`
				V18 [][2]byte `size:"2" bits:"1"` //note the size only applies to the slice (nested slices are not supported)
				V19 [2][]byte `size:"3" bits:"1"` //note the size only applies to the slice (nested slices are not supported)
				V20 [][]byte  `size:"1" bits:"1"` //note size will be applied to all nested slices
				V21 [][]byte  `size:"2" bits:"1"` //note size will be applied to all nested slices
				V22 struct{ I0 struct{ I0 byte } }
			}{},
			[]byte{0, 0, 0, 0, 0, 0},
		},

		1: {
			struct {
				V0  byte
				V1  uint8
				V2  uint16
				V3  uint32
				V4  uint64
				V6  int8
				V7  int16
				V8  int32
				V9  int64
				V10 string `strlen:"2"`
				V11 struct{ I0 byte }
				V12 [1]byte
				V13 [1]uint8
				V14 [1]uint16
				V15 [1]uint32
				V16 [1]uint64
				V17 [1]int8
				V18 [1]int16
				V19 [1]int32
				V20 [1]int64
				V21 [1]string `strlen:"2"`
				v22 [1]struct{ I0 byte }
				V23 []byte              `size:"1"`
				V24 []uint8             `size:"1"`
				V25 []uint16            `size:"1"`
				V26 []uint32            `size:"1"`
				V27 []uint64            `size:"1"`
				V28 []int8              `size:"1"`
				V29 []int16             `size:"1"`
				V30 []int32             `size:"1"`
				V31 []int64             `size:"1"`
				V32 []string            `size:"1" strlen:"2"`
				V33 []struct{ I0 byte } `size:"1"`
				V34 [][2]byte           `size:"2"`   //note the size only applies to the slice (nested slices are not supported)
				V35 [2][]byte           `size:"3"`   //note the size only applies to the slice (nested slices are not supported)
				V36 [2][2][2]string     `strlen:"2"` //note nested arrays are supported
				V37 [][]byte            `size:"1"`   //note size will be applied to all nested slices
				V38 [][]byte            `size:"2"`   //note size will be applied to all nested slices
				V39 struct{ I0 struct{ I0 byte } }
			}{},
			[]byte{
				0,
				0,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				0,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				32, 32,
				0,
				0,
				0,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				0,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				32, 32,
				0,
				0,
				0,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				0,
				0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				32, 32,
				0,
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32,
				0,
				0, 0, 0, 0,
				0,
			},
		},
		2: {
			struct {
				V1 string `strlen:"4"`
			}{"Hello"},
			[]byte("Hell"),
		},
		3: {
			struct {
				V1 string
			}{"Hello"},
			[]byte("Hello"),
		},
		4: {
			struct {
				V1 string `strlen:"4"`
			}{},
			[]byte{32, 32, 32, 32},
		},
		5: {
			struct {
				V1 string
			}{},
			[]byte{},
		},
		6: {
			struct {
				V1 []int
			}{},
			[]byte{},
		},
		7: {
			struct {
				V1 int16
				V2 []int
			}{},
			[]byte{0, 0},
		},
		8: {
			struct {
				V1 [4]int32 `endian:"big"`
			}{},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		9: {
			struct {
				V1 [4]int32 `endian:"big"`
			}{
				[4]int32{1, 2, 3, 4},
			},
			[]byte{0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4},
		},
		10: {
			Thing{},
			[]byte{0, 0, 0, 0, 0},
		},
		11: {
			struct {
				V1 Thing
				V2 struct {
					V3 Thing
				}
			}{},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		12: {
			&Thing{
				V1: 0,
				V2: []byte{1, 2, 3, 4},
			},
			[]byte{0, 1, 2, 3, 4},
		},
		13: {
			struct {
				V1 byte
				V2 []byte `size:"4"`
			}{
				V1: 0,
				V2: []byte{1, 2, 3, 4},
			},
			[]byte{0, 1, 2, 3, 4},
		},
		14: {
			struct {
				V1 byte
				V2 []byte `size:"2"`
			}{
				V1: 0,
				V2: []byte{1, 2, 3, 4},
			},
			[]byte{0, 1, 2},
		},
		15: {
			struct {
				V1 int32
				V2 int64
			}{
				V1: 1,
				V2: 2,
			},
			[]byte{1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0},
		},
		16: {
			struct {
				V1 int32 `endian:"big"`
				V2 int64 `endian:"little"`
			}{
				V1: 1,
				V2: 2,
			},
			[]byte{0, 0, 0, 1, 2, 0, 0, 0, 0, 0, 0, 0},
		},
		17: {
			struct {
				V1 struct {
					I0 int32 //note the struct's endian tag is ignored (defaults to little)
					I1 int64 `endian:"little"`
				} `endian:"big"`
			}{
				V1: struct {
					I0 int32
					I1 int64 `endian:"little"`
				}{
					1, 2,
				},
			},
			[]byte{1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0},
		},
		18: {
			struct {
				V1 *int32
				V2 [1]*int64
			}{},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},

		19: {
			struct {
				V1 int32
				V2 []int32 `size:"V1"`
			}{},
			[]byte{0, 0, 0, 0},
		},

		20: {
			struct {
				V1 int32
				V2 []int32 `size:"V1"`
			}{
				V1: 1,
			},
			[]byte{1, 0, 0, 0, 0, 0, 0, 0},
		},
		21: {
			struct {
				V1 float32
				V2 float64
				V3 float32 `endian:"little"`
				V4 float64 `endian:"big"`
				V5 float32 `endian:"big"`
				V6 float64 `endian:"little"`
			}{0, 0, 1, 1, 1, 1},
			[]byte{
				0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 128, 63,
				63, 240, 0, 0, 0, 0, 0, 0,
				63, 128, 0, 0,
				0, 0, 0, 0, 0, 0, 240, 63,
			},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, err := Encode(test.input)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(test.expected, actual) {
				t.Fatalf("expected \n%v\n but found \n%v\n", test.expected, actual)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	var input interface{} = struct{ I uint64 }{42}
	var store struct{ I uint64 }

	inputBytes, err := Encode(input)
	if err != nil {
		t.Fatal(err)
	}

	err = Decode(inputBytes, &store)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(store, input) {
		t.Fatalf("expected %v but found %v", input, store)
	}
}

func TestDecode2(t *testing.T) {
	tests := []struct {
		input interface{}
		empty func() interface{}
	}{
		{
			&struct {
				V0  uint16  `endian:"little"`
				V1  uint32  `endian:"little"`
				V2  uint64  `endian:"little"`
				V3  int16   `endian:"little"`
				V4  int32   `endian:"little"`
				V5  int64   `endian:"little"`
				V6  float32 `endian:"little"`
				V7  float64 `endian:"little"`
				V8  uint16  `endian:"big"`
				V9  uint32  `endian:"big"`
				V10 uint64  `endian:"big"`
				V11 int16   `endian:"big"`
				V12 int32   `endian:"big"`
				V13 int64   `endian:"big"`
				V14 float32 `endian:"big"`
				V15 float64 `endian:"big"`
				V16 bool
			}{
				V0:  0xAB,
				V1:  0xABCD,
				V2:  0xABCDEFA,
				V3:  0xAB,
				V4:  0xABCD,
				V5:  0xABCDEFA,
				V6:  1.9999999,
				V7:  1.9999999,
				V8:  0xAB,
				V9:  0xABCD,
				V10: 0xABCDEFA,
				V11: 0xAB,
				V12: 0xABCD,
				V13: 0xABCDEFA,
				V14: 1.9999999,
				V15: 1.9999999,
				V16: true,
			},
			func() interface{} {
				var i struct {
					V0  uint16  `endian:"little"`
					V1  uint32  `endian:"little"`
					V2  uint64  `endian:"little"`
					V3  int16   `endian:"little"`
					V4  int32   `endian:"little"`
					V5  int64   `endian:"little"`
					V6  float32 `endian:"little"`
					V7  float64 `endian:"little"`
					V8  uint16  `endian:"big"`
					V9  uint32  `endian:"big"`
					V10 uint64  `endian:"big"`
					V11 int16   `endian:"big"`
					V12 int32   `endian:"big"`
					V13 int64   `endian:"big"`
					V14 float32 `endian:"big"`
					V15 float64 `endian:"big"`
					V16 bool
				}
				return &i
			},
		},
		{
			&struct {
				V0 *struct{ I0 uint32 }
				V1 []*struct{ I1 byte } `size:"1"`
			}{
				&struct{ I0 uint32 }{1},
				[]*struct{ I1 byte }{{22}},
			},
			func() interface{} {
				var i struct {
					V0 *struct{ I0 uint32 }
					V1 []*struct{ I1 byte } `size:"1"`
				}
				return &i
			},
		},
		{
			&struct {
				V1  uint16
				V2  int32
				V10 string `strlen:"2"`
				V11 struct{ I0 byte }
				V12 [1]byte
				V13 [1]uint8
				V14 [1]uint16
				V15 [1]uint32
				V16 [1]uint64
				V17 [1]int8
				V18 [1]int16
				V19 [1]int32
				V20 [1]int64
				V21 [1]string `strlen:"2"`
				V22 [1]struct{ I1 byte }
				V23 []byte                `size:"V1"`
				V24 []uint8               `size:"V2"`
				V25 []uint16              `size:"1"`
				V26 []uint32              `size:"1"`
				V27 []uint64              `size:"1"`
				V28 []int8                `size:"1"`
				V29 []int16               `size:"1"`
				V30 []int32               `size:"1"`
				V31 []int64               `size:"1"`
				V32 []string              `size:"1" strlen:"2"`
				V33 []struct{ I2 byte }   `size:"1"`
				V34 [][2]byte             `size:"2"`   //note the size only applies to the slice (nested slices are not supported)
				V35 [2][]byte             `size:"3"`   //note the size only applies to the slice (nested slices are not supported)
				V36 [2][2][2]string       `strlen:"2"` //note nested arrays are supported
				V37 [][]byte              `size:"1"`   //note size will be applied to all nested slices
				V38 [][]struct{ I1 byte } `size:"2"`   //note size will be applied to all nested slices
				V39 struct{ I3 struct{ I4 byte } }
				V40 *struct{ I0 uint32 }
				V41 []*struct{ I1 byte } `size:"1"`
			}{
				2,
				1,
				"10",
				struct{ I0 byte }{11},
				[1]byte{12},
				[1]uint8{13},
				[1]uint16{14},
				[1]uint32{15},
				[1]uint64{16},
				[1]int8{17},
				[1]int16{18},
				[1]int32{19},
				[1]int64{20},
				[1]string{"21"},
				[1]struct{ I1 byte }{{22}},
				[]byte{23, 0},
				[]uint8{24},
				[]uint16{25},
				[]uint32{26},
				[]uint64{27},
				[]int8{28},
				[]int16{29},
				[]int32{30},
				[]int64{31},
				[]string{"32"},
				[]struct{ I2 byte }{{33}},
				[][2]byte{{34, 34}, {34, 34}},
				[2][]byte{{35, 35, 35}, {35, 35, 35}},
				[2][2][2]string{{{"36", "36"}, {"36", "36"}}, {{"36", "36"}, {"36", "36"}}},
				[][]byte{{37}},
				[][]struct{ I1 byte }{{{22}, {23}}, {{21}, {24}}},
				struct{ I3 struct{ I4 byte } }{struct{ I4 byte }{39}},
				&struct{ I0 uint32 }{1},
				[]*struct{ I1 byte }{{22}},
			},
			func() interface{} {
				var i struct {
					V1  uint16
					V2  int32
					V10 string `strlen:"2"`
					V11 struct{ I0 byte }
					V12 [1]byte
					V13 [1]uint8
					V14 [1]uint16
					V15 [1]uint32
					V16 [1]uint64
					V17 [1]int8
					V18 [1]int16
					V19 [1]int32
					V20 [1]int64
					V21 [1]string `strlen:"2"`
					V22 [1]struct{ I1 byte }
					V23 []byte                `size:"V1"`
					V24 []uint8               `size:"V2"`
					V25 []uint16              `size:"1"`
					V26 []uint32              `size:"1"`
					V27 []uint64              `size:"1"`
					V28 []int8                `size:"1"`
					V29 []int16               `size:"1"`
					V30 []int32               `size:"1"`
					V31 []int64               `size:"1"`
					V32 []string              `size:"1" strlen:"2"`
					V33 []struct{ I2 byte }   `size:"1"`
					V34 [][2]byte             `size:"2"`   //note the size only applies to the slice (nested slices are not supported)
					V35 [2][]byte             `size:"3"`   //note the size only applies to the slice (nested slices are not supported)
					V36 [2][2][2]string       `strlen:"2"` //note nested arrays are supported
					V37 [][]byte              `size:"1"`   //note size will be applied to all nested slices
					V38 [][]struct{ I1 byte } `size:"2"`   //note size will be applied to all nested slices
					V39 struct{ I3 struct{ I4 byte } }
					V40 *struct{ I0 uint32 }
					V41 []*struct{ I1 byte } `size:"1"`
				}
				return &i
			},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			inputBytes, err := Encode(test.input)
			if err != nil {
				t.Fatal(err)
			}

			e := test.empty()
			err = Decode(inputBytes, e)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(test.input, e) {
				t.Fatalf("expected \n%+v\n but found \n%+v\n", test.input, e)
			}
		})
	}
}

func TestDecodeNestedStructSlices(t *testing.T) {
	type Stuff struct {
		Z0 byte
		V1 []byte `size:"Z0"`
		V2 [3]byte
		V3 []byte `size:"3"`
	}

	type Output struct {
		V0  uint32
		V2  []byte `size:"V0"`
		V3  uint32
		V44 []Stuff `size:"1"`
		V5  []byte  `size:"16"`
		V6  []byte  `size:"V3"`
	}
	o1 := Output{
		V0: 3,
		V2: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2}, //truncates to len()==V0
		V3: 3,
		V6: []byte{6}, //adds zeros to len()==V3
	}
	bs, err := Encode(o1)
	if err != nil {
		t.Fatal(err)
	}

	var o2 Output
	err = Decode(bs, &o2)
	if err != nil {
		t.Fatal(err)
	}

	expected := Output{
		V0:  3,
		V2:  []byte{2, 2, 2},
		V3:  3,
		V44: []Stuff{{Z0: 0, V1: []byte{}, V2: [3]byte{0, 0, 0}, V3: []byte{0, 0, 0}}},
		V5:  []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		V6:  []byte{6, 0, 0},
	}

	if !reflect.DeepEqual(expected, o2) {
		t.Fatalf("expected %v but found %v", expected, o2)
	}
}

func TestDecodeExampleIpv4Header(t *testing.T) {
	type IpHeader struct {
		IHL            uint8  `bits:"4" ` //note IHL & Ver seem out of order, it's due to when bytes are written
		Version        uint8  `bits:"4" ` //note IHL & Ver seem out of order, it's due to when bytes are written
		ECN            uint8  `bits:"2" ` //note ECN & DSCP order are swapped, because of the way bytes are written
		DSCP           uint8  `bits:"6" ` //note ECN & DSCP order are swapped, because of the way bytes are written
		Length         uint16 `endian:"big"`
		Identification uint16 `endian:"big"`
		Reserved       bool   `bits:"1"`
		DontFrag       bool   `bits:"1"`
		MoreFrag       bool   `bits:"1"`
		FragOffset     uint16 `bits:"13" endian:"big"`
		TTL            uint8  `endian:"big"`
		Protocol       uint8  `endian:"big"`
		Checksum       uint16 `endian:"big"`
		Source         uint32 `endian:"big"`
		Destination    uint32 `endian:"big"`
	}

	o1 := IpHeader{
		Version:        4,
		IHL:            5,
		DSCP:           0,
		ECN:            0,
		Length:         24,
		Identification: 242,
		Reserved:       false,
		DontFrag:       false,
		MoreFrag:       false,
		FragOffset:     3,
		TTL:            64,
		Protocol:       17,
		Checksum:       0xcf54,
		Source:         0x0a010101, //10.1.1.1
		Destination:    0x0a010101, //10.1.1.1
	}
	bs, err := Encode(o1)
	if err != nil {
		t.Fatal(err)
	}

	expectedBs := []byte{0x45, 0x00, 0x00, 0x18, 0x00, 0xf2, 0x00, 0x03, 0x40, 0x11, 0xcf, 0x54, 0x0a, 0x01, 0x01, 0x01, 0x0a, 0x01, 0x01, 0x01}

	if !reflect.DeepEqual(bs, expectedBs) {
		t.Fatalf("expected \n%x\n but found \n%x\n", expectedBs, bs)
	}

	var o2 IpHeader
	err = Decode(bs, &o2)
	if err != nil {
		t.Fatal(err)
	}

	expected := IpHeader{
		Version:        4,
		IHL:            5,
		DSCP:           0,
		ECN:            0,
		Length:         24,
		Identification: 242,
		Reserved:       false,
		DontFrag:       false,
		MoreFrag:       false,
		FragOffset:     3,
		TTL:            64,
		Protocol:       17,
		Checksum:       0xcf54,
		Source:         0x0a010101,
		Destination:    0x0a010101,
	}

	if !reflect.DeepEqual(expected, o2) {
		t.Fatalf("expected %v but found %v", expected, o2)
	}
}

func TestSizeOf(t *testing.T) {

	type T1 struct {
		I0  bool
		I1  int8
		I2  int16
		I3  int32
		I4  int64
		I5  uint8
		I6  uint16
		I7  uint32
		I8  uint64
		I9  bool     `bits:"1"`
		I10 uint8    `bits:"1"`
		I11 uint16   `bits:"1"`
		I12 uint32   `bits:"1"`
		I13 uint64   `bits:"1"`
		I14 string   `strlen:"1"'`
		I15 []uint8  `size:"1"`
		I16 []uint64 `size:"1"`
	}
	expect := 42
	b := T1{}

	actual := SizeOf(b)
	if actual != expect {
		t.Fatalf("expected %v but found %v", expect, actual)
	}

	actual = SizeOf(&b)
	if actual != expect {
		t.Fatalf("expected %v but found %v", expect, actual)
	}
}
