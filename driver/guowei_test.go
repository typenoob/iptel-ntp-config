package driver

import (
	"net"
	"testing"
)

func TestConfigurateGuoweiNTP(t *testing.T) {
	newDriver, err := GetDriver("GUOWEI")
	if err != nil {
		t.Error(err)
	}
	expectNtp := net.ParseIP("10.20.128.46")
	config := NewNTPConfig(expectNtp)
	driver := newDriver()
	config.SetDriver(driver)
	ip := "10.20.132.78"
	config.Reload(ip)
	config.ExecuteAndLog()
	ntp := driver.getNTP(config.ip)
	if ntp == nil {
		t.Error("测试失败：无法设置此电话")
	} else {
		if !ntp.Equal(expectNtp) {
			t.Errorf("测试失败：预期 %s，实际：%s", expectNtp, *ntp)
		}
	}
}
