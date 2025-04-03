## IP电话批量设置NTP服务器

### 支持设备
- 安科耐特(Equiinet)
- 国威(hb-voice)

### 构建命令

`go build -o inc.exe main.go`

### 用法

`inc [NTP服务器地址] [IP或IP段] [IP或IP段] [IP或IP段]`

### 功能
- 📝日志系统
- 🌐兼容设置IP地址与IP段
- ☎️自动过滤设备
- ⏩智能跳过已设置成功的设备
- 🔀自适应驱动协商
- 🧩可扩展性

### 已知问题

1. 短时间内频繁设置国威电话可能会出现Server Busy的错误