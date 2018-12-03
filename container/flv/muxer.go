package flv

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jslyzt/livego/av"
	"github.com/jslyzt/livego/protocol/amf"
	"github.com/jslyzt/livego/utils/pio"
	"github.com/jslyzt/livego/utils/uid"
)

// 变量定义
var (
	flvHeader = []byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
	flvFile   = flag.String("filFile", "./out.flv", "output flv file name")
)

const (
	headerLen = 11
)

// NewFlv 创建flv
func NewFlv(handler av.Handler, info av.Info) {
	patths := strings.SplitN(info.Key, "/", 2)

	if len(patths) != 2 {
		log.Println("invalid info")
		return
	}

	w, err := os.OpenFile(*flvFile, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Println("open file error: ", err)
	}

	writer := NewFLVWriter(patths[0], patths[1], info.URL, w)

	handler.HandleWriter(writer)

	writer.Wait()
	// close flv file
	log.Println("close flv file")
	writer.ctx.Close()
}

// Writer flv写入
type Writer struct {
	UID string
	av.RWBaser
	app, title, url string
	buf             []byte
	closed          chan struct{}
	ctx             *os.File
}

// NewFLVWriter 新建flv写入
func NewFLVWriter(app, title, url string, ctx *os.File) *Writer {
	ret := &Writer{
		UID:     uid.NewID(),
		app:     app,
		title:   title,
		url:     url,
		ctx:     ctx,
		RWBaser: av.NewRWBaser(time.Second * 10),
		closed:  make(chan struct{}),
		buf:     make([]byte, headerLen),
	}

	ret.ctx.Write(flvHeader)
	pio.PutI32BE(ret.buf[:4], 0)
	ret.ctx.Write(ret.buf[:4])

	return ret
}

// Write 写入接口
func (writer *Writer) Write(p *av.Packet) error {
	writer.RWBaser.SetPreTime()
	h := writer.buf[:headerLen]
	typeID := av.TagVideo
	if !p.IsVideo {
		if p.IsMetadata {
			var err error
			typeID = av.TagScriptDataAmF0
			p.Data, err = amf.MetaDataReform(p.Data, amf.DEL)
			if err != nil {
				return err
			}
		} else {
			typeID = av.TagAudio
		}
	}

	dataLen := len(p.Data)
	timestamp := p.TimeStamp
	timestamp += writer.BaseTimeStamp()
	writer.RWBaser.RecTimeStamp(timestamp, uint32(typeID))

	preDataLen := dataLen + headerLen
	timestampbase := timestamp & 0xffffff
	timestampExt := timestamp >> 24 & 0xff

	pio.PutU8(h[0:1], uint8(typeID))
	pio.PutI24BE(h[1:4], int32(dataLen))
	pio.PutI24BE(h[4:7], int32(timestampbase))
	pio.PutU8(h[7:8], uint8(timestampExt))

	if _, err := writer.ctx.Write(h); err != nil {
		return err
	}

	if _, err := writer.ctx.Write(p.Data); err != nil {
		return err
	}

	pio.PutI32BE(h[:4], int32(preDataLen))
	if _, err := writer.ctx.Write(h[:4]); err != nil {
		return err
	}

	return nil
}

// Wait 等待
func (writer *Writer) Wait() {
	select {
	case <-writer.closed:
		return
	}
}

// Close 关闭
func (writer *Writer) Close(error) {
	writer.ctx.Close()
	close(writer.closed)
}

// Info 信息
func (writer *Writer) Info() (ret av.Info) {
	ret.UID = writer.UID
	ret.URL = writer.url
	ret.Key = writer.app + "/" + writer.title
	return
}
