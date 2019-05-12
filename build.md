## 完全 Debug 版本（输出控制台、无GUI） ##

	go build && danser -stdinLog -noGUI

## 完全 Release 版本（输出日志文件、有GUI） ##

	go build -ldflags="-H windowsgui" && danser

