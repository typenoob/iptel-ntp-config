package driver

import (
	"fmt"
	"net"

	_log "github.com/typenoob/iptel-ntp-config/log"
)

type DriverConstructor func() NTPConfigDriver

var driverMap = map[string]DriverConstructor{}

func RegisterDriver(driver DriverConstructor) {
	tempDriver := driver()
	driverMap[tempDriver.GetName()] = driver
}

func GetDriver(name string) (DriverConstructor, error) {
	n, ok := driverMap[name]
	if !ok {
		return nil, fmt.Errorf("没有名为%s的驱动", name)
	}
	return n, nil
}

func GetDriverList() []string {
	keys := make([]string, 0, len(driverMap))
	for key := range driverMap {
		keys = append(keys, key)
	}
	return keys
}

func NewNTPConfig(ntp net.IP) NTPConfig {
	_log.NewIPNameMap(ntp)
	return NTPConfig{ntp: ntp}
}
