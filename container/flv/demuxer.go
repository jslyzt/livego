package flv

import (
	"errors"

	"github.com/jslyzt/livego/av"
)

// 变量定义
var (
	ErrAvcEndSEQ = errors.New("avc end sequence")
)

// Demuxer 分流器
type Demuxer struct {
}

// NewDemuxer 新的分流器
func NewDemuxer() *Demuxer {
	return &Demuxer{}
}

// DemuxH 处理packet，只支持header
func (d *Demuxer) DemuxH(p *av.Packet) error {
	var tag Tag
	_, err := tag.ParseMeidaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	p.Header = &tag

	return nil
}

// Demux 处理packet
func (d *Demuxer) Demux(p *av.Packet) error {
	var tag Tag
	n, err := tag.ParseMeidaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	if tag.CodecID() == av.VideoH264 &&
		p.Data[0] == 0x17 && p.Data[1] == 0x02 {
		return ErrAvcEndSEQ
	}
	p.Header = &tag
	p.Data = p.Data[n:]

	return nil
}
