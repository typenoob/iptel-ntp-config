package main

import (
	"errors"
	"log"
	"net"
	"os"

	"github.com/typenoob/iptel-ntp-config/driver"
	_log "github.com/typenoob/iptel-ntp-config/log"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalln(errors.New("错误：输入不合法！\n用法：inc [NTP服务器地址] [IP或IP段] [IP或IP段] [IP或IP段]\n示例：inc 192.168.128.198 10.20.128.193"))
	}
	ntp := os.Args[1]
	config := driver.NewNTPConfig(net.ParseIP(ntp))
	for _, arg := range os.Args[2:] {
		for _, name := range driver.GetDriverList() {
			newDriver, err := driver.GetDriver(name)
			if err != nil {
				log.Println(err)
			}
			config.SetDriver(newDriver())
			config.Reload(arg)
			config.ExecuteAndLog()
		}
	}
	_log.GenerateReport()
}
