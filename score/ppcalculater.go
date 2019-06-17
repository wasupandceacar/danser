package score

import (
	"danser/settings"
	"github.com/flesnuk/oppai5"
	"math"
	"os"
)

// 部分载入map
func LoadMapbyNum(filename string, objnum int) *oppai.Map {
	f, _ := os.Open(filename)
	return oppai.ParsebyNum(f, objnum)
}

// 部分载入map，并计算难度信息
func CalculateDiffbyNum(filename string, objnum int, mods uint32) oppai.PP {
	beatmap := LoadMapbyNum(filename, objnum)
	return oppai.PPInfo(beatmap, &oppai.Parameters{
		Combo:  uint16(beatmap.MaxCombo),
		Mods:   mods,
		N300:   uint16(objnum),
		N100:   0,
		N50:    0,
		Misses: 0,
	})
}

// 计算每帧实时数值（PP、UR）
func CalculateRealtimeValue(firstvalue float64, secondvalue float64, firsttime int64, secondtime int64, nowtime float64) (realvalue float64) {
	deltavalue := secondvalue - firstvalue
	deltatime := math.Min(float64(secondtime - firsttime), settings.VSplayer.PlayerInfoUI.RealTimePPGap)
	realvalue = firstvalue + deltavalue * math.Max(math.Min(math.Min(nowtime - float64(firsttime + settings.VSplayer.PlayerFieldUI.HitFadeTime), settings.VSplayer.PlayerInfoUI.RealTimePPGap) / deltatime, 1), 0)
	if math.IsNaN(realvalue) {
		realvalue = 0.0
	}
	return realvalue
}