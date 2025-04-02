package driver

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_log "github.com/typenoob/iptel-ntp-config/log"
	"github.com/typenoob/iptel-ntp-config/util"
)

type Guowei struct {
	name   string
	config DeviceConfig
}

const LOGIN_URI = "http://%s/key==nonce"

var client *http.Client

func (e *Guowei) setNTP(ip net.IP, ntp net.IP) *_log.Entry {
	if !e.IsMatch(ip) {
		_log.SetNameByIP(ip, util.UNKNOWN_DEVICE)
		// 若驱动不匹配，无需执行
		return nil
	}
	if _log.GetNameByIP(ip) == util.HISTORY_DEVICE {
		// 若历史记录中已成功执行，无需执行
		_log.SetNameByIP(ip, e.name)
		return nil
	}
	_log.SetNameByIP(ip, e.name)
	res := &_log.Entry{
		IP:         ip,
		Ntp:        ntp,
		Time:       time.Now(),
		DeviceName: e.name,
	}
	jar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar:       jar,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	defer client.CloseIdleConnections()
	// --------------------------登录-----------------------------
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(LOGIN_URI, ip), nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Connection", "keep-alive")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		// 若发送请求错误，则跳过
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusServiceUnavailable {
		// 若服务不可用，则返回错误
		res.Code = -3
		res.Prompt = "Server Busy"
		return res
	}
	rawAuth, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	auth := strings.TrimSpace(string(rawAuth))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	url, _ := url.Parse(fmt.Sprintf("http://%s", ip))
	client.Jar.SetCookies(url, []*http.Cookie{{Name: "auth", Value: auth, Path: "/"}})
	encoded := md5.Sum(fmt.Appendf(nil, "admin:admin:%s", auth))
	req, err = http.NewRequest(http.MethodPost, url.String(), bytes.NewBufferString(fmt.Sprintf("encoded=admin:%x", encoded)))
	if err != nil {
		log.Fatalln(err)
	}
	client.Do(req)
	// --------------------------配置-----------------------------
	payload := fmt.Sprintf(e.config.Payload, ntp)
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf(e.config.SetUri, ip), strings.NewReader(payload))
	req.Header.Set("Connection", "keep-alive")
	if err != nil {
		log.Fatalln(err)
	}
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		// 若发送请求错误则跳过
		return nil
	}
	defer resp.Body.Close()
	if e.getNTP(ip).Equal(ntp) {
		res.Code = 0
		res.Prompt = "Save Success"
	} else {
		res.Code = -1
		res.Prompt = "Save Failed"
	}
	return res
}

func (e *Guowei) getNTP(ip net.IP) *net.IP {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(e.config.GetUri, ip), nil)
	req.Header.Set("Connection", "keep-alive")
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		// 若发送请求错误则跳过
		return nil
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
		// 若解析HTML错误则跳过
		return nil
	}
	ntp := net.ParseIP(doc.Find("[name=TIM_SntpServer_RW]").AttrOr("value", ""))
	return &ntp
}

func (e *Guowei) IsMatch(ip net.IP) bool {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(e.config.TestUri, ip), nil)
	req.Header.Set("Connection", "keep-alive")
	if err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable
}

func (e *Guowei) GetName() string {
	return e.name
}

func init() {
	RegisterDriver(func() NTPConfigDriver {
		const NAME = "GUOWEI"
		return &Guowei{name: NAME, config: DeviceConfig{
			Name:     NAME,
			Username: "admin",
			Password: "admin",
			Payload:  "TIM_EnableSntp_RW=ON&TIM_SntpServer_RW=%s&TIM_SecSntpServer_RW=pool.ntp.org&TIM_SntpZone=32&TIM_SntpTimeOut_RW=60&TIM_DateFormat_RW=0&TIM_DateSeperator_RW=0&TIM_ConturyLocation=1&TIM_DaylightSetEnable_RW=1&DefaultSubmit=提交",
			SetUri:   "http://%s/time.htm",
			GetUri:   "http://%s/time.htm",
			TestUri:  "http://%s/logon_bg.gif",
		}}
	})
}
