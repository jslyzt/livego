package amf

import (
	"fmt"
	"io"
)

// DecodeBatch 批量解码
func (d *Decoder) DecodeBatch(r io.Reader, ver Version) (ret []interface{}, err error) {
	var v interface{}
	for {
		v, err = d.Decode(r, ver)
		if err != nil {
			break
		}
		ret = append(ret, v)
	}
	return
}

// Decode 解码
func (d *Decoder) Decode(r io.Reader, ver Version) (interface{}, error) {
	switch ver {
	case 0:
		return d.DecodeAmf0(r)
	case 3:
		return d.DecodeAmf3(r)
	}

	return nil, fmt.Errorf("decode amf: unsupported version %d", ver)
}

// EncodeBatch 批量加密
func (e *Encoder) EncodeBatch(w io.Writer, ver Version, val ...interface{}) (int, error) {
	for _, v := range val {
		if _, err := e.Encode(w, v, ver); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

// Encode 加密
func (e *Encoder) Encode(w io.Writer, val interface{}, ver Version) (int, error) {
	switch ver {
	case AMF0:
		return e.EncodeAmf0(w, val)
	case AMF3:
		return e.EncodeAmf3(w, val)
	}

	return 0, Error("encode amf: unsupported version %d", ver)
}
