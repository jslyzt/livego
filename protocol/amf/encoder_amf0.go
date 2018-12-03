package amf

import (
	"encoding/binary"
	"io"
	"reflect"
)

// EncodeAmf0 amf0 polymorphic router
func (e *Encoder) EncodeAmf0(w io.Writer, val interface{}) (int, error) {
	if val == nil {
		return e.EncodeAmf0Null(w, true)
	}

	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return e.EncodeAmf0Null(w, true)
	}

	switch v.Kind() {
	case reflect.String:
		str := v.String()
		if len(str) <= Amf0StringMax {
			return e.EncodeAmf0String(w, str, true)
		}
		return e.EncodeAmf0LongString(w, str, true)
	case reflect.Bool:
		return e.EncodeAmf0Boolean(w, v.Bool(), true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return e.EncodeAmf0Number(w, float64(v.Int()), true)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return e.EncodeAmf0Number(w, float64(v.Uint()), true)
	case reflect.Float32, reflect.Float64:
		return e.EncodeAmf0Number(w, float64(v.Float()), true)
	case reflect.Array, reflect.Slice:
		length := v.Len()
		arr := make(Array, length)
		for i := 0; i < length; i++ {
			arr[i] = v.Index(int(i)).Interface()
		}
		return e.EncodeAmf0StrictArray(w, arr, true)
	case reflect.Map:
		obj, ok := val.(Object)
		if ok != true {
			return 0, Error("encode amf0: unable to create object from map")
		}
		return e.EncodeAmf0Object(w, obj, true)
	}

	if _, ok := val.(TypedObject); ok {
		return 0, Error("encode amf0: unsupported type typed object")
	}

	return 0, Error("encode amf0: unsupported type %s", v.Type())
}

// EncodeAmf0Number marker: 1 byte 0x00
// format: 8 byte big endian float64
func (e *Encoder) EncodeAmf0Number(w io.Writer, val float64, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0NumberMarker); err != nil {
			return
		}
		n++
	}

	err = binary.Write(w, binary.BigEndian, &val)
	if err != nil {
		return
	}
	n += 8

	return
}

// EncodeAmf0Boolean marker: 1 byte 0x01
// format: 1 byte, 0x00 = false, 0x01 = true
func (e *Encoder) EncodeAmf0Boolean(w io.Writer, val bool, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0BooleanMarker); err != nil {
			return
		}
		n++
	}

	var m int
	buf := make([]byte, 1)
	if val {
		buf[0] = Amf0BooleanTrue
	} else {
		buf[0] = Amf0BooleanFalse
	}

	m, err = w.Write(buf)
	if err != nil {
		return
	}
	n += m

	return
}

// EncodeAmf0String marker: 1 byte 0x02
// format:
// - 2 byte big endian uint16 header to determine size
// - n (size) byte utf8 string
func (e *Encoder) EncodeAmf0String(w io.Writer, val string, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0StringMarker); err != nil {
			return
		}
		n++
	}

	var m int
	length := uint16(len(val))
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return n, Error("encode amf0: unable to encode string length: %s", err)
	}
	n += 2

	m, err = w.Write([]byte(val))
	if err != nil {
		return n, Error("encode amf0: unable to encode string value: %s", err)
	}
	n += m

	return
}

// EncodeAmf0Object marker: 1 byte 0x03
// format:
// - loop encoded string followed by encoded value
// - terminated with empty string followed by 1 byte 0x09
func (e *Encoder) EncodeAmf0Object(w io.Writer, val Object, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0ObjectMarker); err != nil {
			return
		}
		n++
	}

	var m int
	for k, v := range val {
		m, err = e.EncodeAmf0String(w, k, false)
		if err != nil {
			return n, Error("encode amf0: unable to encode object key: %s", err)
		}
		n += m

		m, err = e.EncodeAmf0(w, v)
		if err != nil {
			return n, Error("encode amf0: unable to encode object value: %s", err)
		}
		n += m
	}

	m, err = e.EncodeAmf0String(w, "", false)
	if err != nil {
		return n, Error("encode amf0: unable to encode object empty string: %s", err)
	}
	n += m

	err = WriteMarker(w, Amf0ObjectEndMarker)
	if err != nil {
		return n, Error("encode amf0: unable to object end marker: %s", err)
	}
	n++
	return
}

// EncodeAmf0Null marker: 1 byte 0x05
// no additional data
func (e *Encoder) EncodeAmf0Null(w io.Writer, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0NullMarker); err != nil {
			return
		}
		n++
	}
	return
}

// EncodeAmf0Undefined marker: 1 byte 0x06
// no additional data
func (e *Encoder) EncodeAmf0Undefined(w io.Writer, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0UndefinedMarker); err != nil {
			return
		}
		n++
	}
	return
}

// EncodeAmf0EcmaArray marker: 1 byte 0x08
// format:
// - 4 byte big endian uint32 with length of associative array
// - normal object format:
//   - loop encoded string followed by encoded value
//   - terminated with empty string followed by 1 byte 0x09
func (e *Encoder) EncodeAmf0EcmaArray(w io.Writer, val Object, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0EcmaArrayMarker); err != nil {
			return
		}
		n++
	}

	var m int
	length := uint32(len(val))
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return n, Error("encode amf0: unable to encode ecma array length: %s", err)
	}
	n += 4

	m, err = e.EncodeAmf0Object(w, val, false)
	if err != nil {
		return n, Error("encode amf0: unable to encode ecma array object: %s", err)
	}
	n += m
	return
}

// EncodeAmf0StrictArray marker: 1 byte 0x0a
// format:
// - 4 byte big endian uint32 to determine length of associative array
// - n (length) encoded values
func (e *Encoder) EncodeAmf0StrictArray(w io.Writer, val Array, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0StrictArrayMarker); err != nil {
			return
		}
		n++
	}

	var m int
	length := uint32(len(val))
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return n, Error("encode amf0: unable to encode strict array length: %s", err)
	}
	n += 4

	for _, v := range val {
		m, err = e.EncodeAmf0(w, v)
		if err != nil {
			return n, Error("encode amf0: unable to encode strict array element: %s", err)
		}
		n += m
	}
	return
}

// EncodeAmf0LongString marker: 1 byte 0x0c
// format:
// - 4 byte big endian uint32 header to determine size
// - n (size) byte utf8 string
func (e *Encoder) EncodeAmf0LongString(w io.Writer, val string, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0LongStringMarker); err != nil {
			return
		}
		n++
	}

	var m int
	length := uint32(len(val))
	err = binary.Write(w, binary.BigEndian, length)
	if err != nil {
		return n, Error("encode amf0: unable to encode long string length: %s", err)
	}
	n += 4

	m, err = w.Write([]byte(val))
	if err != nil {
		return n, Error("encode amf0: unable to encode long string value: %s", err)
	}
	n += m

	return
}

// EncodeAmf0Unsupported marker: 1 byte 0x0d
// no additional data
func (e *Encoder) EncodeAmf0Unsupported(w io.Writer, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, Amf0UnsupportedMarker); err != nil {
			return
		}
		n++
	}

	return
}

// EncodeAmf0Amf3Marker marker: 1 byte 0x11
func (e *Encoder) EncodeAmf0Amf3Marker(w io.Writer) error {
	return WriteMarker(w, Amf0AcmplusObjectMarker)
}
