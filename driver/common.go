package driver

import (
	"errors"
	"net"
	"time"

	_log "github.com/typenoob/iptel-ntp-config/log"
	"github.com/typenoob/iptel-ntp-config/util"
)

type NTPConfigDriver interface {
	setNTP(ip net.IP, ntp net.IP) *_log.Entry
	getNTP(ip net.IP) *net.IP
	IsMatch(ip net.IP) bool
	GetName() string
}

type IPOrNet interface {
	isIPOrNet()
}

type NTPConfig struct {
	ip    net.IP
	ipNet *net.IPNet
	ntp   net.IP
}

type DeviceConfig struct {
	Name     string
	Username string
	Password string
	Payload  string
	SetUri   string
	GetUri   string
	TestUri  string
}

func (n *NTPConfig) ProbeWebService() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(n.ip.String(), "80"), 2*time.Second)
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		// 若此IP未启用web服务
		return err
	}
	return nil
}

func (n *NTPConfig) Reload(s string) error {
	ip, ipNet, err := net.ParseCIDR(s)
	if err == nil {
		n.ip = ip
		n.ipNet = ipNet
	}
	ip = net.ParseIP(s)
	if ip != nil {
		n.ip = ip
	}
	return errors.New("错误：输入的参数既不是IP地址也不是IP段！")
}

func (n NTPConfig) ExecuteAndLog(driver NTPConfigDriver) {
	if n.ipNet != nil {
		var entries []_log.Entry
		util.IterateCIDR(n.ipNet, func(ip net.IP) {
			res := driver.setNTP(ip, n.ntp)
			if res != nil {
				entries = append(entries, *res)
			}
		})
		_log.BulkAppend(entries)
	} else {
		_log.AppendNonNil(driver.setNTP(n.ip, n.ntp))
	}
	// 追加所有驱动都不匹配的记录
	for key, value := range _log.IPNameMap {
		if value == util.UNKNOWN_DEVICE {
			e := _log.Entry{
				IP:         net.ParseIP(key),
				Code:       -2,
				Prompt:     "All Driver Not Match",
				DeviceName: util.UNKNOWN_DEVICE,
				Time:       time.Now(),
			}
			_log.AppendNonNil(&e)
		}
	}
}
