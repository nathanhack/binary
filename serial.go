package binary

import (
	"encoding/binary"
	"fmt"
	"github.com/nathanhack/binary/internal"
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
	sizeMap := map[string]int{}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		err := enc(sf.Name, sf.Type, v.Field(i), sf.Tag, buf, sizeMap)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func enc(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf internal.BitSetWriter, sizeMap map[string]int) error {
	endianness, err := getEndianness(tag)
	if err != nil {
		return fmt.Errorf("%v: %v", fieldName, err)
	}

	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			val := reflect.New(t.Elem())
			return enc(fieldName, t.Elem(), val.Elem(), tag, buf, sizeMap)
		} else {
			return enc(fieldName, t.Elem(), v.Elem(), tag, buf, sizeMap)
		}
	case reflect.Struct:
		m := map[string]int{}
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
		//now we make empty items! to fill up to the size
		for i := uint64(0); i < blanks; i++ {
			item := reflect.New(t.Elem())
			if err := enc("", t.Elem(), item.Elem(), tag, buf, sizeMap); err != nil {
				return err
			}
		}
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
	case reflect.Bool:
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			tmp := uint64(0)
			if v.Bool() {
				tmp = 1
			}
			err := writeBits(buf, bitSize, endianness, tmp)
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		} else {
			err := binary.Write(buf, endianness, v.Bool())
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		}
	case reflect.Uint8:
		sizeMap[fieldName] = int(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return err
		}
		if hasBits {
			err := writeBits(buf, bitSize, endianness, v.Uint())
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		} else {
			err := binary.Write(buf, endianness, uint8(v.Uint()))
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		}
	case reflect.Uint16:
		sizeMap[fieldName] = int(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 16)
		if err != nil {
			return err
		}
		if hasBits {
			err := writeBits(buf, bitSize, endianness, v.Uint())
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		} else {
			err := binary.Write(buf, endianness, uint16(v.Uint()))
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		}
	case reflect.Uint32:
		sizeMap[fieldName] = int(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 32)
		if err != nil {
			return err
		}
		if hasBits {
			err := writeBits(buf, bitSize, endianness, v.Uint())
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		} else {
			err := binary.Write(buf, endianness, uint32(v.Uint()))
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		}
	case reflect.Uint64:
		sizeMap[fieldName] = int(v.Uint())
		bitSize, hasBits, err := getBits(tag, sizeMap, 64)
		if err != nil {
			return err
		}
		if hasBits {
			err := writeBits(buf, bitSize, endianness, v.Uint())
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		} else {
			err := binary.Write(buf, endianness, v.Uint())
			if err != nil {
				return fmt.Errorf("%v : %v", fieldName, err)
			}
		}
	case reflect.Int8:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int16")
		}
		sizeMap[fieldName] = int(v.Int())
		err := binary.Write(buf, endianness, int8(v.Int()))
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Int16:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int16")
		}
		sizeMap[fieldName] = int(v.Int())
		err := binary.Write(buf, endianness, int16(v.Int()))
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Int32:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int32")
		}
		sizeMap[fieldName] = int(v.Int())
		err := binary.Write(buf, endianness, int32(v.Int()))
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Int64:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on int64")
		}
		sizeMap[fieldName] = int(v.Int())
		err := binary.Write(buf, endianness, v.Int())
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Float32:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on float32")
		}
		err := binary.Write(buf, endianness, float32(v.Float()))
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Float64:
		_, has := tag.Lookup("bits")
		if has {
			return fmt.Errorf("bits not supported on float64")
		}
		err := binary.Write(buf, endianness, v.Float())
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	default:
		return fmt.Errorf("%v not supported", t)
	}

	return nil
}

func writeBits(buf internal.BitSetWriter, numOfBits int, endianness binary.ByteOrder, value uint64) error {
	bits := make([]bool, numOfBits)
	for i := 0; i < numOfBits; i++ {
		bits[i] = value&(1<<i) > 0
	}

	if endianness == binary.LittleEndian {
		n, err := buf.WriteBits(bits)
		if err != nil {
			return err
		}
		if n != numOfBits {
			return fmt.Errorf("only %v of %v bits written", n, numOfBits)
		}
		return nil
	}

	// else endianness == binary.BigEndian
	stack := make([][]bool, 0)
	for start := 0; start < numOfBits; start += 8 {
		end := start + 8
		if end > numOfBits {
			end = numOfBits
		}
		stack = append(stack, bits[start:end])
	}

	for i := len(stack)/2 - 1; i >= 0; i-- {
		opp := len(stack) - 1 - i
		stack[i], stack[opp] = stack[opp], stack[i]
	}
	for _, item := range stack {
		n, err := buf.WriteBits(item)
		if err != nil {
			return err
		}
		if n != len(item) {
			return fmt.Errorf("only %v of %v bits written", n, numOfBits)
		}
	}

	return nil
}

func readBits(buf internal.BitSetReader, numOfBits int, endianness binary.ByteOrder) (uint64, error) {
	bits := make([]bool, numOfBits)
	n, err := buf.ReadBits(bits)
	if err != nil {
		return 0, err
	}
	if n != numOfBits {
		return 0, fmt.Errorf("only %v of %v bits read", n, numOfBits)
	}

	if endianness == binary.LittleEndian {
		value := uint64(0)
		for i := 0; i < numOfBits; i++ {
			if bits[i] {
				value += 1 << i
			}
		}
		return value, nil
	}

	//we need to undo the byte swapping for
	var value uint64
	byteIndex := uint64(0)
	for end := numOfBits; end >= 0; end -= 8 {
		start := end - 8
		if start < 0 {
			start = 0
		}
		bytes := bits[start:end]
		for i := 0; i < len(bytes); i++ {
			if bytes[i] {
				value += (1 << i) << (8 * byteIndex)
			}
		}
		byteIndex++
	}

	return value, nil
}

func getEndianness(tag reflect.StructTag) (binary.ByteOrder, error) {
	value, ok := tag.Lookup("endian")
	if !ok {
		return binary.LittleEndian, nil
	}
	switch value {
	case "little":
		return binary.LittleEndian, nil
	case "big":
		return binary.BigEndian, nil
	}

	return binary.LittleEndian, fmt.Errorf("unsupported endian value: %v", value)
}

func getBits(tag reflect.StructTag, sizeMap map[string]int, limit uint64) (int, bool, error) {
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
			return 0, true, fmt.Errorf("value of %v is %v,to be used for size it must be positive for bits", s, i)
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
		err := decode(sf.Name, sf.Type, vf, sf.Tag, buf, sizeMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func decode(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *internal.BitSetBuffer, sizeMap map[string]int) error {
	endianness, err := getEndianness(tag)
	if err != nil {
		return fmt.Errorf("%v: %v", fieldName, err)
	}

	switch t.Kind() {
	case reflect.Ptr:
		val := reflect.New(t.Elem())
		err := decode(fieldName, t.Elem(), val.Elem(), tag, buf, sizeMap)
		if err != nil {
			return err
		}
		v.Set(val)
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

		if all {
			sb := strings.Builder{}

			for {
				bs := make([]byte, suint)
				n, err := buf.Read(bs)
				if err != nil {
					return fmt.Errorf("%v: %v", fieldName, err)
				}
				if n != len(bs) {
					sb.Write(bs[:n])
					break
				}
				sb.Write(bs)
			}

			v.SetString(sb.String())
		} else {
			bs := make([]byte, suint)
			err := binary.Read(buf, endianness, bs)
			if err != nil {
				return fmt.Errorf("%v: %v", fieldName, err)
			}
			v.SetString(string(bs))
		}
	case reflect.Bool:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x bool
		if hasBits {
			tmp, err := readBits(buf, numOfBits, endianness)
			if err != nil {
				return fmt.Errorf("%v: %v", fieldName, err)
			}
			x = tmp > 0
		} else {
			if err := binary.Read(buf, endianness, &x); err != nil {
				return fmt.Errorf("expected to read bool from %v: %v", fieldName, err)
			}
		}

		v.SetBool(x)
	case reflect.Uint8:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 8)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint8
		if hasBits {
			tmp, err := readBits(buf, numOfBits, endianness)
			if err != nil {
				return fmt.Errorf("%v: %v", fieldName, err)
			}
			x = uint8(tmp)
		} else {
			if err := binary.Read(buf, endianness, &x); err != nil {
				return fmt.Errorf("expected to read uint8 from %v: %v", fieldName, err)
			}
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint16:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 16)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint16
		if hasBits {
			tmp, err := readBits(buf, numOfBits, endianness)
			if err != nil {
				return fmt.Errorf("%v: %v", fieldName, err)
			}
			x = uint16(tmp)
		} else {
			if err := binary.Read(buf, endianness, &x); err != nil {
				return fmt.Errorf("expected to read uint16 from %v: %v", fieldName, err)
			}
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint32:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 32)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint32
		if hasBits {
			tmp, err := readBits(buf, numOfBits, endianness)
			if err != nil {
				return fmt.Errorf("%v: %v", fieldName, err)
			}
			x = uint32(tmp)
		} else {
			if err := binary.Read(buf, endianness, &x); err != nil {
				return fmt.Errorf("expected to read uint32 from %v: %v", fieldName, err)
			}
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint64:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 64)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint64
		if hasBits {
			x, err = readBits(buf, numOfBits, endianness)
			if err != nil {
				return fmt.Errorf("%v: %v", fieldName, err)
			}
		} else {
			if err := binary.Read(buf, endianness, &x); err != nil {
				return fmt.Errorf("expected to read uint64 from %v: %v", fieldName, err)
			}
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(x)
	case reflect.Int8:
		var x int8
		if err := binary.Read(buf, endianness, &x); err != nil {
			return fmt.Errorf("expected to read int8 from %v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int16:
		var x int16
		if err := binary.Read(buf, endianness, &x); err != nil {
			return fmt.Errorf("expected to read int16 from %v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int32:
		var x int32
		if err := binary.Read(buf, endianness, &x); err != nil {
			return fmt.Errorf("expected to read int32 from %v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int64:
		var x int64
		if err := binary.Read(buf, endianness, &x); err != nil {
			return fmt.Errorf("expected to read int64 from %v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(x)
	case reflect.Float32:
		var x float32
		if err := binary.Read(buf, endianness, &x); err != nil {
			return fmt.Errorf("expected to read float32 from %v: %v", fieldName, err)
		}

		v.SetFloat(float64(x))
	case reflect.Float64:
		var x float64
		if err := binary.Read(buf, endianness, &x); err != nil {
			return fmt.Errorf("expected to read float64 from %v: %v", fieldName, err)
		}

		v.SetFloat(x)
	default:
		return fmt.Errorf("%v not supported", t)
	}

	return nil
}
