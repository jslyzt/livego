package cache

import (
	"flag"

	"github.com/jslyzt/livego/av"
)

var (
	gopNum = flag.Int("gopNum", 1, "gop num")
)

// Cache 缓存
type Cache struct {
	gop      *GopCache
	videoSeq *SpecialCache
	audioSeq *SpecialCache
	metadata *SpecialCache
}

// NewCache 新建缓存
func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(*gopNum),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

// Write 写
func (cache *Cache) Write(p av.Packet) {
	if p.IsMetadata {
		cache.metadata.Write(&p)
		return
	}
	if !p.IsVideo {
		ah, ok := p.Header.(av.AudioPacketHeader)
		if ok {
			if ah.SoundFormat() == av.SoundAac &&
				ah.AACPacketType() == av.AacSeqhdr {
				cache.audioSeq.Write(&p)
			}
			return
		}
	} else {
		vh, ok := p.Header.(av.VideoPacketHeader)
		if ok {
			if vh.IsSeq() {
				cache.videoSeq.Write(&p)
				return
			}
		} else {
			return
		}
	}
	cache.gop.Write(&p)
}

// Send 发送
func (cache *Cache) Send(w av.WriteCloser) error {
	if err := cache.metadata.Send(w); err != nil {
		return err
	}

	if err := cache.videoSeq.Send(w); err != nil {
		return err
	}

	if err := cache.audioSeq.Send(w); err != nil {
		return err
	}

	if err := cache.gop.Send(w); err != nil {
		return err
	}
	return nil
}
