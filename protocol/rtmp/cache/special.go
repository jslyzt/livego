package cache

import (
	"bytes"
	"log"

	"github.com/jslyzt/livego/av"
	"github.com/jslyzt/livego/protocol/amf"
)

// 常量
const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}
	if _, err := encoder.Encode(b, SetDataFrame, amf.AMF0); err != nil {
		log.Fatal(err)
	}
	setFrameFrame = b.Bytes()
}

// SpecialCache 特殊cache
type SpecialCache struct {
	full bool
	p    *av.Packet
}

// NewSpecialCache new
func NewSpecialCache() *SpecialCache {
	return &SpecialCache{}
}

// Write 写入
func (specialCache *SpecialCache) Write(p *av.Packet) {
	specialCache.p = p
	specialCache.full = true
}

// Send 发送
func (specialCache *SpecialCache) Send(w av.WriteCloser) error {
	if !specialCache.full {
		return nil
	}
	return w.Write(specialCache.p)
}
