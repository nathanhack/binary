package binary

import (
	"encoding/binary"
	"fmt"
	bits "github.com/nathanhack/bitsetbuffer"
	"reflect"
	"strconv"
	"strings"
)

type BitsMarshaler interface {
	MarshalBits() (data *bits.BitSetBuffer, err error)
}

type BitsUnmarshaler interface {
	UnmarshalBits(data *bits.BitSetBuffer) error
}

type EncDecOption interface {
	Type() reflect.Type
	EncoderFunc() func(fieldName string, v reflect.Value, tag reflect.StructTag, buf bits.BitSetWriter, sizeMap map[string]int, options ...EncDecOption) error
	DecoderFunc() func(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *bits.BitSetBuffer, sizeMap map[string]int, options ...EncDecOption) error
}

type StructEncDec struct {
	StructType reflect.Type
	Encoder    func(fieldName string, v reflect.Value, tag reflect.StructTag, buf bits.BitSetWriter, sizeMap map[string]int, options ...EncDecOption) error
	Decoder    func(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *bits.BitSetBuffer, sizeMap map[string]int, options ...EncDecOption) error
}

func (s *StructEncDec) Type() reflect.Type {
	return s.StructType
}

func (s *StructEncDec) EncoderFunc() func(fieldName string, v reflect.Value, tag reflect.StructTag, buf bits.BitSetWriter, sizeMap map[string]int, options ...EncDecOption) error {
	return s.Encoder
}

func (s *StructEncDec) DecoderFunc() func(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *bits.BitSetBuffer, sizeMap map[string]int, options ...EncDecOption) error {
	return s.Decoder
}

type InterfaceEncDec struct {
	InterfaceType reflect.Type
	Encoder       func(fieldName string, v reflect.Value, tag reflect.StructTag, buf bits.BitSetWriter, sizeMap map[string]int, options ...EncDecOption) error
	Decoder       func(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *bits.BitSetBuffer, sizeMap map[string]int, options ...EncDecOption) error
}

func (i *InterfaceEncDec) Type() reflect.Type {
	return i.InterfaceType
}

func (i *InterfaceEncDec) EncoderFunc() func(fieldName string, v reflect.Value, tag reflect.StructTag, buf bits.BitSetWriter, sizeMap map[string]int, options ...EncDecOption) error {
	return i.Encoder
}

func (i *InterfaceEncDec) DecoderFunc() func(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *bits.BitSetBuffer, sizeMap map[string]int, options ...EncDecOption) error {
	return i.Decoder
}

func validateOptions(options ...EncDecOption) error {
	for _, item := range options {
		if item.Type() == nil || (item.Type().Kind() != reflect.Struct && item.Type().Kind() != reflect.Interface) {
			return fmt.Errorf("Type() must not be nil and be either a struct or interface")
		}
		if item.EncoderFunc() == nil {
			return fmt.Errorf("EncoderFunc() must not be nil")
		}
		if item.DecoderFunc() == nil {
			return fmt.Errorf("DecoderFunc() must not be nil")
		}
	}
	return nil
}

//Encode is the main function to call to encode structs. To add special encoding use BitsMarshaler.
//  InterfaceEncDec options are available to be passed in to support Interfaces types.
//  StructEncDec options are also a way to change the behaviour of struct encoding for structs that do/can not implement
//  BitsMarshaler.
func Encode(st interface{}, options ...EncDecOption) ([]byte, error) {
	buf, err := EncodeToBits(st, options...)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodeToBits(st interface{}, options ...EncDecOption) (*bits.BitSetBuffer, error) {
	if st == nil {
		return nil, fmt.Errorf("nil pointer not alowed")
	}

	if err := validateOptions(options...); err != nil {
		return nil, err
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

	//check it we have a BitMarshaler
	buf := &bits.BitSetBuffer{}
	processed, err := encMarshaler(v, buf)
	if err != nil {
		return nil, err
	}

	if processed {
		return buf, nil
	}

	//so we didn't have a BitMarshaler so we'll
	// work on the Struct Options
	sizeMap := map[string]int{}
	processed, err = encStructSpecial("", v, "", buf, sizeMap, options...)
	if err != nil {
		return nil, err
	}

	if processed {
		return buf, nil
	}

	//lastly it's just a plain struct so we get to work on the fields
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if _, has := sf.Tag.Lookup("omit"); has {
			continue
		}
		err = EncodeField(sf.Name, sf.Type, v.Field(i), sf.Tag, buf, sizeMap, options...)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func encMarshaler(v reflect.Value, buf bits.BitSetWriter) (bool, error) {
	modelType := reflect.TypeOf((*BitsMarshaler)(nil)).Elem()
	t := v.Type()
	var marshaler BitsMarshaler
	if t.Implements(modelType) {
		marshaler = v.Interface().(BitsMarshaler)
	} else if reflect.PtrTo(t).Implements(modelType) && v.CanAddr() {
		marshaler = v.Addr().Interface().(BitsMarshaler)
	} else {
		return false, nil
	}

	b, err := marshaler.MarshalBits()
	if err != nil {
		return false, err
	}

	n, err := buf.WriteBits(b.Set)
	if err != nil {
		return false, err
	}

	if n != len(b.Set) {
		return false, fmt.Errorf("wrote %v expected %v", n, len(b.Set))
	}
	return true, nil
}

func encStructSpecial(fieldName string, v reflect.Value, tag reflect.StructTag, buf bits.BitSetWriter, sizeMap map[string]int, options ...EncDecOption) (bool, error) {
	for _, enc := range options {
		if enc.Type() == v.Type() {
			err := enc.EncoderFunc()(fieldName, v, tag, buf, sizeMap, options...)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

//EncodeField should be only if it's part of one of the encode function in one of the options (StructEncDec or InterfaceEncDec).  When
// called on a field it will do correct encoding. Be careful when calling this function in the options as to avoid recursive explosion.
func EncodeField(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf bits.BitSetWriter, sizeMap map[string]int, options ...EncDecOption) error {
	//we check for the BitsMarshaler
	processed, err := encMarshaler(v, buf)
	if err != nil {
		return err
	}

	if processed {
		return nil
	}

	endianness, err := getEndianness(tag)
	if err != nil {
		return fmt.Errorf("%v: %v", fieldName, err)
	}

	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			val := reflect.New(t.Elem())
			return EncodeField(fieldName, t.Elem(), val.Elem(), tag, buf, sizeMap, options...)
		} else {
			return EncodeField(fieldName, t.Elem(), v.Elem(), tag, buf, sizeMap, options...)
		}
	case reflect.Interface:
		for _, enc := range options {
			if enc.Type() == v.Type() {
				err := enc.EncoderFunc()(fieldName, v, tag, buf, sizeMap, options...)
				if err != nil {
					return err
				}
				return nil
			}
		}
		return fmt.Errorf("interface:%v was not found: interface not supported", t.Name())
	case reflect.Struct:
		processed, err := encStructSpecial(fieldName, v, tag, buf, sizeMap, options...)
		if err != nil {
			return err
		}

		if processed {
			return nil
		}

		m := map[string]int{}
		for k, v := range sizeMap {
			m[k] = v
		}

		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			if _, has := sf.Tag.Lookup("omit"); has {
				continue
			}
			err = EncodeField(sf.Name, sf.Type, v.Field(i), sf.Tag, buf, m, options...)
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := EncodeField("", item.Type(), item, tag, buf, sizeMap, options...); err != nil {
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
			if err := EncodeField("", item.Type(), item, tag, buf, sizeMap, options...); err != nil {
				return err
			}
		}
		//now we make empty items! to fill up to the size
		for i := uint64(0); i < blanks; i++ {
			item := reflect.New(t.Elem())
			if err := EncodeField("", t.Elem(), item.Elem(), tag, buf, sizeMap, options...); err != nil {
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
		bitSize, hasBits, err := getBits(tag, sizeMap, 8, 0)
		if err != nil {
			return err
		}
		if hasBits {
			tmp := uint64(0)
			if v.Bool() {
				tmp = 1
			}
			err = bits.WriteUint(buf, bitSize, endianness, tmp)
		} else {
			err = binary.Write(buf, endianness, v.Bool())
		}
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Uint8:
		sizeMap[fieldName] = int(v.Uint())

		bitSize, hasBits, err := getBits(tag, sizeMap, 8, 0)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteUint(buf, bitSize, endianness, v.Uint())
		} else {
			err = binary.Write(buf, endianness, uint8(v.Uint()))
		}
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Uint16:
		sizeMap[fieldName] = int(v.Uint())

		bitSize, hasBits, err := getBits(tag, sizeMap, 16, 0)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteUint(buf, bitSize, endianness, v.Uint())
		} else {
			err = binary.Write(buf, endianness, uint16(v.Uint()))
		}
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Uint32:
		sizeMap[fieldName] = int(v.Uint())

		bitSize, hasBits, err := getBits(tag, sizeMap, 32, 0)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteUint(buf, bitSize, endianness, v.Uint())
		} else {
			err = binary.Write(buf, endianness, uint32(v.Uint()))
		}
		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Uint64:
		sizeMap[fieldName] = int(v.Uint())

		bitSize, hasBits, err := getBits(tag, sizeMap, 64, 0)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteUint(buf, bitSize, endianness, v.Uint())

		} else {
			err = binary.Write(buf, endianness, v.Uint())
		}

		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Int8:
		sizeMap[fieldName] = int(v.Int())

		bitSize, hasBits, err := getBits(tag, sizeMap, 8, 2)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteInt(buf, bitSize, endianness, v.Int())
		} else {
			err = binary.Write(buf, endianness, int8(v.Int()))
		}

		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Int16:
		sizeMap[fieldName] = int(v.Int())

		bitSize, hasBits, err := getBits(tag, sizeMap, 16, 2)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteInt(buf, bitSize, endianness, v.Int())
		} else {
			err = binary.Write(buf, endianness, int16(v.Int()))
		}

		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Int32:
		sizeMap[fieldName] = int(v.Int())

		bitSize, hasBits, err := getBits(tag, sizeMap, 32, 2)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteInt(buf, bitSize, endianness, v.Int())
		} else {
			err = binary.Write(buf, endianness, int32(v.Int()))
		}

		if err != nil {
			return fmt.Errorf("%v : %v", fieldName, err)
		}
	case reflect.Int64:
		sizeMap[fieldName] = int(v.Int())

		bitSize, hasBits, err := getBits(tag, sizeMap, 64, 2)
		if err != nil {
			return err
		}
		if hasBits {
			err = bits.WriteInt(buf, bitSize, endianness, v.Int())
		} else {
			err = binary.Write(buf, endianness, v.Int())
		}

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

func getBits(tag reflect.StructTag, sizeMap map[string]int, maxLimit, minLimit uint64) (int, bool, error) {
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

	if value > maxLimit {
		return 0, true, fmt.Errorf("bits value was larger than maxLimit")
	}
	if value < minLimit {
		return 0, true, fmt.Errorf("bits value was smaller than minLimit")
	}

	return int(value), true, nil
}

//Decode is the main function to call to decode struct. To add special decoding use BitsUnmarshaler.
//  InterfaceEncDec options are available to be passed in to support Interfaces types.
//  StructEncDec options are also a way to change the behaviour of struct decoding for structs that do/can not implement
//  BitsUnmarshaler.
func Decode(data []byte, value interface{}, options ...EncDecOption) error {
	if data == nil || value == nil {
		return fmt.Errorf("nil parameters not allowed")
	}

	buf, err := bits.NewFromBytes(data)
	if err != nil {
		return err
	}

	return DecodeToBits(buf, value, options...)
}

func DecodeToBits(buf *bits.BitSetBuffer, value interface{}, options ...EncDecOption) error {
	if buf == nil || value == nil {
		return fmt.Errorf("nil parameters not allowed")
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	//we require the struct coming in to be at least pointer to a struct
	// so we can populate it
	if t.Kind() != reflect.Ptr {
		panic("value expected to be a pointer to a structure")
	}

	//we unwrap until we get to the struct
loop:
	for {
		switch t.Kind() {
		case reflect.Ptr:
			t = t.Elem()
			v = v.Elem()
		case reflect.Struct:
			break loop
		default:
			return fmt.Errorf("invalid value")
		}
	}

	//first we check if it's a BitsUnmarshaler
	processed, err := decUnmarshaler(v, buf)
	if err != nil {
		return err
	}

	if processed {
		return nil
	}

	//next we check the options
	sizeMap := map[string]int{}
	processed, err = decStructSpecial("", t, v, "", buf, sizeMap, options...)
	if err != nil {
		return err
	}

	if processed {
		return nil
	}

	//for the last case we take the struct and unmarshal all the fields
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		vf := v.Field(i)
		if _, has := sf.Tag.Lookup("omit"); has {
			continue
		}
		err := DecodeField(sf.Name, sf.Type, vf, sf.Tag, buf, sizeMap, options...)
		if err != nil {
			return err
		}
	}

	return nil
}

func decUnmarshaler(v reflect.Value, buf *bits.BitSetBuffer) (bool, error) {
	modelType := reflect.TypeOf((*BitsUnmarshaler)(nil)).Elem()
	t := v.Type()
	var unmarshaler BitsUnmarshaler
	if t.Implements(modelType) {
		unmarshaler = v.Interface().(BitsUnmarshaler)
	} else if reflect.PtrTo(t).Implements(modelType) {
		unmarshaler = v.Addr().Interface().(BitsUnmarshaler)
	} else {
		return false, nil
	}

	err := unmarshaler.UnmarshalBits(buf)
	if err != nil {
		return false, err
	}
	return true, nil
}

func decStructSpecial(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *bits.BitSetBuffer, sizeMap map[string]int, options ...EncDecOption) (bool, error) {
	for _, dec := range options {
		if dec.Type() == v.Type() {
			err := dec.DecoderFunc()(fieldName, t, v, tag, buf, sizeMap, options...)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

//DecodeField should be only if it's part of one of the decode function in one of the options (StructEncDec or InterfaceEncDec).  When
// called on a field it will do correct decoding. Be careful when calling this function in the options as to avoid recursive explosion.
func DecodeField(fieldName string, t reflect.Type, v reflect.Value, tag reflect.StructTag, buf *bits.BitSetBuffer, sizeMap map[string]int, options ...EncDecOption) error {
	processed, err := decUnmarshaler(v, buf)
	if err != nil {
		return err
	}

	if processed {
		return nil
	}

	endianness, err := getEndianness(tag)
	if err != nil {
		return fmt.Errorf("%v: %v", fieldName, err)
	}

	switch t.Kind() {
	case reflect.Ptr:
		val := reflect.New(t.Elem())
		err := DecodeField(fieldName, t.Elem(), val.Elem(), tag, buf, sizeMap, options...)
		if err != nil {
			return err
		}
		v.Set(val)
	case reflect.Interface:
		for _, enc := range options {
			if enc.Type() == v.Type() {
				err := enc.DecoderFunc()(fieldName, t, v, tag, buf, sizeMap, options...)
				if err != nil {
					return err
				}
				return nil
			}
		}
		return fmt.Errorf("interface:%v was not found: interface not supported", t.Name())
	case reflect.Struct:
		m := make(map[string]int)
		for k, v := range sizeMap {
			m[k] = v
		}
		processed, err := decStructSpecial(fieldName, t, v, tag, buf, m, options...)
		if err != nil {
			return err
		}

		if processed {
			return nil
		}

		for i := 0; i < v.NumField(); i++ {
			sf := t.Field(i)
			vf := v.Field(i)
			if _, has := sf.Tag.Lookup("omit"); has {
				continue
			}
			err := DecodeField(sf.Name, sf.Type, vf, sf.Tag, buf, m, options...)
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := DecodeField("", item.Type(), item, tag, buf, sizeMap, options...); err != nil {
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
			if err := DecodeField("", item.Elem().Type(), item.Elem(), tag, buf, sizeMap, options...); err != nil {
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
		numOfBits, hasBits, err := getBits(tag, sizeMap, 8, 0)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x bool
		if hasBits {
			tmp, err := bits.ReadUint(buf, numOfBits, endianness)
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
		numOfBits, hasBits, err := getBits(tag, sizeMap, 8, 0)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint8
		if hasBits {
			var tmp uint64
			tmp, err = bits.ReadUint(buf, numOfBits, endianness)
			x = uint8(tmp)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint16:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 16, 0)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint16
		if hasBits {
			var tmp uint64
			tmp, err = bits.ReadUint(buf, numOfBits, endianness)
			x = uint16(tmp)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint32:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 32, 0)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint32
		if hasBits {
			var tmp uint64
			tmp, err = bits.ReadUint(buf, numOfBits, endianness)
			x = uint32(tmp)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(uint64(x))
	case reflect.Uint64:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 64, 0)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x uint64
		if hasBits {
			x, err = bits.ReadUint(buf, numOfBits, endianness)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetUint(x)
	case reflect.Int8:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 8, 2)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x int8
		if hasBits {
			var tmp int64
			tmp, err = bits.ReadInt(buf, numOfBits, endianness)
			x = int8(tmp)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int16:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 16, 2)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x int16
		if hasBits {
			var tmp int64
			tmp, err = bits.ReadInt(buf, numOfBits, endianness)
			x = int16(tmp)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int32:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 32, 2)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x int32
		if hasBits {
			var tmp int64
			tmp, err = bits.ReadInt(buf, numOfBits, endianness)
			x = int32(tmp)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(int64(x))
	case reflect.Int64:
		numOfBits, hasBits, err := getBits(tag, sizeMap, 64, 2)
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		var x int64
		if hasBits {
			x, err = bits.ReadInt(buf, numOfBits, endianness)
		} else {
			err = binary.Read(buf, endianness, &x)
		}
		if err != nil {
			return fmt.Errorf("%v: %v", fieldName, err)
		}

		sizeMap[fieldName] = int(x)
		v.SetInt(x)
	case reflect.Float32:
		_, hasBits, _ := getBits(tag, map[string]int{}, 32, 0)
		if hasBits {
			return fmt.Errorf("bits not supported with float32: %v", fieldName)
		}

		var x float32
		if err := binary.Read(buf, endianness, &x); err != nil {
			return fmt.Errorf("expected to read float32 from %v: %v", fieldName, err)
		}

		v.SetFloat(float64(x))
	case reflect.Float64:
		_, hasBits, _ := getBits(tag, map[string]int{}, 64, 0)
		if hasBits {
			return fmt.Errorf("bits not supported with float64: %v", fieldName)
		}

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

//SizeOf returns the minimum number of bytes needed to serialize the structure
func SizeOf(v interface{}, options ...EncDecOption) int {
	bs, err := Encode(v, options...)
	if err != nil {
		panic(err)
	}

	return len(bs)
}
