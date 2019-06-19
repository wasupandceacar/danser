package prof

import (
	"os"
	"runtime/pprof"
)

func ProfStart() {
	f, err := os.OpenFile("vsplayer.prof", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
}

func ProfEnd(){
	pprof.StopCPUProfile()
}
