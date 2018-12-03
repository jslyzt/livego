package amf

import (
	"io"
)

// 常量定义
const (
	AMF0 = 0x00
	AMF3 = 0x03

	Amf0BooleanFalse = 0x00
	Amf0BooleanTrue  = 0x01
	Amf0StringMax    = 65535
	Amf3IntegerMax   = 536870911
)

// AMF0 定义
const (
	Amf0NumberMarker        = 0x00
	Amf0BooleanMarker       = 0x01
	Amf0StringMarker        = 0x02
	Amf0ObjectMarker        = 0x03
	Amf0MovieclipMarker     = 0x04
	Amf0NullMarker          = 0x05
	Amf0UndefinedMarker     = 0x06
	Amf0ReferenceMarker     = 0x07
	Amf0EcmaArrayMarker     = 0x08
	Amf0ObjectEndMarker     = 0x09
	Amf0StrictArrayMarker   = 0x0a
	Amf0DateMarker          = 0x0b
	Amf0LongStringMarker    = 0x0c
	Amf0UnsupportedMarker   = 0x0d
	Amf0RecordsetMarker     = 0x0e
	Amf0XmlDocumentMarker   = 0x0f
	Amf0TypedObjectMarker   = 0x10
	Amf0AcmplusObjectMarker = 0x11
)

// AMF3 定义
const (
	Amf3UndefinedMarker = 0x00
	Amf3NullMarker      = 0x01
	Amf3FalseMarker     = 0x02
	Amf3TrueMarker      = 0x03
	Amf3IntegerMarker   = 0x04
	Amf3DoubleMarker    = 0x05
	Amf3StringMarker    = 0x06
	Amf3XmldocMarker    = 0x07
	Amf3DateMarker      = 0x08
	Amf3ArrayMarker     = 0x09
	Amf3ObjectMarker    = 0x0a
	Amf3XmlstringMarker = 0x0b
	Amf3BytearrayMarker = 0x0c
)

// ExternalHandler 外部处理器
type ExternalHandler func(*Decoder, io.Reader) (interface{}, error)

// Decoder 解码
type Decoder struct {
	refCache         []interface{}
	stringRefs       []string
	objectRefs       []interface{}
	traitRefs        []Trait
	externalHandlers map[string]ExternalHandler
}

// NewDecoder 新解码器
func NewDecoder() *Decoder {
	return &Decoder{
		externalHandlers: make(map[string]ExternalHandler),
	}
}

// RegisterExternalHandler 注册外部处理器
func (d *Decoder) RegisterExternalHandler(name string, f ExternalHandler) {
	d.externalHandlers[name] = f
}

// Encoder 加密
type Encoder struct {
}

// Version 版本
type Version uint8

// Array 数组
type Array []interface{}

// Object 对象
type Object map[string]interface{}

// TypedObject 类型对象
type TypedObject struct {
	Type   string
	Object Object
}

// Trait 特性
type Trait struct {
	Type           string
	Externalizable bool
	Dynamic        bool
	Properties     []string
}

// NewTrait 新特性
func NewTrait() *Trait {
	return &Trait{}
}

// NewTypedObject 新类型对象
func NewTypedObject() *TypedObject {
	return &TypedObject{
		Type:   "",
		Object: make(Object),
	}
}
