package amf

import (
	"encoding/json"
	"fmt"
	"io"
)

// DumpBytes 打印比特数组
func DumpBytes(label string, buf []byte, size int) {
	fmt.Printf("Dumping %s (%d bytes):\n", label, size)
	for i := 0; i < size; i++ {
		fmt.Printf("0x%02x ", buf[i])
	}
	fmt.Printf("\n")
}

// Dump 打印
func Dump(label string, val interface{}) error {
	json, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return Error("Error dumping %s: %s", label, err)
	}
	fmt.Printf("Dumping %s:\n%s\n", label, json)
	return nil
}

// Error 返回错误
func Error(f string, v ...interface{}) error {
	return fmt.Errorf(f, v...)
}

// WriteByte 写入比特
func WriteByte(w io.Writer, b byte) (err error) {
	bytes := make([]byte, 1)
	bytes[0] = b

	_, err = WriteBytes(w, bytes)
	return
}

// WriteBytes 写入比特数组
func WriteBytes(w io.Writer, bytes []byte) (int, error) {
	return w.Write(bytes)
}

// ReadByte 读取比特
func ReadByte(r io.Reader) (byte, error) {
	bytes, err := ReadBytes(r, 1)
	if err != nil {
		return 0x00, err
	}

	return bytes[0], nil
}

// ReadBytes 读取比特数组
func ReadBytes(r io.Reader, n int) ([]byte, error) {
	bytes := make([]byte, n)

	m, err := r.Read(bytes)
	if err != nil {
		return bytes, err
	}

	if m != n {
		return bytes, fmt.Errorf("decode read bytes failed: expected %d got %d", m, n)
	}
	return bytes, nil
}

// WriteMarker 写入
func WriteMarker(w io.Writer, m byte) error {
	return WriteByte(w, m)
}

// ReadMarker 读取
func ReadMarker(r io.Reader) (byte, error) {
	return ReadByte(r)
}

// AssertMarker 断言
func AssertMarker(r io.Reader, checkMarker bool, m byte) error {
	if checkMarker == false {
		return nil
	}

	marker, err := ReadMarker(r)
	if err != nil {
		return err
	}

	if marker != m {
		return Error("decode assert marker failed: expected %v got %v", m, marker)
	}
	return nil
}
