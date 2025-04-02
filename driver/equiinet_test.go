package driver

import (
	"net"
	"testing"
)

func TestConfigurateEquiinetNTP(t *testing.T) {
	newDriver, err := GetDriver("EQUIINET")
	if err != nil {
		t.Error(err)
	}
	expectNtp := net.ParseIP("192.168.128.198")
	driver := newDriver()
	config := NewNTPConfig(expectNtp)
	ip := "10.20.128.193"
	config.Reload(ip)
	config.ExecuteAndLog(driver)
	ntp := driver.getNTP(config.ip)
	if ntp == nil {
		t.Error("测试失败：无法设置此电话")
	} else {
		if !ntp.Equal(expectNtp) {
			t.Errorf("测试失败：预期 %s，实际：%s", expectNtp, *ntp)
		}
	}
}
