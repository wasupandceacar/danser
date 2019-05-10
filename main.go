package main

import (
	"danser/configui"
	"log"
	"os"
)

var IDEDEBUG = false

func main() {
	if !IDEDEBUG {
		// 设置log文件
		file, _ := os.OpenFile("danser.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer file.Close()
		log.SetOutput(file)
		log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	}

	configui.UImain()
}