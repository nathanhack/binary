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
		return enc(fieldName, t.Elem(), v.Elem(), tag, buf, sizeMap)
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

func Decode(data []byte, v interface{}) error {
	if data == nil || v == nil {
		return fmt.Errorf("nil parameters")
	}

	buf, err := internal.NewFromBytes(data)
	if err != nil {
		return err
	}
	tOf := reflect.TypeOf(v)
	vOf := reflect.ValueOf(v)
	if vOf.Kind() != reflect.Ptr {
		return fmt.Errorf("expected v to be a pointer to a struct")
	}
	tOf = tOf.Elem()
	vOf = vOf.Elem()
	if vOf.Kind() != reflect.Struct {
		return fmt.Errorf("expected v to be a pointer to a struct")
	}
	sizeMap := map[string]int{}
	for i := 0; i < vOf.NumField(); i++ {
		sf := tOf.Field(i)
		vf := vOf.Field(i)
		fmt.Printf("Start : %v || %v\n", sf.Type, vf)
		err := decode(sf.Name, sf.Type, vf, sf.Tag, buf, sizeMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func decode(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *internal.BitSetBuffer, sizeMap map[string]int) error {
	endianness := getEndianness(tag)
	switch t.Kind() {
	case reflect.Ptr:
		val := reflect.New(t.Elem())
		err := decode(fieldName, t.Elem(), val.Elem(), tag, buf, sizeMap)
		if err != nil {
			return err
		}
		v.Set(val)
		return nil
	case reflect.Struct:
		sizeMap := map[string]int{}
		for i := 0; i < v.NumField(); i++ {
			sf := t.Field(i)
			vf := v.Field(i)
			err := decode(sf.Name, sf.Type, vf, sf.Tag, buf, sizeMap)
			if err != nil {
				return err
			}

		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := decode("", item.Type(), item, tag, buf, sizeMap); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		all := true
		suint := 0
		if s, ok := tag.Lookup("size"); ok {
			tmp, err := strconv.ParseUint(s, 10, 64)
			suint = int(tmp)
			if err != nil {
				i, has := sizeMap[s]
				switch {
				case !has:
					return fmt.Errorf("size must either be a positive number or a field found prior to this field :%v", err)
				case i < 0:
					return fmt.Errorf("value of %v is %v, to be used for size it must be nonnegative", s, i)
				}

				suint = i
			}
			all = false
		}

		reflectionValue := reflect.New(t)
		reflectionValue.Elem().Set(reflect.MakeSlice(t, 0, 10))
		sliceValuePtr := reflect.ValueOf(reflectionValue.Interface()).Elem()
		for i := 0; i < suint || (all && !buf.PosAtEnd()); i++ {
			item := reflect.New(t.Elem())
			if err := decode("", item.Elem().Type(), item.Elem(), tag, buf, sizeMap); err != nil {
				return err
			}

			sliceValuePtr.Set(reflect.Append(sliceValuePtr, item.Elem()))
		}

		v.Set(sliceValuePtr)
	case reflect.String:
		all := true
		suint := uint64(0)
		if s, ok := tag.Lookup("strlen"); ok {
			var err error
			suint, err = strconv.ParseUint(s, 10, 64)
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
			all = false
		}
		s := ""
		for i := 0; i < int(suint) || (all); i++ {
			b, ok := buf.ReadByte()
			if !ok {
				if i == 0 {
					return fmt.Errorf("error reading value for %v", fieldName)
				}
				break
			}
			s += string(b)
		}

		v.SetString(s)

	case reflect.Uint8:
		x, ok := buf.ReadUint8()
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint16:
		x, ok := buf.ReadUint16(endianness)
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint32:
		x, ok := buf.ReadUint32(endianness)
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint64:
		x, ok := buf.ReadUint64(endianness)
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetUint(x)
	case reflect.Int8:
		x, ok := buf.ReadInt8()
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int16:
		x, ok := buf.ReadInt16(endianness)
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int32:
		x, ok := buf.ReadInt32(endianness)
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int64:
		x, ok := buf.ReadInt64(endianness)
		if !ok {
			return fmt.Errorf("expected to read uint64 from %v", fieldName)
		}
		sizeMap[fieldName] = int(x)
		v.SetInt(x)
	default:
		return fmt.Errorf("%v not supported", t)
	}

	return nil
}
