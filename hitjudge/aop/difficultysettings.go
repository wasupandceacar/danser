package aop

import (
	"danser/osuconst"
	"github.com/Mempler/rplpa"
	"math"
)

type DifficultySettings struct {
	OverallDifficulty, CircleSize                           float64
	HitWindow300, HitWindow100, HitWindow50, RealCircleSize float64
}

func NewDifficultySettings(od, cs float64) DifficultySettings {
	ds := DifficultySettings{
		OverallDifficulty: od,
		CircleSize:        cs,
	}
	ds.calculateSettings()
	return ds
}

func (ds *DifficultySettings) AdjustForReplay(rep *rplpa.Replay) {
	if rep.Mods&osuconst.MOD_HR > 0 {
		ds.adjustForHardRock()
	} else if rep.Mods&osuconst.MOD_EZ > 0 {
		ds.adjustForEasy()
	}
}

func (ds *DifficultySettings) adjustForHardRock() {
	ds.OverallDifficulty = math.Min(ds.OverallDifficulty*osuconst.OD_HR_HENSE, osuconst.OD_MAX)
	ds.CircleSize = math.Min(ds.CircleSize*osuconst.CS_HR_HENSE, osuconst.CS_MAX)
	ds.calculateSettings()
}

func (ds *DifficultySettings) adjustForEasy() {
	ds.OverallDifficulty *= osuconst.OD_EZ_HENSE
	ds.CircleSize *= osuconst.CS_EZ_HENSE
	ds.calculateSettings()
}

func (ds *DifficultySettings) calculateSettings() {
	ds.HitWindow300 = osuconst.HITWINDOW_300_BASE - (ds.OverallDifficulty * osuconst.HITWINDOW_300_MULT) + osuconst.HITWINDOW_OFFSET
	ds.HitWindow100 = osuconst.HITWINDOW_100_BASE - (ds.OverallDifficulty * osuconst.HITWINDOW_100_MULT) + osuconst.HITWINDOW_OFFSET
	ds.HitWindow50 = osuconst.HITWINDOW_50_BASE - (ds.OverallDifficulty * osuconst.HITWINDOW_50_MULT) + osuconst.HITWINDOW_OFFSET
	ds.RealCircleSize = 32 * (1 - 0.7*(ds.CircleSize-5)/5)
}
