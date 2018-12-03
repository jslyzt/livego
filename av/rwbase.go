package av

import (
	"sync"
	"time"
)

// RWBaser 基础读写
type RWBaser struct {
	lock               sync.Mutex
	timeout            time.Duration
	PreTime            time.Time
	BaseTimestamp      uint32
	LastVideoTimestamp uint32
	LastAudioTimestamp uint32
}

// NewRWBaser new方法
func NewRWBaser(duration time.Duration) RWBaser {
	return RWBaser{
		timeout: duration,
		PreTime: time.Now(),
	}
}

// BaseTimeStamp 时间戳
func (rw *RWBaser) BaseTimeStamp() uint32 {
	return rw.BaseTimestamp
}

// CalcBaseTimestamp 取消时间戳
func (rw *RWBaser) CalcBaseTimestamp() {
	if rw.LastAudioTimestamp > rw.LastVideoTimestamp {
		rw.BaseTimestamp = rw.LastAudioTimestamp
	} else {
		rw.BaseTimestamp = rw.LastVideoTimestamp
	}
}

// RecTimeStamp 接收时间戳
func (rw *RWBaser) RecTimeStamp(timestamp, typeID uint32) {
	if typeID == TagVideo {
		rw.LastVideoTimestamp = timestamp
	} else if typeID == TagAudio {
		rw.LastAudioTimestamp = timestamp
	}
}

// SetPreTime 设置timer
func (rw *RWBaser) SetPreTime() {
	rw.lock.Lock()
	rw.PreTime = time.Now()
	rw.lock.Unlock()
}

// Alive 活跃
func (rw *RWBaser) Alive() bool {
	rw.lock.Lock()
	b := !(time.Now().Sub(rw.PreTime) >= rw.timeout)
	rw.lock.Unlock()
	return b
}
