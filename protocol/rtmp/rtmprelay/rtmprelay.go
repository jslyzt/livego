package rtmprelay

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/jslyzt/livego/protocol/amf"
	"github.com/jslyzt/livego/protocol/rtmp/core"
)

// 变量
var (
	StopCtrl = "RTMPRELAY_STOP"
)

// RtmpRelay rtmp转播
type RtmpRelay struct {
	PlayURL              string
	PublishURL           string
	csChan               chan core.ChunkStream
	sndctrlChan          chan string
	connectPlayClient    *core.ConnClient
	connectPublishClient *core.ConnClient
	startflag            bool
}

// NewRtmpRelay 新建rtmp转播
func NewRtmpRelay(playurl *string, publishurl *string) *RtmpRelay {
	return &RtmpRelay{
		PlayURL:              *playurl,
		PublishURL:           *publishurl,
		csChan:               make(chan core.ChunkStream, 500),
		sndctrlChan:          make(chan string),
		connectPlayClient:    nil,
		connectPublishClient: nil,
		startflag:            false,
	}
}

func (rtmp *RtmpRelay) rcvPlayChunkStream() {
	log.Println("rcvPlayRtmpMediaPacket connectClient.Read...")
	for {
		var rc core.ChunkStream

		if rtmp.startflag == false {
			rtmp.connectPlayClient.Close(nil)
			log.Printf("rcvPlayChunkStream close: playurl=%s, publishurl=%s", rtmp.PlayURL, rtmp.PublishURL)
			break
		}
		err := rtmp.connectPlayClient.Read(&rc)

		if err != nil && err == io.EOF {
			break
		}
		//log.Printf("connectPlayClient.Read return rc.TypeID=%v length=%d, err=%v", rc.TypeID, len(rc.Data), err)
		switch rc.TypeID {
		case 20, 17:
			r := bytes.NewReader(rc.Data)
			vs, err := rtmp.connectPlayClient.DecodeBatch(r, amf.AMF0)

			log.Printf("rcvPlayRtmpMediaPacket: vs=%v, err=%v", vs, err)
		case 18:
			log.Printf("rcvPlayRtmpMediaPacket: metadata....")
		case 8, 9:
			rtmp.csChan <- rc
		}
	}
}

func (rtmp *RtmpRelay) sendPublishChunkStream() {
	for {
		select {
		case rc := <-rtmp.csChan:
			//log.Printf("sendPublishChunkStream: rc.TypeID=%v length=%d", rc.TypeID, len(rc.Data))
			rtmp.connectPublishClient.Write(rc)
		case ctrlcmd := <-rtmp.sndctrlChan:
			if ctrlcmd == StopCtrl {
				rtmp.connectPublishClient.Close(nil)
				log.Printf("sendPublishChunkStream close: playurl=%s, publishurl=%s", rtmp.PlayURL, rtmp.PublishURL)
				break
			}
		}
	}
}

// Start 开始
func (rtmp *RtmpRelay) Start() error {
	if rtmp.startflag {
		err := fmt.Errorf("The rtmprelay already started, playurl=%s, publishurl=%s", rtmp.PlayURL, rtmp.PublishURL)
		return err
	}

	rtmp.connectPlayClient = core.NewConnClient()
	rtmp.connectPublishClient = core.NewConnClient()

	log.Printf("play server addr:%v starting....", rtmp.PlayURL)
	err := rtmp.connectPlayClient.Start(rtmp.PlayURL, "play")
	if err != nil {
		log.Printf("connectPlayClient.Start url=%v error", rtmp.PlayURL)
		return err
	}

	log.Printf("publish server addr:%v starting....", rtmp.PublishURL)
	err = rtmp.connectPublishClient.Start(rtmp.PublishURL, "publish")
	if err != nil {
		log.Printf("connectPublishClient.Start url=%v error", rtmp.PublishURL)
		rtmp.connectPlayClient.Close(nil)
		return err
	}

	rtmp.startflag = true
	go rtmp.rcvPlayChunkStream()
	go rtmp.sendPublishChunkStream()
	return nil
}

// Stop 停止
func (rtmp *RtmpRelay) Stop() {
	if !rtmp.startflag {
		log.Printf("The rtmprelay already stoped, playurl=%s, publishurl=%s", rtmp.PlayURL, rtmp.PublishURL)
		return
	}

	rtmp.startflag = false
	rtmp.sndctrlChan <- StopCtrl
}
