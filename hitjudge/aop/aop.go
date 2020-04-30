package aop

import (
	"danser/beatmap"
	"danser/bmath"
	"danser/hitjudge"
	"danser/osuconst"
	"danser/score"
	oppai "github.com/flesnuk/oppai5"
	"math"

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
	//--------------------------------------------------------------------
	//Step 1: Init

	var objectResults []hitjudge.ObjectResult
	var totalResults []hitjudge.TotalResult
	GetCursorPositionAt := interpolateCursorPositions(rep)
	bm = parseBeatmapWithMods(bm, rep.Mods&osuconst.MOD_HR > 0, rep.Mods&osuconst.MOD_EZ > 0)
	difficultySettings := configureDifficultySettings(bm, rep)
	judgeObjects := ConvertToJudgeObjects(bm.HitObjects)
	currentTime := rep.ReplayData[0].Time + rep.ReplayData[1].Time
	l, r := 0, 0

	//--------------------------------------------------------------------
	//Step 2: Judge each judgeObject
	//TODO: Special case when you release the key while doing sliders

	for i := 2; i < len(rep.ReplayData); i++ {
		currentTime += rep.ReplayData[i].Time
		//clear all judge objects that are outdated
		for true {
			if judgeObjects[l].JudgeType == InstantJudge {
				//check judge time
				if judgeObjects[l].JudgeTime <= currentTime {
					if judgeObjects[l].JudgeResult == hitjudge.Unjudged {
						if isAnyKeyPressedAt(rep.ReplayData, judgeObjects[l].JudgeTime, i, currentTime) && judgeFollowCirclePosition(judgeObjects[l], GetCursorPositionAt(judgeObjects[l].JudgeTime), difficultySettings) {
							judgeObjects[l].JudgeResult = hitjudge.Hit300
						} else {
							judgeObjects[l].JudgeResult = hitjudge.HitMiss
						}
					}
					if l < r {
						l++
					} else {
						break
					}
				} else {
					break
				}
			} else if judgeObjects[l].JudgeType == NormalJudge {
				if float64(currentTime-judgeObjects[l].JudgeTime) > difficultySettings.HitWindowMiss {
					if judgeObjects[l].JudgeResult == hitjudge.Unjudged {
						judgeObjects[l].JudgeResult = hitjudge.HitMiss
					}
					l++
				} else {
					break
				}
			}
		}

		//add new judge objects
		for true {
			if judgeObjects[r].JudgeType == InstantJudge {
				//add all instant judges to queue
				if currentTime > judgeObjects[r].JudgeTime {
					if r < len(judgeObjects) {
						r++
					} else {
						break
					}
				}
			} else if judgeObjects[r].JudgeType == NormalJudge {
				if float64(judgeObjects[r].JudgeTime-currentTime) <= difficultySettings.HitWindowMiss {
					if r < len(judgeObjects) {
						r++
					} else {
						break
					}
				} else {
					break
				}
			}
		}
		if isAnyNewKeyPressed(rep.ReplayData[i-1].KeyPressed, rep.ReplayData[i].KeyPressed) {
			for j := l; j <= r; j++ {
				if judgeObjects[j].JudgeType == InstantJudge {
					//leave all instant judges to dequeue action to process
					continue
				} else if judgeObjects[j].JudgeType == NormalJudge {
					if judgeObjects[j].JudgeResult == hitjudge.Unjudged {
						if judgeCirclePosition(judgeObjects[j], GetCursorPositionAt(currentTime), difficultySettings) {
							judgeObjects[j].JudgeResult = judgeTiming(currentTime, judgeObjects[j].JudgeTime, difficultySettings)
							//add one break here and comment the below one to disable notelock
						}
						//it's unsafe to increment l here because there may be some instant judges before this judge
						//notelock simulation and prevent single click from hitting more than one circle
						break
					} else {
						continue
					}
				}
			}
		}
	}

	//set miss to all unjudged objects
	for i := len(judgeObjects) - 1; i >= 0; i-- {
		if judgeObjects[i].JudgeResult == hitjudge.Unjudged {
			judgeObjects[i].JudgeResult = hitjudge.HitMiss
		} else {
			break
		}
	}

	//--------------------------------------------------------------------
	//Step 3: Calculate slider results and hit results

	c300, c100, c50, cMiss, combo := 0, 0, 0, 0, 0

	//TODO: Total Results
	for i := 0; i < len(judgeObjects); {
		if judgeObjects[i].SliderIndex == -1 {
			objectResults = append(objectResults, hitjudge.ObjectResult{
				JudgePos:  judgeObjects[i].JudgePosition,
				JudgeTime: judgeObjects[i].JudgeTime,
				Result:    judgeObjects[i].JudgeResult,
				IsBreak:   judgeObjects[i].JudgeResult == hitjudge.HitMiss,
			})
			combo++
			switch judgeObjects[i].JudgeResult {
			case hitjudge.HitMiss:
				cMiss++
				break
			case hitjudge.Hit300:
				c300++
				break
			case hitjudge.Hit100:
				c100++
				break
			case hitjudge.Hit50:
				c50++
				break
			}
			i++
		} else {
			var tmp []JudgeObject
			currentSliderIndex := judgeObjects[i].SliderIndex
			for ; judgeObjects[i].SliderIndex == currentSliderIndex && i < len(judgeObjects); i++ {
				tmp = append(tmp, judgeObjects[i])
			}
			combo += len(tmp)
			hitResult := judgeSlider(tmp)
			objectResults = append(objectResults, hitjudge.ObjectResult{
				JudgePos:  tmp[0].JudgePosition,
				JudgeTime: tmp[len(tmp)-1].JudgeTime,
				Result:    hitResult,
				IsBreak:   hitResult == hitjudge.HitMiss,
			})
			switch hitResult {
			case hitjudge.HitMiss:
				cMiss++
				break
			case hitjudge.Hit300:
				c300++
				break
			case hitjudge.Hit100:
				c100++
				break
			case hitjudge.Hit50:
				c50++
				break
			}
		}
		totalResults = append(totalResults, hitjudge.TotalResult{
			N300:   uint16(c300),
			N100:   uint16(c100),
			N50:    uint16(c50),
			Misses: uint16(cMiss),
			Combo:  uint16(combo),
			Mods:   rep.Mods,
			Acc:    calculateAccuracy(c300, c100, c50, cMiss),
			Rank:   calculateRank(c300, c100, c50, cMiss, rep.Mods),
			PP:     calculatePP(settings.General.OsuSongsDir+bm.Dir+"/"+bm.File, c300, c100, c50, cMiss, combo, rep.Mods, len(objectResults)),
			UR:     calculateUR(),
		})
	}

	return objectResults, totalResults
}

func isAnyKeyPressed(kp *rplpa.KeyPressed) bool {
	return kp.Key1 || kp.Key2 || kp.LeftClick || kp.RightClick
}

//TODO: Change to use binary search
func isAnyKeyPressedAt(rd []*rplpa.ReplayData, time int64, currentIndex int, currentTime int64) bool {
	if time < currentTime {
		//start from the previous frame
		currentIndex--
		for ; currentIndex >= 2; currentIndex-- {
			currentTime -= rd[currentIndex].Time
			if currentTime <= time {
				return isAnyKeyPressed(rd[currentIndex].KeyPressed)
			}
		}
		return false
	} else if time == currentTime {
		return isAnyKeyPressed(rd[currentIndex].KeyPressed)
	} else {
		//start from the next frame
		currentIndex++
		for ; currentIndex < len(rd); currentIndex++ {
			currentTime += rd[currentIndex].Time
			if currentTime == time {
				return isAnyKeyPressed(rd[currentIndex].KeyPressed)
			} else if currentTime > time && currentIndex-1 >= 2 {
				//use the previous frame
				return isAnyKeyPressed(rd[currentIndex-1].KeyPressed)
			}
		}
		return false
	}
}

func isAnyNewKeyPressed(lastFrame *rplpa.KeyPressed, thisFrame *rplpa.KeyPressed) bool {
	return lastFrame.Key1 != thisFrame.Key1 || lastFrame.Key2 != thisFrame.Key2 || lastFrame.LeftClick != thisFrame.LeftClick || lastFrame.RightClick != thisFrame.RightClick
}

func judgeTiming(currentTime int64, judgeTime int64, difficultySettings DifficultySettings) hitjudge.HitResult {
	delta := math.Abs(float64(currentTime - judgeTime))
	if delta < difficultySettings.HitWindow300 {
		return hitjudge.Hit300
	} else if delta < difficultySettings.HitWindow100 {
		return hitjudge.Hit100
	} else if delta < difficultySettings.HitWindow50 {
		return hitjudge.Hit50
	} else {
		return hitjudge.HitMiss
	}
}

func judgeCirclePosition(object JudgeObject, cursorPosition bmath.Vector2d, difficultySettings DifficultySettings) bool {
	if object.JudgePosition.Dst(cursorPosition) <= difficultySettings.RealCircleSize {
		return true
	} else {
		return false
	}
}

func judgeFollowCirclePosition(object JudgeObject, cursorPosition bmath.Vector2d, difficultySettings DifficultySettings) bool {
	if object.JudgePosition.Dst(cursorPosition) <= difficultySettings.FollowCircleSize {
		return true
	} else {
		return false
	}
}

func judgeSlider(judgeObjects []JudgeObject) hitjudge.HitResult {
	tickMissedCount := 0
	for i := 0; i < len(judgeObjects); i++ {
		if judgeObjects[i].JudgeResult == hitjudge.HitMiss {
			tickMissedCount++
		}
	}
	percentage := float64(tickMissedCount) / float64(len(judgeObjects))
	if percentage == 1 {
		return hitjudge.HitMiss
	} else if percentage >= 0.5 {
		return hitjudge.Hit50
	} else if percentage > 0 {
		return hitjudge.Hit100
	} else {
		return hitjudge.Hit300
	}
}

func calculateAccuracy(c300, c100, c50, cMiss int) float64 {
	return float64(c300*300+c100*100+c50*50) / float64(c300+c100+c50+cMiss) * 300.0
}

func calculateRank(c300, c100, c50, cMiss int, mods uint32) score.Rank {
	countAll := c300 + c100 + c50 + cMiss
	if c300 == countAll {
		if isSilver(mods) {
			return score.SSH
		} else {
			return score.SS
		}
	} else if ((float64(c300) / float64(countAll)) > 0.9) && ((float64(c50) / float64(countAll)) < 0.01) && (cMiss == 0) {
		if isSilver(mods) {
			return score.SH
		} else {
			return score.S
		}
	} else if ((float64(c300) / float64(countAll)) > 0.9) || (((float64(c300) / float64(countAll)) > 0.8) && (cMiss == 0)) {
		return score.A
	} else if ((float64(c300) / float64(countAll)) > 0.8) || (((float64(c300) / float64(countAll)) > 0.7) && (cMiss == 0)) {
		return score.B
	} else if (float64(c300) / float64(countAll)) > 0.6 {
		return score.C
	} else {
		return score.D
	}
}

func isSilver(mods uint32) bool {
	return (mods&osuconst.MOD_HD > 0) || (mods&osuconst.MOD_FL > 0)
}

func calculatePP(filename string, c300, c100, c50, cMiss, combo int, mods uint32, objNum int) oppai.PPv2 {
	return oppai.PPInfo(score.LoadMapByNum(filename, objNum), &oppai.Parameters{
		Combo:  uint16(combo),
		Mods:   mods,
		N300:   uint16(c300),
		N100:   uint16(c100),
		N50:    uint16(c50),
		Misses: uint16(cMiss),
	}).PP
}

//TODO
func calculateUR() float64 {
	return 0.0
}
