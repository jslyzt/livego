package core

import (
	"bufio"
	"io"
)

// ReadWriter 读写
type ReadWriter struct {
	*bufio.ReadWriter
	readError  error
	writeError error
}

// NewReadWriter 创建读写对象
func NewReadWriter(rw io.ReadWriter, bufSize int) *ReadWriter {
	return &ReadWriter{
		ReadWriter: bufio.NewReadWriter(bufio.NewReaderSize(rw, bufSize), bufio.NewWriterSize(rw, bufSize)),
	}
}

// Read 读
func (rw *ReadWriter) Read(p []byte) (int, error) {
	if rw.readError != nil {
		return 0, rw.readError
	}
	n, err := io.ReadAtLeast(rw.ReadWriter, p, len(p))
	rw.readError = err
	return n, err
}

// ReadError 读取错误
func (rw *ReadWriter) ReadError() error {
	return rw.readError
}

// ReadUintBE 读取
func (rw *ReadWriter) ReadUintBE(n int) (uint32, error) {
	if rw.readError != nil {
		return 0, rw.readError
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		b, err := rw.ReadByte()
		if err != nil {
			rw.readError = err
			return 0, err
		}
		ret = ret<<8 + uint32(b)
	}
	return ret, nil
}

// ReadUintLE 读取
func (rw *ReadWriter) ReadUintLE(n int) (uint32, error) {
	if rw.readError != nil {
		return 0, rw.readError
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		b, err := rw.ReadByte()
		if err != nil {
			rw.readError = err
			return 0, err
		}
		ret += uint32(b) << uint32(i*8)
	}
	return ret, nil
}

// Flush 刷新
func (rw *ReadWriter) Flush() error {
	if rw.writeError != nil {
		return rw.writeError
	}

	if rw.ReadWriter.Writer.Buffered() == 0 {
		return nil
	}
	return rw.ReadWriter.Flush()
}

// Write 写
func (rw *ReadWriter) Write(p []byte) (int, error) {
	if rw.writeError != nil {
		return 0, rw.writeError
	}
	return rw.ReadWriter.Write(p)
}

// WriteError 写
func (rw *ReadWriter) WriteError() error {
	return rw.writeError
}

// WriteUintBE 写
func (rw *ReadWriter) WriteUintBE(v uint32, n int) error {
	if rw.writeError != nil {
		return rw.writeError
	}
	for i := 0; i < n; i++ {
		b := byte(v>>uint32((n-i-1)<<3)) & 0xff
		if err := rw.WriteByte(b); err != nil {
			rw.writeError = err
			return err
		}
	}
	return nil
}

// WriteUintLE 写
func (rw *ReadWriter) WriteUintLE(v uint32, n int) error {
	if rw.writeError != nil {
		return rw.writeError
	}
	for i := 0; i < n; i++ {
		b := byte(v) & 0xff
		if err := rw.WriteByte(b); err != nil {
			rw.writeError = err
			return err
		}
		v = v >> 8
	}
	return nil
}
