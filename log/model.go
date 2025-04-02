package log

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/typenoob/iptel-ntp-config/util"
)

func padString(s string, width int) string {
	return runewidth.FillRight(s, width)
}

type Entry struct {
	IP         net.IP
	Code       int    `json:"return_code"`
	Prompt     string `json:"return_prompt"`
	DeviceName string
	Time       time.Time
	Ntp        net.IP
}

type EntryList []Entry

func (e EntryList) Len() int {
	return len(e)
}

func (e EntryList) Less(i, j int) bool {
	if e[i].IP.To4() == nil && e[j].IP.To4() == nil {
		return false
	}
	if e[i].IP.Equal(e[j].IP) {
		return e[j].Time.After(e[i].Time)
	}
	return bytes.Compare(e[i].IP, e[j].IP) < 0
}

func (e EntryList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

var entries EntryList

var IPNameMap map[string]string

func BulkAppend(e []Entry) {
	entries = append(entries, e...)
}

func AppendNonNil(e *Entry) {
	if e != nil {
		entries = append(entries, *e)
	}
}

func NewIPNameMap(ntp net.IP) {
	IPNameMap = make(map[string]string)
	for i, entry := range entries {
		if i+1 == len(entries) || !entry.IP.Equal(entries[i+1].IP) {
			if entry.Code == 0 && entry.Ntp.Equal(ntp) {
				SetNameByIP(entry.IP, util.HISTORY_DEVICE)
			}
		}
	}
}

func GetNameByIP(ip net.IP) string {
	return IPNameMap[ip.String()]
}

func SetNameByIP(ip net.IP, name string) {
	key := ip.String()
	if _, exists := IPNameMap[key]; !exists {
		IPNameMap[key] = name
	} else if IPNameMap[key] == util.UNKNOWN_DEVICE || IPNameMap[key] == util.HISTORY_DEVICE {
		IPNameMap[key] = name
	}
}

func GenerateReport() {
	fmt.Printf("%s%s%s%s%s%s\n",
		padString("IP 地址", 15), padString("结果", 10), padString("消息", 25),
		padString("类型", 12), padString("时间", 20), padString("NTP地址", 15),
	)
	fmt.Println(strings.Repeat("-", 98))
	sort.Sort(entries)
	for i, entry := range entries {
		status := "❌ 失败"
		if entry.Code == 0 {
			status = "✅ 成功"
		}
		// 只输出最晚一次的结果
		if i+1 == len(entries) || !entry.IP.Equal(entries[i+1].IP) {
			if GetNameByIP(entry.IP) != "" && GetNameByIP(entry.IP) != util.HISTORY_DEVICE {
				fmt.Printf("%s%s%s%s%s%s\n",
					padString(entry.IP.String(), 15), padString(status, 10), padString(entry.Prompt, 25),
					padString(entry.DeviceName, 12), padString(entry.Time.Format(time.DateTime), 20), padString(entry.Ntp.String(), 15),
				)
			}
		}
	}
	fmt.Println("完整结果已保存至result.json")
	saveToJSON(entries)
}

func init() {
	entries = readFromJSON()
}
