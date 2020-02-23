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
		{
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

		{
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
		{
			struct {
				V1 string `strlen:"4"`
			}{"Hello"},
			[]byte("Hell"),
		},
		{
			struct {
				V1 string
			}{"Hello"},
			[]byte("Hello"),
		},
		{
			struct {
				V1 string `strlen:"4"`
			}{},
			[]byte{32, 32, 32, 32},
		},
		{
			struct {
				V1 string
			}{},
			[]byte{},
		},
		{
			struct {
				V1 []int
			}{},
			[]byte{},
		},
		{
			struct {
				V1 int16
				V2 []int
			}{},
			[]byte{0, 0},
		},
		{
			struct {
				V1 [4]int32 `endian:"big"`
			}{},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			struct {
				V1 [4]int32 `endian:"big"`
			}{
				[4]int32{1, 2, 3, 4},
			},
			[]byte{0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4},
		},
		{
			Thing{},
			[]byte{0, 0, 0, 0, 0},
		},
		{
			struct {
				V1 Thing
				V2 struct {
					V3 Thing
				}
			}{},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			&Thing{
				V1: 0,
				V2: []byte{1, 2, 3, 4},
			},
			[]byte{0, 1, 2, 3, 4},
		},
		{
			struct {
				V1 byte
				V2 []byte `size:"4"`
			}{
				V1: 0,
				V2: []byte{1, 2, 3, 4},
			},
			[]byte{0, 1, 2, 3, 4},
		},
		{
			struct {
				V1 byte
				V2 []byte `size:"2"`
			}{
				V1: 0,
				V2: []byte{1, 2, 3, 4},
			},
			[]byte{0, 1, 2},
		},
		{
			struct {
				V1 int32
				V2 int64
			}{
				V1: 1,
				V2: 2,
			},
			[]byte{1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			struct {
				V1 int32 `endian:"big"`
				V2 int64 `endian:"little"`
			}{
				V1: 1,
				V2: 2,
			},
			[]byte{0, 0, 0, 1, 2, 0, 0, 0, 0, 0, 0, 0},
		},
		{
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
		{
			struct {
				V1 *int32
				V2 [1]*int64
			}{},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},

		{
			struct {
				V1 int32
				V2 []int32 `size:"V1"`
			}{},
			[]byte{0, 0, 0, 0},
		},

		{
			struct {
				V1 int32
				V2 []int32 `size:"V1"`
			}{
				V1: 1,
			},
			[]byte{1, 0, 0, 0, 0, 0, 0, 0},
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
