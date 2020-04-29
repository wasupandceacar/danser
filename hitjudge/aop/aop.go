package aop

import (
	"danser/beatmap"
	"danser/bmath"
	"danser/hitjudge"
	//"danser/osuconst"
	"danser/settings"
	"github.com/Mempler/rplpa"
)

func parseBeatmapWithMods(bm *beatmap.BeatMap, isHR bool, isEZ bool) *beatmap.BeatMap {
	b := *bm
	beatmap.ParseObjectsByPath(&b, settings.General.OsuSongsDir+b.Dir+"/"+b.File, isHR, isEZ)
	return &b
}

func configureDifficultySettings(bm *beatmap.BeatMap, rep *rplpa.Replay) DifficultySettings {
	ds := NewDifficultySettings(bm.OverallDifficulty, bm.CircleSize)
	ds.AdjustForReplay(rep)
	return ds
}

func interpolateCursorPositions(rep *rplpa.Replay) func(t int64) bmath.Vector2d {
	var ts []float64
	var xs []float64
	var ys []float64
	if len(rep.ReplayData) > 0 {
		ts = append(ts, float64(rep.ReplayData[0].Time))
		xs = append(xs, float64(rep.ReplayData[0].MosueX))
		ys = append(ys, float64(rep.ReplayData[0].MouseY))
	}
	for i := 1; i < len(rep.ReplayData); i++ {
		ts = append(ts, float64(rep.ReplayData[i].Time))
		xs = append(xs, float64(rep.ReplayData[i].MosueX))
		ys = append(ys, float64(rep.ReplayData[i].MouseY))
		ts[i] += ts[i-1]
	}
	fx := CubicSpline(ts, xs)
	fy := CubicSpline(ts, ys)
	return func(t int64) bmath.Vector2d {
		return bmath.Vector2d{
			X: fx(float64(t)),
			Y: fy(float64(t)),
		}
	}
}

func Judge(bm *beatmap.BeatMap, rep *rplpa.Replay) ([]hitjudge.ObjectResult, []hitjudge.TotalResult) {
	var objectResults []hitjudge.ObjectResult
	var totalResults []hitjudge.TotalResult
	//bm = ParseBeatmapWithMods(bm, rep.Mods&osuconst.MOD_HR > 0, rep.Mods&osuconst.MOD_EZ > 0)
	//difficultySettings := configureDifficultySettings(bm, rep)
	//strainTime := bm.HitObjects[len(bm.HitObjects) - 1].GetBasicData().EndTime
	//var currentTime uint64 = 0
	//for true {
	//
	//}
	return objectResults, totalResults
}
