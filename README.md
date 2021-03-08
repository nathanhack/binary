# binary
A simple binary serialization of structs to byte aligned binary output.


## Getting Started

#### Import
Import the library `import "github.com/nathanhack/binary"`.


#### Annotate a struct
Create a type and annotate it with tags  that will be used during serialization.
```
type Thing struct {
	V1 uint16   `endian:"little"`
	V2 []byte   `size:"4"`
}
```
In the above example the struct has two fields. Fields must be public for serialization to work.  The `V1` field is a `uint16` and when it is serialized we want it to in little endianness so we add the ``` `endian:"little"` ``` tag (see below for more information). The second field is `V2` and we expect it to only have a size(length) of 4 (see below for more information). 

#### Encode to bytes
Now that the struct is annotated we can now serialize it. We'll serialize empty version of the struct above.
```
bytes, err := Encode(&Thing{})
if err != nil {
	...
}
``` 

The `bytes` would be `[0 0 0 0 0 0]`. Note that size made enforces how many items are serialized even if they don't exist.


### Decode from bytes
Using the above annotated `Thing` struct we'll decode `[0 0 0 0 0 0]`.

```
bytes := []bytes{0,0,0,0,0,0}

var thing Thing
err := Decode(bytes, &thing)
if err != nil {
	...
}
```

What you'd expect to find is `thing` would be equivalent to `expected`:
```
expected := Thing{
		V1: 0,
		V2: []byte{0, 0, 0, 0},
	}
```

## Tags
Tags are annotations to specify specific handling for each field.  All tags are case sensitive. 

#### endian
When a integer value larger than 8bits is tagged with ``` `endian:"little"` ``` or ``` `endian:"big"` ``` (note case does matter).

#### size
When a slice is tagged with ``` `size:"X"` ``` it enforces a size requirement upon the slice.  This means if the slice has a length bigger than `X` then it will only serialize the first `X` items of the slice; additionally, if the slice's length is smaller than `X` then it will serialize zero initialized items until the size requirement is met. Note the `X` can either be a positive integer or an integer field name in the struct that occurs before the current field.

#### strlen
When a string is tagged with ``` `strlen:"X"` ``` it enforces a length requiement upon the string. This means if the string has a length larger than `X` then it will only serialize the first `X` bytes of the string; additionally, if the string's length is smaller than `X` then it will serialize the space char ```   ``` until the size requirement is met. Note the `X` can either be a positive integer or an integer field name in the struct that occurs before the current field.

#### bits
When an unsigned integer or `bool` field is tagged with ``` `bits:"X" ``` then the number will be serialized into `X` number of bits. `X` must be a positive integer value and must be equal to or smaller than the max number of bits for the field (ie. `8` bits for an `uint8`). Note the `X` can either be a positive integer or an integer field name in the struct that occurs before the current field.

Warning when using `bits` care must be taken, as it might differ from what you expect.  The following example shows an example of how bits are managed when the number of bits extends beyond byte boundaries.
```
type Thing struct {
	V1 uint16   `bits:"9" endian:"little"`
}


bytes,_ := Encode(&Thing{0x155}) // 0x155 = 0b101010101
```

`bytes` would contain `[0x55 0x1]` 

```
type Thing struct {
	V1 uint16   `bits:"9" endian:"big"`
}

bytes, err := Encode(&Thing{0x155})  // 0x155 = 0b101010101
```

`bytes` would contain `[171 0]`

At first glance this my look unexpected but when looking at the bits it will all make sense.

Given `0x155` == `101010101` where the most left `1` is the LSB. 

We first take the ``` `endian:"little"` ``` case. First break it into bytes: `10101010` and `1`.In little endian the least significant byte comes first thus one would expect them to be combined in this order ``` [10101010,1] ``` with the resulting byte stream `[0b10101010,0b10000000]` == `[0x55 0x1]`. 

Where as in big endian the most significant byte come first thus one would expect them to be combined in this order [`1`,`10101010`] with the resulting byte stream `[0b11010101,0b00000000]`== `[171 0]`.
  


## Supported field types

`bool`, `uint8`, `uint16`, `uint32`, `uint64`, `int8`, `int16`, `int32`, `int64`, `float32`, `float64`, `struct`, `string`, slices, and arrays 

## Unsupported field types

`map`, `interface{}`, `chan`, `func`, `int` `uint`
