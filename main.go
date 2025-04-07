package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/typenoob/iptel-ntp-config/driver"
	_log "github.com/typenoob/iptel-ntp-config/log"
)

func main() {
	hFlag := flag.Bool("h", false, "帮助参数")
	flag.Parse()
	if len(os.Args) <= 1 || *hFlag {
		fmt.Println("用法：inc [NTP服务器地址] [IP或IP段] [IP或IP段] [IP或IP段]\n示例：inc 192.168.128.198 10.20.128.193")
		return
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
			if err := config.Reload(arg); err != nil {
				log.Println(err)
			} else {
				config.ExecuteAndLog()
			}
		}
	}
	_log.GenerateReport()
}
