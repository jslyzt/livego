package httpopera

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/jslyzt/livego/av"
	"github.com/jslyzt/livego/protocol/rtmp"
	"github.com/jslyzt/livego/protocol/rtmp/rtmprelay"
)

// Response 返回
type Response struct {
	w       http.ResponseWriter
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// SendJSON 发送json
func (r *Response) SendJSON() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
	return r.w.Write(resp)
}

// Operation 操作
type Operation struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Stop   bool   `json:"stop"`
}

// OperationChange 操作变化
type OperationChange struct {
	Method    string `json:"method"`
	SourceURL string `json:"source_url"`
	TargetURL string `json:"target_url"`
	Stop      bool   `json:"stop"`
}

// ClientInfo 客户端信息
type ClientInfo struct {
	url              string
	rtmpRemoteClient *rtmp.Client
	rtmpLocalClient  *rtmp.Client
}

// Server 服务
type Server struct {
	handler  av.Handler
	session  map[string]*rtmprelay.RtmpRelay
	rtmpAddr string
}

// NewServer 新建服务
func NewServer(h av.Handler, rtmpAddr string) *Server {
	return &Server{
		handler:  h,
		session:  make(map[string]*rtmprelay.RtmpRelay),
		rtmpAddr: rtmpAddr,
	}
}

// Serve 服务
func (s *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()

	mux.Handle("/statics", http.FileServer(http.Dir("statics")))

	mux.HandleFunc("/control/push", func(w http.ResponseWriter, r *http.Request) {
		s.handlePush(w, r)
	})
	mux.HandleFunc("/control/pull", func(w http.ResponseWriter, r *http.Request) {
		s.handlePull(w, r)
	})
	mux.HandleFunc("/stat/livestat", func(w http.ResponseWriter, r *http.Request) {
		s.GetLiveStatics(w, r)
	})
	http.Serve(l, mux)
	return nil
}

type stream struct {
	Key             string `json:"key"`
	URL             string `json:"Url"`
	StreamID        uint32 `json:"StreamId"`
	VideoTotalBytes uint64 `json:"videototalbytes"`
	VideoSpeed      uint64 `json:"videospeed"`
	AudioTotalBytes uint64 `json:"audiototalbytes"`
	AudioSpeed      uint64 `json:"audiospeed"`
}

type streams struct {
	Publishers []stream `json:"publishers"`
	Players    []stream `json:"players"`
}

//http://127.0.0.1:8090/stat/livestat

// GetLiveStatics 直播
func (s *Server) GetLiveStatics(w http.ResponseWriter, req *http.Request) {
	rtmpStream := s.handler.(*rtmp.RtmpStream)
	if rtmpStream == nil {
		io.WriteString(w, "<h1>Get rtmp stream information error</h1>")
		return
	}

	msgs := new(streams)
	for item := range rtmpStream.GetStreams().IterBuffered() {
		if s, ok := item.Val.(*rtmp.Stream); ok {
			if s.GetReader() != nil {
				switch s.GetReader().(type) {
				case *rtmp.VirReader:
					v := s.GetReader().(*rtmp.VirReader)
					msg := stream{item.Key, v.Info().URL, v.ReadBWInfo.StreamID, v.ReadBWInfo.VideoDatainBytes, v.ReadBWInfo.VideoSpeedInBytesperMS,
						v.ReadBWInfo.AudioDatainBytes, v.ReadBWInfo.AudioSpeedInBytesperMS}
					msgs.Publishers = append(msgs.Publishers, msg)
				}
			}
		}
	}

	for item := range rtmpStream.GetStreams().IterBuffered() {
		ws := item.Val.(*rtmp.Stream).GetWs()
		for s := range ws.IterBuffered() {
			if pw, ok := s.Val.(*rtmp.PackWriterCloser); ok {
				if pw.GetWriter() != nil {
					switch pw.GetWriter().(type) {
					case *rtmp.VirWriter:
						v := pw.GetWriter().(*rtmp.VirWriter)
						msg := stream{item.Key, v.Info().URL, v.WriteBWInfo.StreamID, v.WriteBWInfo.VideoDatainBytes, v.WriteBWInfo.VideoSpeedInBytesperMS,
							v.WriteBWInfo.AudioDatainBytes, v.WriteBWInfo.AudioSpeedInBytesperMS}
						msgs.Players = append(msgs.Players, msg)
					}
				}
			}
		}
	}
	resp, _ := json.Marshal(msgs)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

//http://127.0.0.1:8090/control/push?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456
func (s *Server) handlePull(w http.ResponseWriter, req *http.Request) {
	var retString string
	var err error

	req.ParseForm()

	oper := req.Form["oper"]
	app := req.Form["app"]
	name := req.Form["name"]
	url := req.Form["url"]

	log.Printf("control pull: oper=%v, app=%v, name=%v, url=%v", oper, app, name, url)
	if (len(app) <= 0) || (len(name) <= 0) || (len(url) <= 0) {
		io.WriteString(w, "control push parameter error, please check them.</br>")
		return
	}

	remoteurl := "rtmp://127.0.0.1" + s.rtmpAddr + "/" + app[0] + "/" + name[0]
	localurl := url[0]

	keyString := "pull:" + app[0] + "/" + name[0]
	if oper[0] == "stop" {
		pullRtmprelay, found := s.session[keyString]

		if !found {
			retString = fmt.Sprintf("session key[%s] not exist, please check it again.", keyString)
			io.WriteString(w, retString)
			return
		}
		log.Printf("rtmprelay stop push %s from %s", remoteurl, localurl)
		pullRtmprelay.Stop()

		delete(s.session, keyString)
		retString = fmt.Sprintf("<h1>push url stop %s ok</h1></br>", url[0])
		io.WriteString(w, retString)
		log.Printf("pull stop return %s", retString)
	} else {
		pullRtmprelay := rtmprelay.NewRtmpRelay(&localurl, &remoteurl)
		log.Printf("rtmprelay start push %s from %s", remoteurl, localurl)
		err = pullRtmprelay.Start()
		if err != nil {
			retString = fmt.Sprintf("push error=%v", err)
		} else {
			s.session[keyString] = pullRtmprelay
			retString = fmt.Sprintf("<h1>push url start %s ok</h1></br>", url[0])
		}
		io.WriteString(w, retString)
		log.Printf("pull start return %s", retString)
	}
}

//http://127.0.0.1:8090/control/push?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456
func (s *Server) handlePush(w http.ResponseWriter, req *http.Request) {
	var retString string
	var err error

	req.ParseForm()

	oper := req.Form["oper"]
	app := req.Form["app"]
	name := req.Form["name"]
	url := req.Form["url"]

	log.Printf("control push: oper=%v, app=%v, name=%v, url=%v", oper, app, name, url)
	if (len(app) <= 0) || (len(name) <= 0) || (len(url) <= 0) {
		io.WriteString(w, "control push parameter error, please check them.</br>")
		return
	}

	localurl := "rtmp://127.0.0.1" + s.rtmpAddr + "/" + app[0] + "/" + name[0]
	remoteurl := url[0]

	keyString := "push:" + app[0] + "/" + name[0]
	if oper[0] == "stop" {
		pushRtmprelay, found := s.session[keyString]
		if !found {
			retString = fmt.Sprintf("<h1>session key[%s] not exist, please check it again.</h1>", keyString)
			io.WriteString(w, retString)
			return
		}
		log.Printf("rtmprelay stop push %s from %s", remoteurl, localurl)
		pushRtmprelay.Stop()

		delete(s.session, keyString)
		retString = fmt.Sprintf("<h1>push url stop %s ok</h1></br>", url[0])
		io.WriteString(w, retString)
		log.Printf("push stop return %s", retString)
	} else {
		pushRtmprelay := rtmprelay.NewRtmpRelay(&localurl, &remoteurl)
		log.Printf("rtmprelay start push %s from %s", remoteurl, localurl)
		err = pushRtmprelay.Start()
		if err != nil {
			retString = fmt.Sprintf("push error=%v", err)
		} else {
			retString = fmt.Sprintf("<h1>push url start %s ok</h1></br>", url[0])
			s.session[keyString] = pushRtmprelay
		}

		io.WriteString(w, retString)
		log.Printf("push start return %s", retString)
	}
}
