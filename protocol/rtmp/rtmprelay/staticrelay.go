package rtmprelay

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jslyzt/livego/av"
	"github.com/jslyzt/livego/configure"
	"github.com/jslyzt/livego/protocol/rtmp/core"
)

// StaticPush 静态push
type StaticPush struct {
	RtmpURL       string
	packetChan    chan *av.Packet
	sndctrlChan   chan string
	connectClient *core.ConnClient
	startflag     bool
}

var gStaticPushMap = make(map[string](*StaticPush))
var gMapLock = new(sync.RWMutex)

var (
	staticRelayStopCtrl = "STATIC_RTMPRELAY_STOP"
)

// GetStaticPushList 获取静态推送列表
func GetStaticPushList(appname string) ([]string, error) {
	pushurlList, ok := configure.GetStaticPushURLList(appname)
	if !ok {
		return nil, errors.New("no static push url")
	}
	return pushurlList, nil
}

// GetAndCreateStaticPushObject 创建静态推送对象
func GetAndCreateStaticPushObject(rtmpurl string) *StaticPush {
	gMapLock.RLock()
	staticpush, ok := gStaticPushMap[rtmpurl]
	log.Printf("GetAndCreateStaticPushObject: %s, return %v", rtmpurl, ok)
	if !ok {
		gMapLock.RUnlock()
		newStaticpush := NewStaticPush(rtmpurl)

		gMapLock.Lock()
		gStaticPushMap[rtmpurl] = newStaticpush
		gMapLock.Unlock()

		return newStaticpush
	}
	gMapLock.RUnlock()

	return staticpush
}

// GetStaticPushObject 获取静态推送对象
func GetStaticPushObject(rtmpurl string) (*StaticPush, error) {
	gMapLock.RLock()
	if staticpush, ok := gStaticPushMap[rtmpurl]; ok {
		gMapLock.RUnlock()
		return staticpush, nil
	}
	gMapLock.RUnlock()

	return nil, fmt.Errorf("gStaticPushMap[%v] not exist", rtmpurl)
}

// ReleaseStaticPushObject 释放静态推送对象
func ReleaseStaticPushObject(rtmpurl string) {
	gMapLock.RLock()
	if _, ok := gStaticPushMap[rtmpurl]; ok {
		gMapLock.RUnlock()

		log.Printf("ReleaseStaticPushObject %s ok", rtmpurl)
		gMapLock.Lock()
		delete(gStaticPushMap, rtmpurl)
		gMapLock.Unlock()
	} else {
		gMapLock.RUnlock()
		log.Printf("ReleaseStaticPushObject: not find %s", rtmpurl)
	}
}

// NewStaticPush 新的静态推送资源
func NewStaticPush(rtmpurl string) *StaticPush {
	return &StaticPush{
		RtmpURL:       rtmpurl,
		packetChan:    make(chan *av.Packet, 500),
		sndctrlChan:   make(chan string),
		connectClient: nil,
		startflag:     false,
	}
}

// Start 开始
func (push *StaticPush) Start() error {
	if push.startflag {
		return fmt.Errorf("StaticPush already start %s", push.RtmpURL)
	}

	push.connectClient = core.NewConnClient()

	log.Printf("static publish server addr:%v starting....", push.RtmpURL)
	err := push.connectClient.Start(push.RtmpURL, "publish")
	if err != nil {
		log.Printf("connectClient.Start url=%v error", push.RtmpURL)
		return err
	}
	log.Printf("static publish server addr:%v started, streamid=%d", push.RtmpURL, push.connectClient.GetStreamID())
	go push.HandleAvPacket()

	push.startflag = true
	return nil
}

// Stop 关闭
func (push *StaticPush) Stop() {
	if !push.startflag {
		return
	}

	log.Printf("StaticPush Stop: %s", push.RtmpURL)
	push.sndctrlChan <- staticRelayStopCtrl
	push.startflag = false
}

// WriteAvPacket 写入
func (push *StaticPush) WriteAvPacket(packet *av.Packet) {
	if !push.startflag {
		return
	}
	push.packetChan <- packet
}

func (push *StaticPush) sendPacket(p *av.Packet) {
	if !push.startflag {
		return
	}
	var cs core.ChunkStream

	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = push.connectClient.GetStreamID()
	cs.Timestamp = p.TimeStamp
	//cs.Timestamp += v.BaseTimeStamp()

	//log.Printf("Static sendPacket: rtmpurl=%s, length=%d, streamid=%d",
	//	push.RtmpUrl, len(p.Data), cs.StreamID)
	if p.IsVideo {
		cs.TypeID = av.TagVideo
	} else {
		if p.IsMetadata {
			cs.TypeID = av.TagScriptDataAmF0
		} else {
			cs.TypeID = av.TagAudio
		}
	}
	push.connectClient.Write(cs)
}

// HandleAvPacket 处理
func (push *StaticPush) HandleAvPacket() {
	if !push.IsStart() {
		log.Printf("static push %s not started", push.RtmpURL)
		return
	}

	for {
		select {
		case packet := <-push.packetChan:
			push.sendPacket(packet)
		case ctrlcmd := <-push.sndctrlChan:
			if ctrlcmd == staticRelayStopCtrl {
				push.connectClient.Close(nil)
				log.Printf("Static HandleAvPacket close: publishurl=%s", push.RtmpURL)
				break
			}
		}
	}
}

// IsStart 是否已启动
func (push *StaticPush) IsStart() bool {
	return push.startflag
}
