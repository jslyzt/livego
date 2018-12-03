package hls

import "bytes"

const (
	cacheMaxFrames byte = 6
	cacheAudioLen  int  = 10 * 1024
)

type audioCache struct {
	soundFormat byte
	num         byte
	offset      int
	pts         uint64
	buf         *bytes.Buffer
}

func newAudioCache() *audioCache {
	return &audioCache{
		buf: bytes.NewBuffer(make([]byte, cacheAudioLen)),
	}
}

func (a *audioCache) Cache(src []byte, pts uint64) bool {
	if a.num == 0 {
		a.offset = 0
		a.pts = pts
		a.buf.Reset()
	}
	a.buf.Write(src)
	a.offset += len(src)
	a.num++

	return false
}

func (a *audioCache) GetFrame() (int, uint64, []byte) {
	a.num = 0
	return a.offset, a.pts, a.buf.Bytes()
}

func (a *audioCache) CacheNum() byte {
	return a.num
}
