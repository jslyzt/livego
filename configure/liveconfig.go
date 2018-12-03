package configure

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Application 应用
type Application struct {
	Appname    string
	Liveon     string
	Hlson      string
	StaticPush []string
}

// ServerCfg 服务配置
type ServerCfg struct {
	Server []Application
}

// RtmpServercfg rtmp服务配置
var RtmpServercfg ServerCfg

// LoadConfig 加载配置
func LoadConfig(configfilename string) error {
	log.Printf("starting load configure file(%s)......", configfilename)
	data, err := ioutil.ReadFile(configfilename)
	if err != nil {
		log.Printf("ReadFile %s error:%v", configfilename, err)
		return err
	}

	log.Printf("loadconfig: \r\n%s", string(data))

	err = json.Unmarshal(data, &RtmpServercfg)
	if err != nil {
		log.Printf("json.Unmarshal error:%v", err)
		return err
	}
	log.Printf("get config json data:%v", RtmpServercfg)
	return nil
}

// CheckAppName 检查app名字
func CheckAppName(appname string) bool {
	for _, app := range RtmpServercfg.Server {
		if (app.Appname == appname) && (app.Liveon == "on") {
			return true
		}
	}
	return false
}

// GetStaticPushURLList 获取静态推送列表
func GetStaticPushURLList(appname string) ([]string, bool) {
	for _, app := range RtmpServercfg.Server {
		if (app.Appname == appname) && (app.Liveon == "on") {
			if len(app.StaticPush) > 0 {
				return app.StaticPush, true
			}
			return nil, false
		}

	}
	return nil, false
}
