package av

import (
	"fmt"
	"io"
)

// 常量
const (
	TagAudio          = 8
	TagVideo          = 9
	TagScriptDataAmF0 = 18
	TagScriptDataAmF3 = 0xf

	MetadatAMF0  = 0x12
	MetadataAMF3 = 0xf

	SoundMp3                 = 2
	SoundNellymoser16khzMono = 4
	SoundNellymoser8khzMono  = 5
	SoundNellymoser          = 6
	SoundAlaw                = 7
	SoundMulaw               = 8
	SoundAac                 = 10
	SoundSpeex               = 11

	Sound55khz = 0
	Sound11khz = 1
	Sound22khz = 2
	Sound44khz = 3

	Sound8bit  = 0
	Sound16bit = 1

	SoundMono   = 0
	SoundStereo = 1

	AacSeqhdr = 0
	AacRaw    = 1

	AvcSeqhdr = 0
	AvcNalu   = 1
	AvcEos    = 2

	FrameKey   = 1
	FrameInter = 2

	VideoH264 = 7
)

// 变量
var (
	PUBLISH = "publish"
	PLAY    = "play"
)

// Packet Header can be converted to AudioHeaderInfo or VideoHeaderInfo
type Packet struct {
	IsAudio    bool
	IsVideo    bool
	IsMetadata bool
	TimeStamp  uint32 // dts
	StreamID   uint32
	Header     PacketHeader
	Data       []byte
}

// PacketHeader 包处理
type PacketHeader interface {
}

// AudioPacketHeader 音频包处理
type AudioPacketHeader interface {
	PacketHeader
	SoundFormat() uint8
	AACPacketType() uint8
}

// VideoPacketHeader 视频包处理
type VideoPacketHeader interface {
	PacketHeader
	IsKeyFrame() bool
	IsSeq() bool
	CodecID() uint8
	CompositionTime() int32
}

// Demuxer 分流器
type Demuxer interface {
	Demux(*Packet) (ret *Packet, err error)
}

// Muxer muxer
type Muxer interface {
	Mux(*Packet, io.Writer) error
}

// SampleRater 简单评估
type SampleRater interface {
	SampleRate() (int, error)
}

// CodecParser 解码
type CodecParser interface {
	SampleRater
	Parse(*Packet, io.Writer) error
}

// GetWriter 写入
type GetWriter interface {
	GetWriter(Info) WriteCloser
}

// Handler 处理
type Handler interface {
	HandleReader(ReadCloser)
	HandleWriter(WriteCloser)
}

// Alive 活跃
type Alive interface {
	Alive() bool
}

// Closer 关闭
type Closer interface {
	Info() Info
	Close(error)
}

// CalcTime 任务
type CalcTime interface {
	CalcBaseTimestamp()
}

// Info 信息
type Info struct {
	Key   string
	URL   string
	UID   string
	Inter bool
}

// IsInterval 时候间隔
func (info Info) IsInterval() bool {
	return info.Inter
}

func (info Info) String() string {
	return fmt.Sprintf("<key: %s, URL: %s, UID: %s, Inter: %v>",
		info.Key, info.URL, info.UID, info.Inter)
}

// ReadCloser 自动读取关闭
type ReadCloser interface {
	Closer
	Alive
	Read(*Packet) error
}

// WriteCloser 自动写完关闭
type WriteCloser interface {
	Closer
	Alive
	CalcTime
	Write(*Packet) error
}
