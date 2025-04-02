package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	_log "github.com/typenoob/iptel-ntp-config/log"
	"github.com/typenoob/iptel-ntp-config/util"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type Equiinet struct {
	name   string
	config DeviceConfig
}

func (e *Equiinet) setNTP(ip net.IP, ntp net.IP) *_log.Entry {
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
	res := new(_log.Entry)
	payload := fmt.Sprintf(e.config.Payload, ntp)
	req := dac.NewRequest(e.config.Username, e.config.Password, "POST", fmt.Sprintf(e.config.SetUri, ip), payload)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := req.Execute()
	if err != nil {
		log.Println(err)
		// 若服务器主动关闭请求则跳过
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Println(err)
		// 返回的不是JSON则跳过
		return nil
	}
	res.IP = ip
	res.Ntp = ntp
	res.Time = time.Now()
	res.DeviceName = e.name
	return res
}

func (e *Equiinet) getNTP(ip net.IP) *net.IP {
	var res map[string]any
	const USERNAME = "admin"
	const PASSWORD = "admin"
	req := dac.NewRequest(USERNAME, PASSWORD, "GET", fmt.Sprintf(e.config.GetUri, ip), "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := req.Execute()
	if err != nil {
		log.Println(err)
		// 若服务器主动关闭请求则跳过
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Println(err)
		// 返回的不是JSON则跳过
		return nil
	}
	ntp := net.ParseIP(res["primaryNtp"].(string))
	return &ntp
}

func (e *Equiinet) IsMatch(ip net.IP) bool {
	resp, err := http.Get(fmt.Sprintf(e.config.TestUri, ip))
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func (e *Equiinet) GetName() string {
	return e.name
}

func init() {
	RegisterDriver(func() NTPConfigDriver {
		const NAME = "EQUIINET"
		return &Equiinet{name: NAME, config: DeviceConfig{
			Name:     NAME,
			Username: "admin",
			Password: "admin",
			Payload:  "dhcpTime=0&timeZone=92&primaryNtp=%s&secondaryNtp=time.windows.com&updateInterval=1000&daylight=0&fixedType=0&startMonth=1&startDate=1&startHourDay=0&startDayWeek=0&startWeekMonth=1&stopMonth=1&stopDate=1&stopHourDay=0&stopDayWeek=0&stopWeekMonth=1&offset=0&manualTime=0&dateYmd=&timeHms=&timeFormat=0&dateFormat=0&backlightTime=60&backlightLevel=9&ringTones=1&user_set_phone_preference",
			SetUri:   "http://%s/cgi-bin/web_cgi_main.cgi?user_set_phone_preference",
			GetUri:   "http://%s/cgi-bin/web_cgi_main.cgi?user_get_phone_preference",
			TestUri:  "http://%s/images/earth.png",
		}}
	})
}
