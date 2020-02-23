package binary

import (
	"encoding/binary"
	"fmt"
	"github.com/nathanhack/serialstruct/internal"
	"reflect"
	"strconv"
	"strings"
)

func Encode(st interface{}) ([]byte, error) {
	if st == nil {
		return nil, fmt.Errorf("nil pointer not alowed")
	}
	t := reflect.TypeOf(st)
	v := reflect.ValueOf(st)
loop:
	for {
		switch t.Kind() {
		case reflect.Ptr:
			t = t.Elem()
			v = v.Elem()
		case reflect.Struct:
			break loop
		default:
			return nil, fmt.Errorf("invalid value")
		}
	}

	buf := &internal.BitSetBuffer{}
	sizeMap := map[string]int64{}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		err := enc(sf.Name, sf.Type, v.Field(i), sf.Tag, buf, sizeMap)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func enc(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf internal.BitSetWriter, sizeMap map[string]int64) error {
	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return encEmpty("", t.Elem(), tag, buf, sizeMap)
		}
		return enc("", t.Elem(), v.Elem(), tag, buf, sizeMap)
	case reflect.Struct:
		m := map[string]int64{}
		for k, v := range sizeMap {
			m[k] = v
		}
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			err := enc(sf.Name, sf.Type, v.Field(i), sf.Tag, buf, m)
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := enc("", item.Type(), item, tag, buf, sizeMap); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		itemslen := v.Len()
		blanks := uint64(0)
		if s, ok := tag.Lookup("size"); ok {
			suint, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				i, has := sizeMap[s]
				switch {
				case !has:
					return fmt.Errorf("size must either be a positive number or a field found prior to this field :%v", err)
				case i < 0:
					return fmt.Errorf("value of %v is %v,to be used for size it must be nonnegative", s, i)
				}

				suint = uint64(i)
			}
			if uint64(itemslen) > suint {
				itemslen = int(suint)
			} else if uint64(itemslen) < suint {
				blanks = suint - uint64(itemslen)
			}
		}
		for i := 0; i < itemslen; i++ {
			item := v.Index(i)
			if err := enc("", item.Type(), item, tag, buf, sizeMap); err != nil {
				return err
			}
		}
		for i := uint64(0); i < blanks; i++ {
			if err := encEmpty("", t.Elem(), tag, buf, sizeMap); err != nil {
				return err
			}
		}

		return nil
	case reflect.String:
		s := v.String()
		itemslen := len(s)
		blanks := uint64(0)
		if s, ok := tag.Lookup("strlen"); ok {
			suint, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				i, has := sizeMap[s]
				switch {
				case !has:
					return fmt.Errorf("strlen must either be a positive number or a field found prior to this field :%v", err)
				case i < 0:
					return fmt.Errorf("value of %v is %v, to be used for strlen it must be nonnegative", s, i)
				}

				suint = uint64(i)
			}
			if uint64(itemslen) > suint {
				itemslen = int(suint)
			} else if uint64(itemslen) < suint {
				blanks = suint - uint64(itemslen)
			}
		}
		str := s[:itemslen] + strings.Repeat(" ", int(blanks))
		n, err := buf.Write([]byte(str))
		if err != nil {
			return err
		}
		if n != len(str) {
			return fmt.Errorf("writing string value `%v` failed", str)
		}
		return nil
	case reflect.Uint8:
		sizeMap[fieldName] = int64(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, v.Uint())
		} else {
			return binary.Write(buf, getEndianness(tag), uint8(v.Uint()))
		}
	case reflect.Uint16:
		sizeMap[fieldName] = int64(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, v.Uint())
		} else {
			return binary.Write(buf, getEndianness(tag), uint16(v.Uint()))
		}
	case reflect.Uint32:
		sizeMap[fieldName] = int64(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, v.Uint())
		} else {
			return binary.Write(buf, getEndianness(tag), uint32(v.Uint()))
		}
	case reflect.Uint64:
		sizeMap[fieldName] = int64(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, v.Uint())
		} else {
			return binary.Write(buf, getEndianness(tag), v.Uint())
		}
	case reflect.Int8:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int16")
		}
		sizeMap[fieldName] = v.Int()
		return binary.Write(buf, getEndianness(tag), int8(v.Int()))
	case reflect.Int16:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int16")
		}
		sizeMap[fieldName] = v.Int()
		return binary.Write(buf, getEndianness(tag), int16(v.Int()))
	case reflect.Int32:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int32")
		}
		sizeMap[fieldName] = v.Int()
		return binary.Write(buf, getEndianness(tag), int32(v.Int()))
	case reflect.Int64:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int64")
		}
		sizeMap[fieldName] = v.Int()
		return binary.Write(buf, getEndianness(tag), v.Int())
	default:
		return fmt.Errorf("%v not supported", t)
	}

	return nil
}

func writeBits(buf internal.BitSetWriter, numOfBits int, value uint64) {
	bits := make([]bool, numOfBits)
	for i := 0; i < numOfBits; i++ {
		bits[i] = value&(1<<i) > 0
	}
	buf.WriteBits(bits)
}

func encEmpty(fieldName string, t reflect.Type, tag reflect.StructTag, buf internal.BitSetWriter, sizeMap map[string]int64) error {
	switch t.Kind() {
	case reflect.Ptr:
		return encEmpty("", t.Elem(), tag, buf, sizeMap)
	case reflect.Struct:
		m := map[string]int64{}
		for k, v := range sizeMap {
			m[k] = v
		}
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			err := encEmpty(sf.Name, sf.Type, sf.Tag, buf, m)
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		for i := 0; i < t.Len(); i++ {
			if err := encEmpty("", t.Elem(), tag, buf, sizeMap); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		blanks := uint64(0)
		if s, ok := tag.Lookup("size"); ok {
			var err error
			blanks, err = strconv.ParseUint(s, 10, 64)
			if err != nil {
				return fmt.Errorf("unsigned integer for size not parsable:%v", err)
			}

		}
		for i := uint64(0); i < blanks; i++ {
			if err := encEmpty("", t.Elem(), tag, buf, sizeMap); err != nil {
				return err
			}
		}
		return nil
	case reflect.String:
		blanks := uint64(0)
		if s, ok := tag.Lookup("strlen"); ok {
			var err error
			blanks, err = strconv.ParseUint(s, 10, 64)
			if err != nil {
				return fmt.Errorf("unsigned integer for size not parsable:%v", err)
			}
		}
		str := strings.Repeat(" ", int(blanks))
		n, err := buf.Write([]byte(str))
		if err != nil {
			return err
		}
		if n != len(str) {
			return fmt.Errorf("writing string value `%v` failed", str)
		}
		return nil
	case reflect.Uint8:
		sizeMap[fieldName] = 0
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, 0)
		} else {
			return binary.Write(buf, getEndianness(tag), uint8(0))
		}
	case reflect.Uint16:
		sizeMap[fieldName] = 0
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, 0)
		} else {
			return binary.Write(buf, getEndianness(tag), uint16(0))
		}
	case reflect.Uint32:
		sizeMap[fieldName] = 0
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, 0)
		} else {
			return binary.Write(buf, getEndianness(tag), uint32(0))
		}
	case reflect.Uint64:
		sizeMap[fieldName] = 0
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			writeBits(buf, bitSize, 0)
		} else {
			return binary.Write(buf, getEndianness(tag), uint64(0))
		}
	case reflect.Int8:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int8")
		}
		sizeMap[fieldName] = 0
		return binary.Write(buf, getEndianness(tag), int8(0))
	case reflect.Int16:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int16")
		}
		sizeMap[fieldName] = 0
		return binary.Write(buf, getEndianness(tag), int16(0))
	case reflect.Int32:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int32")
		}
		sizeMap[fieldName] = 0
		return binary.Write(buf, getEndianness(tag), int32(0))
	case reflect.Int64:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int64")
		}
		sizeMap[fieldName] = 0
		return binary.Write(buf, getEndianness(tag), int64(0))
	default:
		return fmt.Errorf("%v not supported", t)
	}

	return nil
}

func getEndianness(tag reflect.StructTag) binary.ByteOrder {
	value, ok := tag.Lookup("endian")
	if !ok {
		return binary.LittleEndian
	}
	switch value {
	case "little":
		return binary.LittleEndian
	case "big":
		return binary.BigEndian
	}

	panic("endian only supports values of `little` or `big`")
}

func getBits(tag reflect.StructTag, sizeMap map[string]int64, limit uint64) (int, bool, error) {
	s, ok := tag.Lookup("bits")
	if !ok {
		return 0, false, nil
	}
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		i, has := sizeMap[s]
		switch {
		case !has:
			return 0, true, fmt.Errorf("bits must either be a positive number or a field found prior to this field :%v", err)
		case i < 0:
			return 0, true, fmt.Errorf("value of %v is %v,to be used for size it must be nonnegative for bits", s, i)
		}

		value = uint64(i)
	}

	if value > limit {
		return 0, true, fmt.Errorf("bits value was larger than type limit")
	}

	return int(value), true, nil
}
