package internal

import (
	"reflect"
	"testing"
)

func TestNewFromBytes(t *testing.T) {
	b, err := NewFromBytes([]byte{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	expected := []bool{false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false}
	actual := b.Set
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestNewFromBits(t *testing.T) {
	b, err := NewFromBits([]bool{true, false, true})
	if err != nil {
		t.Fatal(err)
	}
	expected := []bool{true, false, true}
	actual := b.Set
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestBitSetBuffer_Read(t *testing.T) {
	b, err := NewFromBytes([]byte{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0, 1}

	actual := make([]byte, 2)
	n, err := b.Read(actual)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 but found %v", n)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestBitSetBuffer_WriteBits(t *testing.T) {
	b := BitSetBuffer{}
	expectedBits := []bool{true, false, true}
	n, err := b.WriteBits(expectedBits)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(expectedBits) {
		t.Fatalf("expected %v but found %v", len(expectedBits), n)
	}

	actual := b.Set
	if !reflect.DeepEqual(actual, expectedBits) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expectedBits, actual)
	}

	b.ResetToStart()
	bytes := make([]byte, 2) //1 more than we should get
	n, err = b.Read(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected %v but found %v", 1, n)
	}
	bytes = bytes[:1]
	expectedBytes := []byte{5}
	if !reflect.DeepEqual(expectedBytes, bytes) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expectedBytes, bytes)
	}
}

func TestBitSetBuffer_ReadBits(t *testing.T) {
	expected := []bool{true, false, true}
	b, err := NewFromBits(expected)
	if err != nil {
		t.Fatal(err)
	}
	actual := make([]bool, 4)
	n, err := b.ReadBits(actual)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("expected 3 but found %v", n)
	}
	actual = actual[:3]
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestBitSetBuffer_Read2(t *testing.T) {
	b := BitSetBuffer{}
	n, err := b.Read(nil)
	if err == nil {
		t.Fatalf("expected and error but found none")
	}
	if n != 0 {
		t.Fatalf("expected to find 0 but found %v", n)
	}
}

func TestBitSetBuffer_ReadBits2(t *testing.T) {
	b := BitSetBuffer{}
	n, err := b.ReadBits(nil)
	if err == nil {
		t.Fatalf("expected and error but found none")
	}
	if n != 0 {
		t.Fatalf("expected to find 0 but found %v", n)
	}
}

func TestBitSetBuffer_Write(t *testing.T) {
	b, err := NewFromBits([]bool{false, false, false, false})
	if err != nil {
		t.Fatal(err)
	}

	b.ResetToEnd()

	n, err := b.Write([]byte{0xff})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 but found %v", n)
	}

	expectedBits := []bool{false, false, false, false, true, true, true, true, true, true, true, true}
	if !reflect.DeepEqual(b.Set, expectedBits) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expectedBits, b.Set)
	}
	expecteBytes := []byte{0xf0, 0x0f}
	if !reflect.DeepEqual(b.Bytes(), expecteBytes) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expecteBytes, b.Bytes())
	}

}
