package driver

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	_log "github.com/typenoob/iptel-ntp-config/log"
	"github.com/typenoob/iptel-ntp-config/util"
)

type NTPConfigDriver interface {
	setNTP(ip net.IP, ntp net.IP, ntp2 net.IP) *_log.Entry
	getNTP(ip net.IP) (*net.IP, *net.IP)
	IsMatch(ip net.IP) bool
	GetName() string
}

type IPOrNet interface {
	isIPOrNet()
}

type NTPConfig struct {
	ip     net.IP
	ipNet  *net.IPNet
	ntp    net.IP
	ntp2   net.IP
	driver NTPConfigDriver
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

type X func(ip net.IP, ntp net.IP) *_log.Entry

func (n NTPConfig) GetWebHandler(ip net.IP) func() *_log.Entry {
	if _log.GetNameByIP(ip) == util.HISTORY_DEVICE {
		// 若历史记录中已成功执行，则直接跳过
		_log.SetNameByIP(ip, n.driver.GetName())
		return func() *_log.Entry {
			return nil
		}
	}
	if !_log.GetValid(ip) {
		// 若其他驱动已判断过该地址无效，则直接跳过
		return func() *_log.Entry {
			return nil
		}
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), "80"), 2*time.Second)
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		log.Println(err)
		// 若此IP未启用web服务，则跳过
		_log.SetInvalid(ip)
		return func() *_log.Entry {
			return nil
		}
	}
	//----------自动跳过电子时钟-----------
	resp, err := http.Get(fmt.Sprintf("http://%s/web/heading_iot_clock.jpg", ip))
	if err == nil && resp.StatusCode == http.StatusOK {
		_log.SetInvalid(ip)
		return func() *_log.Entry {
			return nil
		}
	}
	////----------------------------------
	_log.SetValid(ip)
	return func() *_log.Entry {
		return n.driver.setNTP(ip, n.ntp, n.ntp2)
	}
}

func (n *NTPConfig) Reload(s string) error {
	ip, ipNet, err := net.ParseCIDR(s)
	if err == nil {
		n.ip = ip
		n.ipNet = ipNet
		return nil
	}
	ip = net.ParseIP(s)
	if ip != nil {
		n.ip = ip
		return nil
	}
	return errors.New("错误：输入的参数既不是IP地址也不是IP段！")
}

func (n *NTPConfig) SetDriver(driver NTPConfigDriver) {
	n.driver = driver
}

func (n NTPConfig) ExecuteAndLog() {
	if n.ipNet != nil {
		var entries []_log.Entry
		util.IterateCIDR(n.ipNet, func(ip net.IP) {
			res := n.GetWebHandler(ip)()
			if res != nil {
				entries = append(entries, *res)
			}
		})
		_log.BulkAppend(entries)
	} else {
		log.Println(n.GetWebHandler(n.ip))
		_log.AppendNonNil(n.GetWebHandler(n.ip)())
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
