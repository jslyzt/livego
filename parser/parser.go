package parser

import (
	"errors"
	"io"

	"github.com/jslyzt/livego/av"
	"github.com/jslyzt/livego/parser/aac"
	"github.com/jslyzt/livego/parser/h264"
	"github.com/jslyzt/livego/parser/mp3"
)

var (
	errNoAudio = errors.New("demuxer no audio")
)

// CodecParser 解码器
type CodecParser struct {
	aac  *aac.Parser
	mp3  *mp3.Parser
	h264 *h264.Parser
}

// NewCodecParser 新解码器
func NewCodecParser() *CodecParser {
	return &CodecParser{}
}

// SampleRate 简单码率
func (codeParser *CodecParser) SampleRate() (int, error) {
	if codeParser.aac == nil && codeParser.mp3 == nil {
		return 0, errNoAudio
	}
	if codeParser.aac != nil {
		return codeParser.aac.SampleRate(), nil
	}
	return codeParser.mp3.SampleRate(), nil
}

// Parse 解码
func (codeParser *CodecParser) Parse(p *av.Packet, w io.Writer) (err error) {

	switch p.IsVideo {
	case true:
		f, ok := p.Header.(av.VideoPacketHeader)
		if ok {
			if f.CodecID() == av.VideoH264 {
				if codeParser.h264 == nil {
					codeParser.h264 = h264.NewParser()
				}
				err = codeParser.h264.Parse(p.Data, f.IsSeq(), w)
			}
		}
	case false:
		f, ok := p.Header.(av.AudioPacketHeader)
		if ok {
			switch f.SoundFormat() {
			case av.SoundAac:
				if codeParser.aac == nil {
					codeParser.aac = aac.NewParser()
				}
				err = codeParser.aac.Parse(p.Data, f.AACPacketType(), w)
			case av.SoundMp3:
				if codeParser.mp3 == nil {
					codeParser.mp3 = mp3.NewParser()
				}
				err = codeParser.mp3.Parse(p.Data)
			}
		}
	}
	return
}
