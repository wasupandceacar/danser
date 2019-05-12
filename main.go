package main

import (
	"danser/configui"
	"flag"
	"log"
	"os"
)

func main() {
	stdinLog := flag.Bool("stdinLog", false, "")
	noGUI := flag.Bool("noGUI", false, "")
	flag.Parse()

	if !*stdinLog{
		// 设置log文件
		file, _ := os.OpenFile("vsplayer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer file.Close()
		log.SetOutput(file)
		log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	}

	configui.UImain(*noGUI)
}