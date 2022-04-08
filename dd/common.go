package dd

import "net/http"

type CommonHeader struct {
	Cookie    string
	DeviceId  string
	Longitude string
	Latitude  string
	Uid       string
}

var commonHeader CommonHeader

func InitCommonHeader(header CommonHeader) {
	commonHeader = header
}

func BindCommonHeader(req *http.Request) {
	req.Header.Set("cookie", commonHeader.Cookie)

	req.Header.Set("host", "maicai.api.ddxq.mobi")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("ddmc-build-version", "2.82.0")
	req.Header.Set("ddmc-device-id", commonHeader.DeviceId)
	req.Header.Set("ddmc-channel", "applet")
	req.Header.Set("ddmc-os-version", "[object Undefined]")
	req.Header.Set("ddmc-app-client-id", "4")
	req.Header.Set("ddmc-ip", "")
	req.Header.Set("ddmc-longitude", commonHeader.Longitude)
	req.Header.Set("ddmc-latitude", commonHeader.Latitude)
	req.Header.Set("ddmc-api-version", "9.49.2")
	req.Header.Set("ddmc-uid", commonHeader.Uid)
	req.Header.Set("user-agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.18(0x1800123c) NetType/4G Language/zh_CN")
	req.Header.Set("referer", "https://servicewechat.com/wx1e113254eda17715/422/page-frame.html")

	req.Header.Set("accept", "application/json, text/plain, */*")
}
